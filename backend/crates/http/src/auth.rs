//! `/api/v1/auth/*`: login, register, refresh, logout. Same requests,
//! responses, and errors as the Go `AuthHandler`/`AuthService`; sessions are
//! interchangeable, so a token issued here works on the Go API and vice versa.

use axum::{
    Router,
    extract::State,
    http::{HeaderMap, StatusCode, header},
    routing::post,
};
use chrono::Utc;
use nexpos_db::auth::{self as db, NewSession, Session};
use nexpos_domain::auth as rules;
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Reply, success, success_message},
    extract::{ClientInfo, JsonBody, Rule, Validator},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/auth/login", post(login))
        .route("/api/v1/auth/register", post(register))
        .route("/api/v1/auth/refresh", post(refresh))
        .route("/api/v1/auth/logout", post(logout))
}

// ============== Requests ==============

#[derive(Default, Deserialize)]
#[serde(default)]
struct LoginRequest {
    email: Option<String>,
    password: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct RegisterRequest {
    name: Option<String>,
    email: Option<String>,
    password: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct RefreshRequest {
    refresh_token: Option<String>,
}

// ============== Responses ==============

#[derive(Serialize)]
struct Info {
    id: i64,
    code: String,
    name: String,
}

impl From<db::Named> for Info {
    fn from(n: db::Named) -> Self {
        Self {
            id: n.id,
            code: n.code,
            name: n.name,
        }
    }
}

#[derive(Serialize)]
struct AuthUser {
    id: i64,
    name: String,
    email: String,
    role: String,
    company: Info,
    #[serde(skip_serializing_if = "Option::is_none")]
    default_branch: Option<Info>,
}

#[derive(Serialize)]
struct SessionResponse {
    user: AuthUser,
    access_token: String,
    refresh_token: String,
    expires_in: u64,
}

#[derive(Serialize)]
struct RefreshResponse {
    access_token: String,
    refresh_token: String,
    expires_in: u64,
}

// ============== Handlers ==============

async fn login(
    State(state): State<AppState>,
    client: ClientInfo,
    JsonBody(req): JsonBody<LoginRequest>,
) -> AppResult<Reply<SessionResponse>> {
    let email = req.email.unwrap_or_default();
    let password = req.password.unwrap_or_default();
    Validator::new()
        .field("email", &email, &[Rule::Required, Rule::Email])
        .field("password", &password, &[Rule::Required])
        .finish()?;

    let user = db::user_by_email(&state.db, &email)
        .await?
        .ok_or_else(AppError::invalid_credentials)?;
    if !verify_password(password, user.password_hash.clone()).await? {
        return Err(AppError::invalid_credentials());
    }
    if !user.is_active() {
        return Err(AppError::user_inactive());
    }

    // Go ignores lookup errors here: no default branch, role "user".
    let branch = db::default_branch(&state.db, user.id).await.ok().flatten();
    let tokens = Tokens::new(&state);
    db::insert_session(
        &state.db,
        &tokens.session(
            user.id,
            user.company_id,
            branch.as_ref().map(|b| b.id),
            &client,
        ),
    )
    .await?;

    let company = db::company(&state.db, user.company_id).await?;
    let role = match user.role_id {
        Some(id) => db::role_code(&state.db, id).await.ok().flatten(),
        None => None,
    }
    .unwrap_or_else(|| "user".into());

    let body = tokens.response(AuthUser {
        id: user.id,
        name: user.name,
        email: user.email,
        role,
        company: company.into(),
        default_branch: branch.map(Info::from),
    });
    Ok(success(StatusCode::OK, "Login successful", body))
}

async fn register(
    State(state): State<AppState>,
    client: ClientInfo,
    JsonBody(req): JsonBody<RegisterRequest>,
) -> AppResult<Reply<SessionResponse>> {
    let name = req.name.unwrap_or_default();
    let email = req.email.unwrap_or_default();
    let password = req.password.unwrap_or_default();
    Validator::new()
        .field(
            "name",
            &name,
            &[Rule::Required, Rule::Min(3), Rule::Max(100)],
        )
        .field("email", &email, &[Rule::Required, Rule::Email])
        .field("password", &password, &[Rule::Required, Rule::Min(8)])
        .finish()?;

    if db::user_by_email(&state.db, &email).await?.is_some() {
        return Err(AppError::email_already_exists());
    }
    let (role_id, role) = db::system_role(&state.db, "owner")
        .await?
        .ok_or_else(|| AppError::Internal(anyhow::anyhow!("owner role not found")))?;
    let password_hash = hash_password(password).await?;
    let tokens = Tokens::new(&state);

    // Go creates these rows one by one; here a failure leaves nothing behind.
    let mut tx = state.db.begin().await?;
    let company = db::create_company(
        &mut *tx,
        &rules::company_code(&name),
        &format!("{name}'s Business"),
    )
    .await?;
    let branch = db::create_branch(&mut *tx, company.id, "HQ", "Headquarters").await?;
    let user_id =
        match db::create_user(&mut *tx, company.id, role_id, &email, &password_hash, &name).await {
            Err(e) if is_unique_violation(&e) => return Err(AppError::email_already_exists()),
            other => other?,
        };
    db::assign_branch(&mut *tx, user_id, branch.id, true).await?;
    db::insert_session(
        &mut *tx,
        &tokens.session(user_id, company.id, Some(branch.id), &client),
    )
    .await?;
    tx.commit().await?;

    let body = tokens.response(AuthUser {
        id: user_id,
        name,
        email,
        role,
        company: company.into(),
        default_branch: Some(branch.into()),
    });
    Ok(success(
        StatusCode::CREATED,
        "Registration successful",
        body,
    ))
}

async fn refresh(
    State(state): State<AppState>,
    JsonBody(req): JsonBody<RefreshRequest>,
) -> AppResult<Reply<RefreshResponse>> {
    let token = req.refresh_token.unwrap_or_default();
    Validator::new()
        .field("refresh_token", &token, &[Rule::Required])
        .finish()?;

    let session = db::session_by_refresh_hash(&state.db, &rules::hash_token(&token))
        .await?
        .ok_or_else(|| AppError::unauthorized("invalid refresh token"))?;
    if Utc::now() > session.refresh_token_expires_at {
        return Err(AppError::unauthorized("refresh token expired"));
    }

    let tokens = Tokens::new(&state);
    let mut tx = state.db.begin().await?;
    if !db::rotate_session(&mut tx, session.id, &tokens.replacing(&session)).await? {
        // Another request exchanged this refresh token first.
        return Err(AppError::unauthorized("invalid refresh token"));
    }
    tx.commit().await?;

    let body = RefreshResponse {
        access_token: tokens.access,
        refresh_token: tokens.refresh,
        expires_in: state.auth.access_ttl.as_secs(),
    };
    Ok(success(
        StatusCode::OK,
        "Token refreshed successfully",
        body,
    ))
}

/// Public, like in Go: revokes the session of the bearer token if there is one,
/// and always answers 200.
async fn logout(State(state): State<AppState>, headers: HeaderMap) -> Reply<()> {
    let token = headers
        .get(header::AUTHORIZATION)
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.strip_prefix("Bearer "))
        .filter(|t| !t.is_empty());
    if let Some(token) = token
        && let Ok(Some(session)) =
            db::session_by_access_hash(&state.db, &rules::hash_token(token)).await
    {
        let _ = db::revoke_session(&state.db, session.id).await;
    }
    success_message(StatusCode::OK, "Logged out successfully")
}

// ============== Helpers ==============

/// A fresh access/refresh pair and their expiry times.
struct Tokens {
    access: String,
    refresh: String,
    access_hash: String,
    refresh_hash: String,
    access_expires: chrono::DateTime<Utc>,
    refresh_expires: chrono::DateTime<Utc>,
    access_ttl_secs: u64,
}

impl Tokens {
    fn new(state: &AppState) -> Self {
        let (access, refresh) = (rules::new_token(), rules::new_token());
        let now = Utc::now();
        Self {
            access_hash: rules::hash_token(&access),
            refresh_hash: rules::hash_token(&refresh),
            access,
            refresh,
            access_expires: now + state.auth.access_ttl,
            refresh_expires: now + state.auth.refresh_ttl,
            access_ttl_secs: state.auth.access_ttl.as_secs(),
        }
    }

