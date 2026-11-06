//! `/api/v1/users`: same requests, responses, and errors as the Go
//! `UserHandler`/`UserService`.
//!
//! Go quirks kept for the web app: `role_id` is accepted but never stored, and an
//! update's response shows the `updated_at` from before the update.

use axum::{
    Router,
    extract::State,
    http::StatusCode,
    routing::{get, patch},
};
use chrono::{DateTime, Utc};
use nexpos_db::users::{self as db, UserRow};
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    auth::hash_password,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/users", get(list).post(create))
        .route("/api/v1/users/{id}", get(show).put(update).delete(destroy))
        .route("/api/v1/users/{id}/status", patch(update_status))
}

const DEFAULT_LIMIT: i64 = 10;
const MAX_LIMIT: i64 = 100;

#[derive(Default, Deserialize)]
#[serde(default)]
struct CreateRequest {
    email: Option<String>,
    password: Option<String>,
    name: Option<String>,
    /// Type-checked like Go's, but unused.
    #[serde(rename = "role_id")]
    _role_id: Option<i64>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct UpdateRequest {
    email: Option<String>,
    name: Option<String>,
    #[serde(rename = "role_id")]
    _role_id: Option<i64>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct StatusRequest {
    status: Option<String>,
}

#[derive(Serialize)]
struct UserResponse {
    id: i64,
    company_id: i64,
    role_id: Option<i64>,
    email: String,
    name: String,
    status: String,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<UserRow> for UserResponse {
    fn from(u: UserRow) -> Self {
        Self {
            id: u.id,
            company_id: u.company_id,
            role_id: u.role_id,
            email: u.email,
            name: u.name,
            status: u.status,
            created_at: u.created_at,
            updated_at: u.updated_at,
        }
    }
}

fn not_found() -> AppError {
    AppError::not_found("user")
}

fn parse_user_id(raw: &str) -> AppResult<i64> {
    // The only resource whose message says "id" in lowercase.
    raw.parse()
        .map_err(|_| AppError::bad_request("invalid user id"))
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<UserResponse>>> {
    let invalid = |_| AppError::bad_request("invalid query params");
    let page = query
        .strict_int("page")
        .map_err(invalid)?
        .filter(|p| *p >= 1)
        .unwrap_or(1);
    let limit = query
        .strict_int("limit")
        .map_err(invalid)?
        .filter(|l| (1..=MAX_LIMIT).contains(l))
        .unwrap_or(DEFAULT_LIMIT);

    let (rows, total) = db::list(&state.db, user.company_id, limit, (page - 1) * limit).await?;
    let pagination = Pagination::new(total, page, limit, rows.len() as i64);
    let data = rows.into_iter().map(UserResponse::from).collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "User list retrieved",
        data,
        pagination,
        None,
    ))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<CreateRequest>,
) -> AppResult<Reply<UserResponse>> {
    let email = req.email.unwrap_or_default();
    let password = req.password.unwrap_or_default();
    let name = req.name.unwrap_or_default();
    Validator::new()
        .field("email", &email, &[Rule::Required, Rule::Email])
        .field("password", &password, &[Rule::Required, Rule::Min(8)])
        .field(
            "name",
            &name,
            &[Rule::Required, Rule::Min(2), Rule::Max(100)],
        )
        .finish()?;

    if db::email_taken(&state.db, &email, 0).await? {
        return Err(AppError::email_already_exists());
    }
    let password_hash = hash_password(password).await?;
    let row = db::create(&state.db, user.company_id, &email, &password_hash, &name)
        .await
        .map_err(email_conflict)?;
    Ok(success(
        StatusCode::CREATED,
        "User created successfully",
        row.into(),
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<UserResponse>> {
    let id = parse_user_id(&id)?;
    let row = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "User retrieved successfully",
        row.into(),
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<UpdateRequest>,
) -> AppResult<Reply<UserResponse>> {
    let id = parse_user_id(&id)?;
    let email = req.email.unwrap_or_default();
    let name = req.name.unwrap_or_default();
    // Plain strings with `omitempty`: empty means "leave it".
    Validator::new()
        .optional(
            "email",
            (!email.is_empty()).then_some(email.as_str()),
            &[Rule::Email],
        )
        .optional(
            "name",
            (!name.is_empty()).then_some(name.as_str()),
            &[Rule::Min(2), Rule::Max(100)],
        )
        .finish()?;

    let mut row = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if !email.is_empty() && email != row.email {
        if db::email_taken(&state.db, &email, id).await? {
            return Err(AppError::email_already_exists());
        }
        row.email = email;
    }
    if !name.is_empty() {
        row.name = name;
    }
    if !db::update(&state.db, user.company_id, id, &row.email, &row.name)
        .await
        .map_err(email_conflict)?
    {
        return Err(not_found());
    }
    Ok(success(
        StatusCode::OK,
        "User updated successfully",
        row.into(),
    ))
}

async fn update_status(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<StatusRequest>,
) -> AppResult<Reply<()>> {
    let id = parse_user_id(&id)?;
    let status = req.status.unwrap_or_default();
    Validator::new()
        .field(
            "status",
            &status,
            &[Rule::Required, Rule::OneOf("active inactive")],
        )
        .finish()?;
    if !db::update_status(&state.db, user.company_id, id, &status).await? {
        return Err(not_found());
    }
    Ok(success_message(
        StatusCode::OK,
        "User status updated successfully",
    ))
}

async fn destroy(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_id(&id, "user").map_err(|_| AppError::bad_request("invalid user id"))?;
    if !db::delete(&state.db, user.company_id, id).await? {
        return Err(not_found());
    }
    Ok(success_message(StatusCode::OK, "User deleted successfully"))
}

/// Two requests racing for one email: the unique index decides.
fn email_conflict(err: sqlx::Error) -> AppError {
    if err
        .as_database_error()
        .is_some_and(|e| e.is_unique_violation())
    {
        AppError::email_already_exists()
    } else {
        err.into()
    }
}
