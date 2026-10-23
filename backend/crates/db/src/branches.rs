//! `branches`, like the Go `branchRepository` (always scoped to a company here).

use chrono::{DateTime, Utc};
use sqlx::{PgExecutor, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Branch {
    pub id: i64,
    pub code: String,
    pub name: String,
    pub address: Option<String>,
    pub phone: Option<String>,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub struct BranchInput<'a> {
    pub code: &'a str,
    pub name: &'a str,
    pub address: Option<&'a str>,
    pub phone: Option<&'a str>,
    pub is_active: bool,
}

pub struct ListFilter<'a> {
    pub search: &'a str,
    pub is_active: Option<bool>,
    pub limit: i64,
    pub offset: i64,
}

const COLUMNS: &str = "id, code, name, address, phone, is_active, created_at, updated_at";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<Branch>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM branches WHERE company_id = $1 AND id = $2"
    ))
    .bind(company_id)
    .bind(id)
    .fetch_optional(db)
    .await
}

/// Same filters, order (by name), and count as the Go repository.
pub async fn list(
    db: &sqlx::PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<Branch>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (name ILIKE ")
                .push_bind(like.clone())
                .push(" OR code ILIKE ")
                .push_bind(like.clone())
                .push(" OR address ILIKE ")
                .push_bind(like.clone())
                .push(" OR phone ILIKE ")
                .push_bind(like)
                .push(")");
        }
        if let Some(active) = f.is_active {
            q.push(" AND is_active = ").push_bind(active);
        }
    };

    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM branches");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM branches"));
    filter(&mut rows);
    rows.push(" ORDER BY name LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    let branches = rows.build_query_as().fetch_all(db).await?;
    Ok((branches, total))
}

pub async fn code_exists(
    db: impl PgExecutor<'_>,
    company_id: i64,
    code: &str,
    exclude_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM branches WHERE company_id = $1 AND code = $2 AND id != $3)",
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
    b: &BranchInput<'_>,
) -> sqlx::Result<Branch> {
    sqlx::query_as(&format!(
        "INSERT INTO branches (company_id, code, name, address, phone, is_active, created_at, updated_at) \
         VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING {COLUMNS}"
    ))
    .bind(company_id)
    .bind(b.code)
    .bind(b.name)
    .bind(b.address)
    .bind(b.phone)
    .bind(b.is_active)
    .fetch_one(db)
    .await
}

pub async fn update(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    b: &BranchInput<'_>,
) -> sqlx::Result<Option<Branch>> {
    sqlx::query_as(&format!(
        "UPDATE branches SET code = $3, name = $4, address = $5, phone = $6, is_active = $7, updated_at = NOW() \
         WHERE company_id = $1 AND id = $2 RETURNING {COLUMNS}"
    ))
    .bind(company_id)
    .bind(id)
    .bind(b.code)
    .bind(b.name)
    .bind(b.address)
    .bind(b.phone)
    .bind(b.is_active)
    .fetch_optional(db)
    .await
}

/// `true` if a row was deleted.
pub async fn delete(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    let res = sqlx::query("DELETE FROM branches WHERE company_id = $1 AND id = $2")
        .bind(company_id)
        .bind(id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}
