//! `/api/platform/v1`: the internal team's panel. It exists so support work
//! goes through the application instead of the database: every read is limited
//! to the companies a staff member may see, every write needs a reason, and
//! every write leaves an audit line.
//!
//! Staff are not tenant users. They sign in here, with their own accounts and
//! tokens, and reach a company's screens through a read-only support session.

use std::time::Duration;

use axum::{
    Router,
    extract::{FromRequestParts, State},
    http::{HeaderMap, StatusCode, header, request::Parts},
    routing::{get, post, put},
};
use chrono::{DateTime, Utc};
use nexpos_db::{
    auth::{self as tenant, NewSession},
    platform::{
        self as db, AuditFilter, AuditRow, CompanyFilter, CompanyRow, NewAudit, Scope, Staff,
    },
};
use nexpos_domain::{
    auth::{hash_token, new_token},
    platform::{LOCK_MINUTES, MAX_FAILED_LOGINS, Permission, Role, allows},
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};

use crate::{
    AppError, AppResult, AppState,
    auth::{hash_password, verify_password},
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{ClientInfo, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

/// Staff sign in again each working day; there is no refresh token.
const SESSION_TTL: Duration = Duration::from_secs(8 * 3600);
/// Long enough to look into a ticket, short enough not to linger.
const SUPPORT_SESSION_TTL: Duration = Duration::from_secs(30 * 60);

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/platform/v1/auth/login", post(login))
        .route("/api/platform/v1/auth/logout", post(logout))
        .route("/api/platform/v1/auth/me", get(me))
        .route("/api/platform/v1/companies", get(list_companies))
        .route("/api/platform/v1/companies/{id}", get(show_company))
        .route("/api/platform/v1/companies/{id}/suspend", post(suspend))
        .route("/api/platform/v1/companies/{id}/activate", post(activate))
        .route(
            "/api/platform/v1/companies/{id}/support-session",
            post(support_session),
        )
        .route("/api/platform/v1/audit-logs", get(audit_logs))
        .route("/api/platform/v1/staff", get(list_staff).post(create_staff))
        .route("/api/platform/v1/staff/{id}", put(update_staff))
}

// ============== Who is asking ==============

/// The signed-in staff member, from `Authorization: Bearer <platform token>`.
/// A tenant token is not accepted here, and a platform token is not accepted
/// by the tenant API: they live in different tables.
pub struct PlatformUser {
    pub id: i64,
    pub email: String,
    pub name: String,
    pub role: Role,
    all_companies: bool,
}

impl PlatformUser {
    fn scope(&self) -> Scope {
        Scope {
            staff_id: self.id,
            all_companies: self.all_companies || self.role == Role::Admin,
        }
    }

    fn require(&self, permission: Permission) -> AppResult<()> {
        if allows(self.role, permission) {
            Ok(())
        } else {
            Err(AppError::Forbidden(
                "your role does not allow this action".into(),
            ))
        }
    }
}

fn bearer(headers: &HeaderMap) -> Option<&str> {
    headers
        .get(header::AUTHORIZATION)
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.strip_prefix("Bearer "))
        .filter(|t| !t.is_empty())
}

impl FromRequestParts<AppState> for PlatformUser {
    type Rejection = AppError;

    async fn from_request_parts(parts: &mut Parts, state: &AppState) -> Result<Self, AppError> {
        let token = bearer(&parts.headers).ok_or_else(|| AppError::unauthorized("unauthorized"))?;
        let staff = db::staff_by_token(&state.db, &hash_token(token))
            .await?
            .ok_or_else(|| AppError::unauthorized("invalid or expired token"))?;
        let role = Role::parse(&staff.role)
            .ok_or_else(|| AppError::unauthorized("invalid or expired token"))?;
        Ok(Self {
            id: staff.id,
            email: staff.email,
            name: staff.name,
            role,
            all_companies: staff.all_companies,
        })
    }
}

/// Every write carries why it was done; it goes into the audit log.
#[derive(Default, Deserialize)]
#[serde(default)]
struct ReasonRequest {
    reason: Option<String>,
}

