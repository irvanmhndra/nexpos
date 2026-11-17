//! The platform panel's own tables (`platform_*`) and its cross-company reads.
//! Every company read takes a `Scope`, so staff limited to some companies can
//! never reach the others.

use std::collections::HashMap;

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use serde_json::Value;
use sqlx::{PgConnection, PgExecutor, PgPool, Postgres, QueryBuilder, types::Json};

/// Which companies a staff member may see.
#[derive(Debug, Clone, Copy)]
pub struct Scope {
    pub staff_id: i64,
    pub all_companies: bool,
}

impl Scope {
    /// Appends ` AND <column> IN (their companies)` unless they see everything.
    fn restrict(&self, q: &mut QueryBuilder<'_, Postgres>, column: &str) {
        if !self.all_companies {
            q.push(format!(
                " AND {column} IN (SELECT company_id FROM platform_user_companies WHERE platform_user_id = "
            ))
            .push_bind(self.staff_id)
            .push(")");
        }
    }
}

// ============== Staff ==============

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Staff {
    pub id: i64,
    pub email: String,
    pub name: String,
    pub password_hash: String,
    pub role: String,
    pub all_companies: bool,
    pub is_active: bool,
    pub failed_logins: i32,
    pub locked_until: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

const STAFF_COLUMNS: &str = "id, email, name, password_hash, role, all_companies, is_active, \
     failed_logins, locked_until, created_at, updated_at";

pub async fn staff_by_email(db: impl PgExecutor<'_>, email: &str) -> sqlx::Result<Option<Staff>> {
    sqlx::query_as(&format!(
        "SELECT {STAFF_COLUMNS} FROM platform_users WHERE LOWER(email) = LOWER($1)"
    ))
    .bind(email)
    .fetch_optional(db)
    .await
}

pub async fn staff_by_id(db: impl PgExecutor<'_>, id: i64) -> sqlx::Result<Option<Staff>> {
    sqlx::query_as(&format!(
        "SELECT {STAFF_COLUMNS} FROM platform_users WHERE id = $1"
    ))
    .bind(id)
    .fetch_optional(db)
    .await
}

pub async fn list_staff(db: impl PgExecutor<'_>) -> sqlx::Result<Vec<Staff>> {
    sqlx::query_as(&format!(
        "SELECT {STAFF_COLUMNS} FROM platform_users ORDER BY id"
    ))
    .fetch_all(db)
    .await
}

pub async fn create_staff(
    db: impl PgExecutor<'_>,
    email: &str,
    name: &str,
    password_hash: &str,
    role: &str,
    all_companies: bool,
) -> sqlx::Result<Staff> {
    sqlx::query_as(&format!(
        "INSERT INTO platform_users (email, name, password_hash, role, all_companies) \
         VALUES ($1, $2, $3, $4, $5) RETURNING {STAFF_COLUMNS}"
    ))
    .bind(email)
    .bind(name)
    .bind(password_hash)
    .bind(role)
    .bind(all_companies)
    .fetch_one(db)
    .await
}

/// Saves the editable fields of `staff`.
pub async fn update_staff(db: impl PgExecutor<'_>, staff: &Staff) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE platform_users SET name = $1, role = $2, all_companies = $3, is_active = $4, \
             password_hash = $5, updated_at = NOW() WHERE id = $6",
    )
    .bind(&staff.name)
    .bind(&staff.role)
    .bind(staff.all_companies)
    .bind(staff.is_active)
    .bind(&staff.password_hash)
    .bind(staff.id)
    .execute(db)
    .await?;
    Ok(())
}

/// Replaces the companies a staff member is limited to.
pub async fn set_companies(
    conn: &mut PgConnection,
    staff_id: i64,
    company_ids: &[i64],
) -> sqlx::Result<()> {
    sqlx::query("DELETE FROM platform_user_companies WHERE platform_user_id = $1")
        .bind(staff_id)
        .execute(&mut *conn)
        .await?;
    sqlx::query(
        "INSERT INTO platform_user_companies (platform_user_id, company_id) \
         SELECT $1, id FROM companies WHERE id = ANY($2)",
    )
    .bind(staff_id)
    .bind(company_ids)
    .execute(&mut *conn)
    .await?;
    Ok(())
}

/// How many of these ids are real companies.
pub async fn count_companies(db: impl PgExecutor<'_>, ids: &[i64]) -> sqlx::Result<i64> {
    sqlx::query_scalar("SELECT COUNT(*) FROM companies WHERE id = ANY($1)")
        .bind(ids)
        .fetch_one(db)
        .await
}

pub async fn companies_of(
    db: impl PgExecutor<'_>,
    staff_ids: &[i64],
) -> sqlx::Result<HashMap<i64, Vec<i64>>> {
    let rows: Vec<(i64, i64)> = sqlx::query_as(
        "SELECT platform_user_id, company_id FROM platform_user_companies \
         WHERE platform_user_id = ANY($1) ORDER BY company_id",
    )
    .bind(staff_ids)
    .fetch_all(db)
    .await?;
    let mut by_staff: HashMap<i64, Vec<i64>> = HashMap::new();
    for (staff, company) in rows {
        by_staff.entry(staff).or_default().push(company);
    }
    Ok(by_staff)
}

