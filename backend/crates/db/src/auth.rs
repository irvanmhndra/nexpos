//! Users, companies, branches, roles, and sessions as the auth flow needs them.
//! Mirrors the Go repositories, so rows written here look the same to Go.

use chrono::{DateTime, Utc};
use sqlx::{PgConnection, PgExecutor};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct User {
    pub id: i64,
    pub company_id: i64,
    pub role_id: Option<i64>,
    pub email: String,
    pub password_hash: String,
    pub name: String,
    pub status: String,
}

impl User {
    pub fn is_active(&self) -> bool {
        self.status == "active"
    }
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Named {
    pub id: i64,
    pub code: String,
    pub name: String,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Session {
    pub id: i64,
    pub user_id: i64,
    pub company_id: i64,
    pub branch_id: Option<i64>,
    pub access_token_expires_at: DateTime<Utc>,
    pub refresh_token_expires_at: DateTime<Utc>,
    pub device_info: Option<String>,
    pub ip_address: Option<String>,
    pub user_agent: Option<String>,
}

pub struct NewSession<'a> {
    pub user_id: i64,
    pub company_id: i64,
    pub branch_id: Option<i64>,
    pub access_token_hash: &'a str,
    pub access_token_expires_at: DateTime<Utc>,
    pub refresh_token_hash: &'a str,
    pub refresh_token_expires_at: DateTime<Utc>,
    pub device_info: Option<&'a str>,
    pub ip_address: Option<&'a str>,
    pub user_agent: Option<&'a str>,
}

const SESSION_COLUMNS: &str = "id, user_id, company_id, branch_id, access_token_expires_at, \
     refresh_token_expires_at, device_info, ip_address, user_agent";

pub async fn user_by_email(db: impl PgExecutor<'_>, email: &str) -> sqlx::Result<Option<User>> {
    sqlx::query_as(
        "SELECT id, company_id, role_id, email, password_hash, name, status FROM users WHERE email = $1",
    )
    .bind(email)
    .fetch_optional(db)
    .await
}

pub async fn role_code(db: impl PgExecutor<'_>, role_id: i64) -> sqlx::Result<Option<String>> {
    sqlx::query_scalar("SELECT code FROM roles WHERE id = $1")
        .bind(role_id)
        .fetch_optional(db)
        .await
}

/// The system (company-less) role with this code, e.g. `owner`.
pub async fn system_role(
    db: impl PgExecutor<'_>,
    code: &str,
) -> sqlx::Result<Option<(i64, String)>> {
    sqlx::query_as("SELECT id, code FROM roles WHERE code = $1 AND company_id IS NULL")
        .bind(code)
        .fetch_optional(db)
        .await
}

pub async fn company(db: impl PgExecutor<'_>, id: i64) -> sqlx::Result<Named> {
    sqlx::query_as("SELECT id, code, name FROM companies WHERE id = $1")
        .bind(id)
        .fetch_one(db)
        .await
}

pub async fn default_branch(db: impl PgExecutor<'_>, user_id: i64) -> sqlx::Result<Option<Named>> {
    sqlx::query_as(
        "SELECT b.id, b.code, b.name FROM branches b \
         JOIN user_branches ub ON ub.branch_id = b.id \
         WHERE ub.user_id = $1 AND ub.is_default = true LIMIT 1",
    )
    .bind(user_id)
    .fetch_optional(db)
    .await
}

pub async fn create_company(
    db: impl PgExecutor<'_>,
    code: &str,
    name: &str,
) -> sqlx::Result<Named> {
    sqlx::query_as(
        "INSERT INTO companies (code, name, is_active, created_at, updated_at) \
         VALUES ($1, $2, true, NOW(), NOW()) RETURNING id, code, name",
    )
    .bind(code)
    .bind(name)
    .fetch_one(db)
    .await
}

pub async fn create_branch(
    db: impl PgExecutor<'_>,
    company_id: i64,
    code: &str,
    name: &str,
) -> sqlx::Result<Named> {
    sqlx::query_as(
        "INSERT INTO branches (company_id, code, name, is_active, created_at, updated_at) \
         VALUES ($1, $2, $3, true, NOW(), NOW()) RETURNING id, code, name",
    )
    .bind(company_id)
    .bind(code)
    .bind(name)
    .fetch_one(db)
    .await
}

pub async fn create_user(
    db: impl PgExecutor<'_>,
    company_id: i64,
    role_id: i64,
    email: &str,
    password_hash: &str,
    name: &str,
) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "INSERT INTO users (company_id, role_id, email, password_hash, name, status, created_at, updated_at) \
         VALUES ($1, $2, $3, $4, $5, 'active', NOW(), NOW()) RETURNING id",
    )
    .bind(company_id)
    .bind(role_id)
    .bind(email)
    .bind(password_hash)
    .bind(name)
    .fetch_one(db)
    .await
}

