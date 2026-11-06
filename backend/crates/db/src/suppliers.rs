//! `suppliers`, like the Go `SupplierRepository`.

use chrono::{DateTime, Utc};
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Supplier {
    pub id: i64,
    pub name: String,
    pub code: String,
    pub contact_name: Option<String>,
    pub phone: Option<String>,
    pub email: Option<String>,
    pub address: Option<String>,
    pub notes: Option<String>,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub struct SupplierInput<'a> {
    pub name: &'a str,
    pub code: &'a str,
    pub contact_name: Option<&'a str>,
    pub phone: Option<&'a str>,
    pub email: Option<&'a str>,
    pub address: Option<&'a str>,
    pub notes: Option<&'a str>,
    pub is_active: bool,
}

pub struct ListFilter<'a> {
    pub search: &'a str,
    pub is_active: Option<bool>,
    pub limit: i64,
    pub offset: i64,
}

const COLUMNS: &str =
    "id, name, code, contact_name, phone, email, address, notes, is_active, created_at, updated_at";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<Supplier>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM suppliers WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn list(
    db: &PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<Supplier>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (name ILIKE ")
                .push_bind(like.clone())
                .push(" OR code ILIKE ")
                .push_bind(like.clone())
                .push(" OR contact_name ILIKE ")
                .push_bind(like)
                .push(")");
        }
        if let Some(active) = f.is_active {
            q.push(" AND is_active = ").push_bind(active);
        }
    };

    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM suppliers");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM suppliers"));
    filter(&mut rows);
    rows.push(" ORDER BY name ASC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    let suppliers = rows.build_query_as().fetch_all(db).await?;
    Ok((suppliers, total))
}

/// Whether another supplier of the company has this code.
pub async fn code_taken(
    db: impl PgExecutor<'_>,
    company_id: i64,
    code: &str,
    exclude_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM suppliers WHERE company_id = $1 AND code = $2 AND id != $3)",
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
    s: &SupplierInput<'_>,
) -> sqlx::Result<Supplier> {
    sqlx::query_as(&format!(
        "INSERT INTO suppliers (company_id, name, code, contact_name, phone, email, address, notes, is_active) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING {COLUMNS}"
    ))
    .bind(company_id)
    .bind(s.name)
    .bind(s.code)
    .bind(s.contact_name)
    .bind(s.phone)
    .bind(s.email)
    .bind(s.address)
    .bind(s.notes)
    .bind(s.is_active)
    .fetch_one(db)
    .await
}

pub async fn update(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    s: &SupplierInput<'_>,
) -> sqlx::Result<Option<Supplier>> {
    sqlx::query_as(&format!(
        "UPDATE suppliers SET name = $1, code = $2, contact_name = $3, phone = $4, email = $5, address = $6, \
             notes = $7, is_active = $8, updated_at = NOW() \
         WHERE id = $9 AND company_id = $10 RETURNING {COLUMNS}"
    ))
    .bind(s.name)
    .bind(s.code)
    .bind(s.contact_name)
    .bind(s.phone)
    .bind(s.email)
    .bind(s.address)
    .bind(s.notes)
    .bind(s.is_active)
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// `true` if a row was deleted.
pub async fn delete(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    let res = sqlx::query("DELETE FROM suppliers WHERE id = $1 AND company_id = $2")
        .bind(id)
        .bind(company_id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}