fn reason_rules() -> [Rule; 3] {
    [Rule::Required, Rule::Min(5), Rule::Max(500)]
}

impl ReasonRequest {
    fn validated(&self) -> AppResult<&str> {
        let reason = self.reason.as_deref().unwrap_or_default().trim();
        Validator::new()
            .field("reason", reason, &reason_rules())
            .finish()?;
        Ok(reason)
    }
}

// ============== Sign in ==============

#[derive(Default, Deserialize)]
#[serde(default)]
struct LoginRequest {
    email: Option<String>,
    password: Option<String>,
}

#[derive(Serialize)]
struct StaffResponse {
    id: i64,
    email: String,
    name: String,
    role: String,
    all_companies: bool,
    company_ids: Vec<i64>,
    is_active: bool,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

fn staff_response(s: Staff, company_ids: Vec<i64>) -> StaffResponse {
    StaffResponse {
        id: s.id,
        email: s.email,
        name: s.name,
        role: s.role,
        all_companies: s.all_companies,
        company_ids,
        is_active: s.is_active,
        created_at: s.created_at,
        updated_at: s.updated_at,
    }
}

async fn with_companies(state: &AppState, staff: Staff) -> AppResult<StaffResponse> {
    let mut companies = db::companies_of(&state.db, &[staff.id]).await?;
    let ids = companies.remove(&staff.id).unwrap_or_default();
    Ok(staff_response(staff, ids))
}

#[derive(Serialize)]
struct LoginResponse {
    access_token: String,
    expires_in: u64,
    user: StaffResponse,
}

async fn login(
    State(state): State<AppState>,
    client: ClientInfo,
    JsonBody(req): JsonBody<LoginRequest>,
) -> AppResult<Reply<LoginResponse>> {
    let email = req.email.unwrap_or_default();
    let password = req.password.unwrap_or_default();
    Validator::new()
        .field("email", &email, &[Rule::Required, Rule::Email])
        .field("password", &password, &[Rule::Required])
        .finish()?;

    // One answer for every way this can fail, so it tells an attacker nothing
    // about which accounts exist, are locked, or are switched off.
    let staff = db::staff_by_email(&state.db, &email)
        .await?
        .ok_or_else(AppError::invalid_credentials)?;
    if staff.locked_until.is_some_and(|until| until > Utc::now()) {
        return Err(AppError::invalid_credentials());
    }
    if !verify_password(password, staff.password_hash.clone()).await? {
        db::record_failed_login(&state.db, staff.id, MAX_FAILED_LOGINS, LOCK_MINUTES).await?;
        return Err(AppError::invalid_credentials());
    }
    if !staff.is_active {
        return Err(AppError::invalid_credentials());
    }
    db::clear_failed_logins(&state.db, staff.id).await?;

    let token = new_token();
    db::insert_session(
        &state.db,
        staff.id,
        &hash_token(&token),
        Utc::now() + SESSION_TTL,
        &client.ip,
        &client.user_agent,
    )
    .await?;

    Ok(success(
        StatusCode::OK,
        "Login successful",
        LoginResponse {
            access_token: token,
            expires_in: SESSION_TTL.as_secs(),
            user: with_companies(&state, staff).await?,
        },
    ))
}

async fn logout(State(state): State<AppState>, headers: HeaderMap) -> AppResult<Reply<()>> {
    if let Some(token) = bearer(&headers) {
        db::revoke_session(&state.db, &hash_token(token)).await?;
    }
    Ok(success_message(StatusCode::OK, "Logged out successfully"))
}

async fn me(State(state): State<AppState>, user: PlatformUser) -> AppResult<Reply<StaffResponse>> {
    let staff = db::staff_by_id(&state.db, user.id)
        .await?
        .ok_or_else(|| AppError::unauthorized("invalid or expired token"))?;
    Ok(success(
        StatusCode::OK,
        "Profile retrieved",
        with_companies(&state, staff).await?,
    ))
}

// ============== Companies ==============

#[derive(Serialize)]
struct CompanyResponse {
    id: i64,
    code: String,
    name: String,
    is_active: bool,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    users: i64,
    branches: i64,
    orders_30d: i64,
    #[serde(serialize_with = "crate::money::serialize")]
    revenue_30d: Decimal,
    #[serde(serialize_with = "crate::time::serialize_some")]
    last_order_at: Option<DateTime<Utc>>,
}

impl From<CompanyRow> for CompanyResponse {
    fn from(c: CompanyRow) -> Self {
        Self {
            id: c.id,
            code: c.code,
            name: c.name,
            is_active: c.is_active,
            created_at: c.created_at,
            users: c.users,
            branches: c.branches,
            orders_30d: c.orders_30d,
            revenue_30d: c.revenue_30d,
            last_order_at: c.last_order_at,
        }
    }
}

/// A company the caller can't see is reported like one that doesn't exist.
fn company_not_found() -> AppError {
    AppError::NotFound("Company".into())
}

async fn visible_company(state: &AppState, user: &PlatformUser, id: i64) -> AppResult<CompanyRow> {
    db::company(&state.db, user.scope(), id)
        .await?
        .ok_or_else(company_not_found)
}

async fn list_companies(
    State(state): State<AppState>,
    user: PlatformUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<CompanyResponse>>> {
    user.require(Permission::ViewCompanies)?;
    let (page, per_page) = query.page();
    let filter = CompanyFilter {
        search: query.get("search"),
        status: query.get("status"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list_companies(&state.db, user.scope(), &filter).await?;
    let data: Vec<CompanyResponse> = rows.into_iter().map(CompanyResponse::from).collect();
    let pagination = Pagination::new(total, page, per_page, data.len() as i64);
    Ok(success_with_pagination(
        StatusCode::OK,
        "Companies retrieved",
        data,
        pagination,
        None,
    ))
}

async fn show_company(
    State(state): State<AppState>,
    user: PlatformUser,
    RawId(id): RawId,
) -> AppResult<Reply<CompanyResponse>> {
    user.require(Permission::ViewCompanies)?;
    let id = parse_id(&id, "company")?;
    let company = visible_company(&state, &user, id).await?;
    Ok(success(StatusCode::OK, "Company retrieved", company.into()))
}

async fn set_active(
    state: AppState,
    user: PlatformUser,
    client: ClientInfo,
    id: String,
    req: ReasonRequest,
    active: bool,
) -> AppResult<Reply<CompanyResponse>> {
    user.require(Permission::SuspendCompany)?;
    let id = parse_id(&id, "company")?;
    let reason = req.validated()?;
    visible_company(&state, &user, id).await?;

    // The change and its audit line commit together, or not at all.
    let mut tx = state.db.begin().await?;
    if !db::set_company_active(&mut *tx, id, active).await? {
        return Err(AppError::bad_request(if active {
            "Company is already active"
        } else {
            "Company is already suspended"
        }));
    }
    db::record_audit(
        &mut *tx,
        NewAudit {
            staff_id: user.id,
            actor_email: &user.email,
            action: if active {
                "company.activate"
            } else {
                "company.suspend"
            },
            company_id: Some(id),
            target_type: "company",
            target_id: Some(id),
            reason,
            before: Some(json!({ "is_active": !active })),
            after: Some(json!({ "is_active": active })),
            ip_address: &client.ip,
        },
    )
    .await?;
    tx.commit().await?;

    let company = visible_company(&state, &user, id).await?;
    Ok(success(
        StatusCode::OK,
        if active {
            "Company activated"
        } else {
            "Company suspended"
        },
        company.into(),
    ))
}

/// Suspends a company: its users can't sign in, and their open sessions stop
/// working on their next request.
async fn suspend(
    State(state): State<AppState>,
    user: PlatformUser,
    client: ClientInfo,
    RawId(id): RawId,
    JsonBody(req): JsonBody<ReasonRequest>,
) -> AppResult<Reply<CompanyResponse>> {
    set_active(state, user, client, id, req, false).await
}

async fn activate(
    State(state): State<AppState>,
    user: PlatformUser,
    client: ClientInfo,
    RawId(id): RawId,
    JsonBody(req): JsonBody<ReasonRequest>,
) -> AppResult<Reply<CompanyResponse>> {
    set_active(state, user, client, id, req, true).await
}

#[derive(Serialize)]
struct SupportUser {
    id: i64,
    name: String,
    email: String,
}

#[derive(Serialize)]
struct SupportSessionResponse {
    access_token: String,
    expires_in: u64,
    company_id: i64,
    user: SupportUser,
}

/// Opens a read-only tenant session as the company's owner, so staff see exactly
/// what the customer sees, in the ordinary web app. It can't be refreshed.
async fn support_session(
    State(state): State<AppState>,
    user: PlatformUser,
    client: ClientInfo,
    RawId(id): RawId,
    JsonBody(req): JsonBody<ReasonRequest>,
) -> AppResult<Reply<SupportSessionResponse>> {
    user.require(Permission::Impersonate)?;
    let id = parse_id(&id, "company")?;
    let reason = req.validated()?;
    let company = visible_company(&state, &user, id).await?;
    if !company.is_active {
        return Err(AppError::bad_request(
            "Company is suspended; activate it to open a support session",
        ));
    }
    let owner = db::primary_user(&state.db, id)
        .await?
        .ok_or_else(|| AppError::bad_request("Company has no active user to open a session as"))?;

    let access = new_token();
    // Stored because the column requires one, never handed out.
    let unused_refresh = hash_token(&new_token());
    let expires = Utc::now() + SUPPORT_SESSION_TTL;
    let user_agent = format!("platform-support:{}", user.email);

    let mut tx = state.db.begin().await?;
    let session_id = tenant::insert_session(
        &mut *tx,
        &NewSession {
            user_id: owner.id,
            company_id: id,
            branch_id: owner.branch_id,
            access_token_hash: &hash_token(&access),
            access_token_expires_at: expires,
            refresh_token_hash: &unused_refresh,
            refresh_token_expires_at: expires,
            device_info: Some("platform support session"),
            ip_address: Some(&client.ip),
            user_agent: Some(&user_agent),
        },
    )
    .await?;
    db::mark_impersonation(&mut *tx, session_id, user.id, id, reason).await?;
    db::record_audit(
        &mut *tx,
        NewAudit {
            staff_id: user.id,
            actor_email: &user.email,
            action: "company.support_session",
            company_id: Some(id),
            target_type: "user",
            target_id: Some(owner.id),
            reason,
            before: None,
            after: Some(json!({ "session_id": session_id, "as_user": owner.email })),
            ip_address: &client.ip,
        },
    )
    .await?;
    tx.commit().await?;

    Ok(success(
        StatusCode::CREATED,
        "Support session opened",
        SupportSessionResponse {
            access_token: access,
            expires_in: SUPPORT_SESSION_TTL.as_secs(),
            company_id: id,
            user: SupportUser {
                id: owner.id,
                name: owner.name,
                email: owner.email,
            },
        },
    ))
}

// ============== Audit log ==============

#[derive(Serialize)]
struct AuditResponse {
    id: i64,
    platform_user_id: i64,
    actor_email: String,
    action: String,
    company_id: Option<i64>,
    target_type: String,
    target_id: Option<i64>,
    reason: String,
    before: Option<Value>,
    after: Option<Value>,
    ip_address: Option<String>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
}

impl From<AuditRow> for AuditResponse {
    fn from(a: AuditRow) -> Self {
        Self {
            id: a.id,
            platform_user_id: a.platform_user_id,
            actor_email: a.actor_email,
            action: a.action,
            company_id: a.company_id,
            target_type: a.target_type,
            target_id: a.target_id,
            reason: a.reason,
            before: a.before.map(|j| j.0),
            after: a.after.map(|j| j.0),
            ip_address: a.ip_address,
            created_at: a.created_at,
        }
    }
}

async fn audit_logs(
    State(state): State<AppState>,
    user: PlatformUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<AuditResponse>>> {
    user.require(Permission::ViewAudit)?;
    let (page, per_page) = query.page();
    let filter = AuditFilter {
        company_id: query.int("company_id"),
        staff_id: query.int("platform_user_id"),
        action: query.get("action"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list_audit(&state.db, user.scope(), &filter).await?;
    let data: Vec<AuditResponse> = rows.into_iter().map(AuditResponse::from).collect();
    let pagination = Pagination::new(total, page, per_page, data.len() as i64);
    Ok(success_with_pagination(
        StatusCode::OK,
        "Audit logs retrieved",
        data,
        pagination,
        None,
    ))
}

// ============== Staff accounts (admins only) ==============

const ROLES: &str = "admin ops support";

#[derive(Default, Deserialize)]
#[serde(default)]
struct StaffRequest {
    email: Option<String>,
    name: Option<String>,
    password: Option<String>,
    role: Option<String>,
    all_companies: Option<bool>,
    company_ids: Option<Vec<i64>>,
    is_active: Option<bool>,
    reason: Option<String>,
}

fn staff_snapshot(staff: &Staff, company_ids: &[i64]) -> Value {
    json!({
        "email": staff.email,
        "name": staff.name,
        "role": staff.role,
        "all_companies": staff.all_companies,
        "company_ids": company_ids,
        "is_active": staff.is_active,
    })
}

async fn check_company_ids(state: &AppState, ids: &[i64]) -> AppResult<()> {
    let mut unique = ids.to_vec();
    unique.sort_unstable();
    unique.dedup();
    if db::count_companies(&state.db, &unique).await? != unique.len() as i64 {
        return Err(AppError::bad_request(
            "company_ids contains an unknown company",
        ));
    }
    Ok(())
}

async fn list_staff(
    State(state): State<AppState>,
    user: PlatformUser,
) -> AppResult<Reply<Vec<StaffResponse>>> {
    user.require(Permission::ManageStaff)?;
    let staff = db::list_staff(&state.db).await?;
    let ids: Vec<i64> = staff.iter().map(|s| s.id).collect();
    let mut companies = db::companies_of(&state.db, &ids).await?;
    let data = staff
        .into_iter()
        .map(|s| {
            let company_ids = companies.remove(&s.id).unwrap_or_default();
            staff_response(s, company_ids)
        })
        .collect();
    Ok(success(StatusCode::OK, "Staff retrieved", data))
}

async fn create_staff(
    State(state): State<AppState>,
    user: PlatformUser,
    client: ClientInfo,
    JsonBody(req): JsonBody<StaffRequest>,
) -> AppResult<Reply<StaffResponse>> {
    user.require(Permission::ManageStaff)?;
    let email = req.email.as_deref().unwrap_or_default().trim();
    let name = req.name.as_deref().unwrap_or_default().trim();
    let password = req.password.clone().unwrap_or_default();
    let role = req.role.as_deref().unwrap_or_default();
    let reason = req.reason.as_deref().unwrap_or_default().trim();
    Validator::new()
        .field(
            "email",
            email,
            &[Rule::Required, Rule::Email, Rule::Max(255)],
        )
        .field(
            "name",
            name,
            &[Rule::Required, Rule::Min(3), Rule::Max(255)],
        )
        .field("password", &password, &[Rule::Required, Rule::Min(12)])
        .field("role", role, &[Rule::Required, Rule::OneOf(ROLES)])
        .field("reason", reason, &reason_rules())
        .finish()?;
    let company_ids = req.company_ids.unwrap_or_default();
    check_company_ids(&state, &company_ids).await?;
    if db::staff_by_email(&state.db, email).await?.is_some() {
        return Err(AppError::email_already_exists());
    }

    let password_hash = hash_password(password).await?;
    let mut tx = state.db.begin().await?;
    let staff = db::create_staff(
        &mut *tx,
        email,
        name,
        &password_hash,
        role,
        req.all_companies.unwrap_or(false),
    )
    .await?;
    db::set_companies(&mut tx, staff.id, &company_ids).await?;
    db::record_audit(
        &mut *tx,
        NewAudit {
            staff_id: user.id,
            actor_email: &user.email,
            action: "staff.create",
            company_id: None,
            target_type: "platform_user",
            target_id: Some(staff.id),
            reason,
            before: None,
            after: Some(staff_snapshot(&staff, &company_ids)),
            ip_address: &client.ip,
        },
    )
    .await?;
    tx.commit().await?;

    Ok(success(
        StatusCode::CREATED,
        "Staff created",
        with_companies(&state, staff).await?,
    ))
}

/// Changes only the fields that were sent. Switching an account off, or
/// changing its password, signs it out everywhere.
async fn update_staff(
    State(state): State<AppState>,
    user: PlatformUser,
    client: ClientInfo,
    RawId(id): RawId,
    JsonBody(req): JsonBody<StaffRequest>,
) -> AppResult<Reply<StaffResponse>> {
    user.require(Permission::ManageStaff)?;
    let id = parse_id(&id, "staff")?;
    let reason = req.reason.as_deref().unwrap_or_default().trim();
    let mut v = Validator::new();
    v.optional("name", req.name.as_deref(), &[Rule::Min(3), Rule::Max(255)])
        .optional("password", req.password.as_deref(), &[Rule::Min(12)])
        .optional("role", req.role.as_deref(), &[Rule::OneOf(ROLES)])
        .field("reason", reason, &reason_rules());
    v.finish()?;

    let mut staff = db::staff_by_id(&state.db, id)
        .await?
        .ok_or_else(|| AppError::NotFound("Staff".into()))?;
    let old_companies = db::companies_of(&state.db, &[id])
        .await?
        .remove(&id)
        .unwrap_or_default();
    let before = staff_snapshot(&staff, &old_companies);

    // An admin who demotes or disables themselves can leave nobody able to undo it.
    let demotes_self = req.role.as_deref().is_some_and(|r| r != "admin");
    if id == user.id && (demotes_self || req.is_active == Some(false)) {
        return Err(AppError::bad_request(
            "You cannot demote or deactivate your own account",
        ));
    }
    if let Some(ids) = &req.company_ids {
        check_company_ids(&state, ids).await?;
    }

    if let Some(name) = req.name {
        staff.name = name.trim().to_owned();
    }
    if let Some(role) = req.role {
        staff.role = role;
    }
    if let Some(all) = req.all_companies {
        staff.all_companies = all;
    }
    let deactivated = staff.is_active && req.is_active == Some(false);
    if let Some(active) = req.is_active {
        staff.is_active = active;
    }
    let password_changed = req.password.is_some();
    if let Some(password) = req.password {
        staff.password_hash = hash_password(password).await?;
    }
    let company_ids = req.company_ids.unwrap_or(old_companies);

    let mut tx = state.db.begin().await?;
    db::update_staff(&mut *tx, &staff).await?;
    db::set_companies(&mut tx, id, &company_ids).await?;
    if deactivated || password_changed {
        db::revoke_sessions_of(&mut *tx, id).await?;
    }
    let mut after = staff_snapshot(&staff, &company_ids);
    if password_changed {
        after["password_changed"] = json!(true);
    }
    db::record_audit(
        &mut *tx,
        NewAudit {
            staff_id: user.id,
            actor_email: &user.email,
            action: "staff.update",
            company_id: None,
            target_type: "platform_user",
            target_id: Some(id),
            reason,
            before: Some(before),
            after: Some(after),
            ip_address: &client.ip,
        },
    )
    .await?;
    tx.commit().await?;

    let staff = db::staff_by_id(&state.db, id)
        .await?
        .ok_or_else(|| AppError::NotFound("Staff".into()))?;
    Ok(success(
        StatusCode::OK,
        "Staff updated",
        with_companies(&state, staff).await?,
    ))
}

/// Creates the first admin from `PLATFORM_ADMIN_EMAIL` / `PLATFORM_ADMIN_PASSWORD`
/// when no account with that email exists yet. Returns whether it created one.
pub async fn ensure_admin(db: &sqlx::PgPool, email: &str, password: &str) -> anyhow::Result<bool> {
    if db::staff_by_email(db, email).await?.is_some() {
        return Ok(false);
    }
    let hash = nexpos_domain::auth::hash_password(password)?;
    db::create_staff(db, email, "Platform Admin", &hash, "admin", true).await?;
    Ok(true)
}