pub async fn assign_branch(
    db: impl PgExecutor<'_>,
    user_id: i64,
    branch_id: i64,
    is_default: bool,
) -> sqlx::Result<()> {
    sqlx::query("INSERT INTO user_branches (user_id, branch_id, is_default) VALUES ($1, $2, $3)")
        .bind(user_id)
        .bind(branch_id)
        .bind(is_default)
        .execute(db)
        .await?;
    Ok(())
}

pub async fn insert_session(db: impl PgExecutor<'_>, s: &NewSession<'_>) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "INSERT INTO user_sessions (user_id, company_id, branch_id, \
             access_token_hash, access_token_expires_at, refresh_token_hash, refresh_token_expires_at, \
             is_revoked, device_info, ip_address, user_agent, created_at) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, false, $8, $9, $10, NOW()) RETURNING id",
    )
    .bind(s.user_id)
    .bind(s.company_id)
    .bind(s.branch_id)
    .bind(s.access_token_hash)
    .bind(s.access_token_expires_at)
    .bind(s.refresh_token_hash)
    .bind(s.refresh_token_expires_at)
    .bind(s.device_info)
    .bind(s.ip_address)
    .bind(s.user_agent)
    .fetch_one(db)
    .await
}

pub async fn session_by_access_hash(
    db: impl PgExecutor<'_>,
    hash: &str,
) -> sqlx::Result<Option<Session>> {
    sqlx::query_as(&format!(
        "SELECT {SESSION_COLUMNS} FROM user_sessions WHERE access_token_hash = $1 AND NOT is_revoked"
    ))
    .bind(hash)
    .fetch_optional(db)
    .await
}

pub async fn session_by_refresh_hash(
    db: impl PgExecutor<'_>,
    hash: &str,
) -> sqlx::Result<Option<Session>> {
    sqlx::query_as(&format!(
        "SELECT {SESSION_COLUMNS} FROM user_sessions WHERE refresh_token_hash = $1 AND NOT is_revoked"
    ))
    .bind(hash)
    .fetch_optional(db)
    .await
}

/// Replaces a session in one transaction. The conditional revoke is the claim:
/// of two concurrent refreshes with the same token only one wins (`true`).
pub async fn rotate_session(
    tx: &mut PgConnection,
    old_id: i64,
    next: &NewSession<'_>,
) -> sqlx::Result<bool> {
    let claimed =
        sqlx::query("UPDATE user_sessions SET is_revoked = true WHERE id = $1 AND NOT is_revoked")
            .bind(old_id)
            .execute(&mut *tx)
            .await?
            .rows_affected();
    if claimed == 0 {
        return Ok(false);
    }
    insert_session(&mut *tx, next).await?;
    Ok(true)
}

pub async fn touch_session(db: impl PgExecutor<'_>, id: i64) -> sqlx::Result<()> {
    sqlx::query("UPDATE user_sessions SET last_used_at = NOW() WHERE id = $1")
        .bind(id)
        .execute(db)
        .await?;
    Ok(())
}

pub async fn revoke_session(db: impl PgExecutor<'_>, id: i64) -> sqlx::Result<()> {
    sqlx::query("UPDATE user_sessions SET is_revoked = true WHERE id = $1")
        .bind(id)
        .execute(db)
        .await?;
    Ok(())
}