/// Counts a wrong password; the account locks once `max` is reached.
pub async fn record_failed_login(
    db: impl PgExecutor<'_>,
    staff_id: i64,
    max: i32,
    lock_minutes: i32,
) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE platform_users SET failed_logins = failed_logins + 1, \
             locked_until = CASE WHEN failed_logins + 1 >= $2 \
                 THEN NOW() + make_interval(mins => $3) ELSE locked_until END \
         WHERE id = $1",
    )
    .bind(staff_id)
    .bind(max)
    .bind(lock_minutes)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn clear_failed_logins(db: impl PgExecutor<'_>, staff_id: i64) -> sqlx::Result<()> {
    sqlx::query("UPDATE platform_users SET failed_logins = 0, locked_until = NULL WHERE id = $1")
        .bind(staff_id)
        .execute(db)
        .await?;
    Ok(())
}

// ============== Sessions ==============

pub async fn insert_session(
    db: impl PgExecutor<'_>,
    staff_id: i64,
    token_hash: &str,
    expires_at: DateTime<Utc>,
    ip_address: &str,
    user_agent: &str,
) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO platform_sessions (platform_user_id, token_hash, expires_at, ip_address, user_agent) \
         VALUES ($1, $2, $3, $4, $5)",
    )
    .bind(staff_id)
    .bind(token_hash)
    .bind(expires_at)
    .bind(ip_address)
    .bind(user_agent)
    .execute(db)
    .await?;
    Ok(())
}

/// The staff member behind a live token: not revoked, not expired, still active.
pub async fn staff_by_token(
    db: impl PgExecutor<'_>,
    token_hash: &str,
) -> sqlx::Result<Option<Staff>> {
    sqlx::query_as(
        "UPDATE platform_sessions s SET last_used_at = NOW() FROM platform_users u \
         WHERE s.token_hash = $1 AND NOT s.is_revoked AND s.expires_at > NOW() \
           AND u.id = s.platform_user_id AND u.is_active \
         RETURNING u.id, u.email, u.name, u.password_hash, u.role, u.all_companies, u.is_active, \
             u.failed_logins, u.locked_until, u.created_at, u.updated_at",
    )
    .bind(token_hash)
    .fetch_optional(db)
    .await
}

pub async fn revoke_session(db: impl PgExecutor<'_>, token_hash: &str) -> sqlx::Result<()> {
    sqlx::query("UPDATE platform_sessions SET is_revoked = true WHERE token_hash = $1")
        .bind(token_hash)
        .execute(db)
        .await?;
    Ok(())
}

pub async fn revoke_sessions_of(db: impl PgExecutor<'_>, staff_id: i64) -> sqlx::Result<()> {
    sqlx::query("UPDATE platform_sessions SET is_revoked = true WHERE platform_user_id = $1")
        .bind(staff_id)
        .execute(db)
        .await?;
    Ok(())
}

// ============== Companies ==============

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct CompanyRow {
    pub id: i64,
    pub code: String,
    pub name: String,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
    pub users: i64,
    pub branches: i64,
    pub orders_30d: i64,
    pub revenue_30d: Decimal,
    pub last_order_at: Option<DateTime<Utc>>,
}

pub struct CompanyFilter<'a> {
    pub search: &'a str,
    /// "active", "suspended", or anything else for both.
    pub status: &'a str,
    pub limit: i64,
    pub offset: i64,
}

const COMPANY_SELECT: &str = "SELECT c.id, c.code, c.name, c.is_active, c.created_at, \
     (SELECT COUNT(*) FROM users u WHERE u.company_id = c.id) AS users, \
     (SELECT COUNT(*) FROM branches b WHERE b.company_id = c.id) AS branches, \
     (SELECT COUNT(*) FROM orders o WHERE o.company_id = c.id AND o.status = 'completed' \
         AND o.created_at >= NOW() - interval '30 days') AS orders_30d, \
     (SELECT COALESCE(SUM(o.grand_total), 0) FROM orders o WHERE o.company_id = c.id \
         AND o.status = 'completed' AND o.created_at >= NOW() - interval '30 days') AS revenue_30d, \
     (SELECT MAX(o.created_at) FROM orders o WHERE o.company_id = c.id) AS last_order_at \
     FROM companies c WHERE true";

