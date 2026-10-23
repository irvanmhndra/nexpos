//! `customers`, like the Go `CustomerRepository`.

use chrono::{DateTime, Utc};
use sqlx::{PgExecutor, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Customer {
    pub id: i64,
    pub code: String,
    pub name: String,
    pub phone: Option<String>,
    pub email: Option<String>,
    pub is_member: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub struct CustomerInput<'a> {
    pub code: &'a str,
    pub name: &'a str,
    pub phone: Option<&'a str>,
    pub email: Option<&'a str>,
    pub is_member: bool,
}

pub struct ListFilter<'a> {
    pub search: &'a str,
    pub is_member: Option<bool>,
    pub limit: i64,
    pub offset: i64,
}

const COLUMNS: &str = "id, code, name, phone, email, is_member, created_at, updated_at";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<Customer>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM customers WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// Same filters, order (newest first), and count as the Go repository.
pub async fn list(
    db: &sqlx::PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<Customer>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (name ILIKE ")
                .push_bind(like.clone())
                .push(" OR code ILIKE ")
                .push_bind(like.clone())
                .push(" OR phone ILIKE ")
                .push_bind(like.clone())
                .push(" OR email ILIKE ")
                .push_bind(like)
                .push(")");
        }
        if let Some(member) = f.is_member {
            q.push(" AND is_member = ").push_bind(member);
        }
    };

    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM customers");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM customers"));
    filter(&mut rows);
    rows.push(" ORDER BY created_at DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    let customers = rows.build_query_as().fetch_all(db).await?;
    Ok((customers, total))
}

pub async fn code_exists(
    db: impl PgExecutor<'_>,
    company_id: i64,
    code: &str,
    exclude_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM customers WHERE company_id = $1 AND code = $2 AND id != $3)",
    )
    .bind(company_id)
    .bind(code)
    .bind(exclude_id)
    .fetch_one(db)
    .await
}

pub async fn create(
    db: impl PgExecutor<'_>,
    company_id: i64,
    c: &CustomerInput<'_>,
) -> sqlx::Result<Customer> {
    sqlx::query_as(&format!(
        "INSERT INTO customers (company_id, code, name, phone, email, is_member) \
         VALUES ($1, $2, $3, $4, $5, $6) RETURNING {COLUMNS}"
    ))
    .bind(company_id)
    .bind(c.code)
    .bind(c.name)
    .bind(c.phone)
    .bind(c.email)
    .bind(c.is_member)
    .fetch_one(db)
    .await
}

pub async fn update(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    c: &CustomerInput<'_>,
) -> sqlx::Result<Option<Customer>> {
    sqlx::query_as(&format!(
        "UPDATE customers SET code = $3, name = $4, phone = $5, email = $6, is_member = $7, updated_at = NOW() \
         WHERE company_id = $1 AND id = $2 RETURNING {COLUMNS}"
    ))
    .bind(company_id)
    .bind(id)
    .bind(c.code)
    .bind(c.name)
    .bind(c.phone)
    .bind(c.email)
    .bind(c.is_member)
    .fetch_optional(db)
    .await
}

/// `true` if a row was deleted.
pub async fn delete(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    let res = sqlx::query("DELETE FROM customers WHERE company_id = $1 AND id = $2")
        .bind(company_id)
        .bind(id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}