    fn session<'a>(
        &'a self,
        user_id: i64,
        company_id: i64,
        branch_id: Option<i64>,
        client: &'a ClientInfo,
    ) -> NewSession<'a> {
        NewSession {
            user_id,
            company_id,
            branch_id,
            access_token_hash: &self.access_hash,
            access_token_expires_at: self.access_expires,
            refresh_token_hash: &self.refresh_hash,
            refresh_token_expires_at: self.refresh_expires,
            device_info: None,
            ip_address: Some(&client.ip),
            user_agent: Some(&client.user_agent),
        }
    }

    /// The session that replaces `old` on refresh; it keeps the old client info.
    fn replacing<'a>(&'a self, old: &'a Session) -> NewSession<'a> {
        NewSession {
            user_id: old.user_id,
            company_id: old.company_id,
            branch_id: old.branch_id,
            access_token_hash: &self.access_hash,
            access_token_expires_at: self.access_expires,
            refresh_token_hash: &self.refresh_hash,
            refresh_token_expires_at: self.refresh_expires,
            device_info: old.device_info.as_deref(),
            ip_address: old.ip_address.as_deref(),
            user_agent: old.user_agent.as_deref(),
        }
    }

    fn response(self, user: AuthUser) -> SessionResponse {
        SessionResponse {
            user,
            access_token: self.access,
            refresh_token: self.refresh,
            expires_in: self.access_ttl_secs,
        }
    }
}

/// bcrypt is deliberately slow, so it runs off the async workers.
async fn hash_password(password: String) -> AppResult<String> {
    tokio::task::spawn_blocking(move || rules::hash_password(&password))
        .await
        .map_err(anyhow::Error::from)?
        .map_err(|e| AppError::Internal(e.into()))
}

async fn verify_password(password: String, hash: String) -> AppResult<bool> {
    Ok(
        tokio::task::spawn_blocking(move || rules::verify_password(&password, &hash))
            .await
            .map_err(anyhow::Error::from)?,
    )
}

fn is_unique_violation(err: &sqlx::Error) -> bool {
    err.as_database_error()
        .is_some_and(|e| e.is_unique_violation())
}
