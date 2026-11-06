//! Managing `users` within a company, like the Go `userRepository`.

use chrono::{DateTime, Utc};
use sqlx::{PgExecutor, PgPool};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct UserRow {
    pub id: i64,
    pub company_id: i64,
    pub role_id: Option<i64>,
    pub email: String,
    pub name: String,
    pub status: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

const COLUMNS: &str = "id, company_id, role_id, email, name, status, created_at, updated_at";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<UserRow>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM users WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn list(
    db: &PgPool,
    company_id: i64,
    limit: i64,
    offset: i64,
) -> sqlx::Result<(Vec<UserRow>, i64)> {
    let total: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM users WHERE company_id = $1")
        .bind(company_id)
        .fetch_one(db)
        .await?;
    let users = sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM users WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
    ))
    .bind(company_id)
    .bind(limit)
    .bind(offset)
    .fetch_all(db)
    .await?;
    Ok((users, total))
}

pub async fn email_taken(
    db: impl PgExecutor<'_>,
    email: &str,
    exclude_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)")
        .bind(email)
        .bind(exclude_id)
        .fetch_one(db)
        .await
}

/// A new active user with no role.
pub async fn create(
    db: impl PgExecutor<'_>,
    company_id: i64,
    email: &str,
    password_hash: &str,
    name: &str,
) -> sqlx::Result<UserRow> {
    sqlx::query_as(&format!(
        "INSERT INTO users (company_id, email, password_hash, name, status, created_at, updated_at) \
         VALUES ($1, $2, $3, $4, 'active', NOW(), NOW()) RETURNING {COLUMNS}"
    ))
    .bind(company_id)
    .bind(email)
    .bind(password_hash)
    .bind(name)
    .fetch_one(db)
    .await
}

/// `true` if a row was updated.
pub async fn update(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    email: &str,
    name: &str,
) -> sqlx::Result<bool> {
    let res = sqlx::query("UPDATE users SET email = $1, name = $2, updated_at = NOW() WHERE id = $3 AND company_id = $4")
        .bind(email)
        .bind(name)
        .bind(id)
        .bind(company_id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}

/// `true` if a row was updated.
pub async fn update_status(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    status: &str,
) -> sqlx::Result<bool> {
    let res = sqlx::query(
        "UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2 AND company_id = $3",
    )
    .bind(status)
    .bind(id)
    .bind(company_id)
    .execute(db)
    .await?;
    Ok(res.rows_affected() > 0)
}

/// `true` if a row was deleted.
pub async fn delete(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    let res = sqlx::query("DELETE FROM users WHERE id = $1 AND company_id = $2")
        .bind(id)
        .bind(company_id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}