pub async fn list_companies(
    db: &PgPool,
    scope: Scope,
    f: &CompanyFilter<'_>,
) -> sqlx::Result<(Vec<CompanyRow>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        scope.restrict(q, "c.id");
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (c.name ILIKE ")
                .push_bind(like.clone())
                .push(" OR c.code ILIKE ")
                .push_bind(like)
                .push(")");
        }
        match f.status {
            "active" => {
                q.push(" AND c.is_active");
            }
            "suspended" => {
                q.push(" AND NOT c.is_active");
            }
            _ => {}
        }
    };
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM companies c WHERE true");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(COMPANY_SELECT);
    filter(&mut rows);
    rows.push(" ORDER BY c.id DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

/// One company, or `None` when it doesn't exist or is outside the scope.
pub async fn company(db: &PgPool, scope: Scope, id: i64) -> sqlx::Result<Option<CompanyRow>> {
    let mut q = QueryBuilder::new(COMPANY_SELECT);
    q.push(" AND c.id = ").push_bind(id);
    scope.restrict(&mut q, "c.id");
    q.build_query_as().fetch_optional(db).await
}

/// Sets `is_active` and returns whether it changed.
pub async fn set_company_active(
    db: impl PgExecutor<'_>,
    id: i64,
    active: bool,
) -> sqlx::Result<bool> {
    let changed = sqlx::query(
        "UPDATE companies SET is_active = $2, updated_at = NOW() WHERE id = $1 AND is_active <> $2",
    )
    .bind(id)
    .bind(active)
    .execute(db)
    .await?
    .rows_affected();
    Ok(changed > 0)
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct CompanyUser {
    pub id: i64,
    pub name: String,
    pub email: String,
    pub branch_id: Option<i64>,
}

/// The account support sessions run as: the owner, else the oldest active user.
pub async fn primary_user(
    db: impl PgExecutor<'_>,
    company_id: i64,
) -> sqlx::Result<Option<CompanyUser>> {
    sqlx::query_as(
        "SELECT u.id, u.name, u.email, \
             (SELECT ub.branch_id FROM user_branches ub WHERE ub.user_id = u.id \
              ORDER BY ub.is_default DESC, ub.branch_id LIMIT 1) AS branch_id \
         FROM users u LEFT JOIN roles r ON r.id = u.role_id \
         WHERE u.company_id = $1 AND u.status = 'active' \
         ORDER BY (r.code = 'owner') DESC NULLS LAST, u.id LIMIT 1",
    )
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn mark_impersonation(
    db: impl PgExecutor<'_>,
    session_id: i64,
    staff_id: i64,
    company_id: i64,
    reason: &str,
) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO platform_impersonations (session_id, platform_user_id, company_id, reason) \
         VALUES ($1, $2, $3, $4)",
    )
    .bind(session_id)
    .bind(staff_id)
    .bind(company_id)
    .bind(reason)
    .execute(db)
    .await?;
    Ok(())
}

// ============== Audit log ==============

pub struct NewAudit<'a> {
    pub staff_id: i64,
    pub actor_email: &'a str,
    pub action: &'a str,
    pub company_id: Option<i64>,
    pub target_type: &'a str,
    pub target_id: Option<i64>,
    pub reason: &'a str,
    pub before: Option<Value>,
    pub after: Option<Value>,
    pub ip_address: &'a str,
}

pub async fn record_audit(db: impl PgExecutor<'_>, a: NewAudit<'_>) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO platform_audit_logs (platform_user_id, actor_email, action, company_id, \
             target_type, target_id, reason, before, after, ip_address) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
    )
    .bind(a.staff_id)
    .bind(a.actor_email)
    .bind(a.action)
    .bind(a.company_id)
    .bind(a.target_type)
    .bind(a.target_id)
    .bind(a.reason)
    .bind(a.before.map(Json))
    .bind(a.after.map(Json))
    .bind(a.ip_address)
    .execute(db)
    .await?;
    Ok(())
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct AuditRow {
    pub id: i64,
    pub platform_user_id: i64,
    pub actor_email: String,
    pub action: String,
    pub company_id: Option<i64>,
    pub target_type: String,
    pub target_id: Option<i64>,
    pub reason: String,
    pub before: Option<Json<Value>>,
    pub after: Option<Json<Value>>,
    pub ip_address: Option<String>,
    pub created_at: DateTime<Utc>,
}

pub struct AuditFilter<'a> {
    pub company_id: Option<i64>,
    pub staff_id: Option<i64>,
    pub action: &'a str,
    pub limit: i64,
    pub offset: i64,
}

/// Newest first. Staff limited to some companies see only those companies'
/// lines, and none of the lines that aren't about a company.
pub async fn list_audit(
    db: &PgPool,
    scope: Scope,
    f: &AuditFilter<'_>,
) -> sqlx::Result<(Vec<AuditRow>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        scope.restrict(q, "company_id");
        if let Some(company) = f.company_id {
            q.push(" AND company_id = ").push_bind(company);
        }
        if let Some(staff) = f.staff_id {
            q.push(" AND platform_user_id = ").push_bind(staff);
        }
        if !f.action.is_empty() {
            q.push(" AND action = ").push_bind(f.action.to_owned());
        }
    };
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM platform_audit_logs WHERE true");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(
        "SELECT id, platform_user_id, actor_email, action, company_id, target_type, target_id, \
             reason, before, after, ip_address, created_at FROM platform_audit_logs WHERE true",
    );
    filter(&mut rows);
    rows.push(" ORDER BY id DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}
