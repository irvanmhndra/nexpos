//! `/api/v1/branches`: same requests, responses, and errors as the Go
//! `BranchHandler`/`BranchService`.

use axum::{Router, extract::State, http::StatusCode, routing::get};
use chrono::{DateTime, Utc};
use nexpos_db::branches::{self as db, Branch, BranchInput, ListFilter};
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/branches", get(list).post(create))
        .route(
            "/api/v1/branches/{id}",
            get(show).put(update).delete(destroy),
        )
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct BranchRequest {
    code: Option<String>,
    name: Option<String>,
    address: Option<String>,
    phone: Option<String>,
    is_active: Option<bool>,
}

impl BranchRequest {
    /// `CreateBranchRequest` / `UpdateBranchRequest` tags.
    fn validate(&self) -> AppResult<BranchInput<'_>> {
        let code = self.code.as_deref().unwrap_or_default();
        let name = self.name.as_deref().unwrap_or_default();
        Validator::new()
            .field("code", code, &[Rule::Required, Rule::Min(1), Rule::Max(50)])
            .field(
                "name",
                name,
                &[Rule::Required, Rule::Min(1), Rule::Max(255)],
            )
            .optional("address", self.address.as_deref(), &[Rule::Max(500)])
            .optional("phone", self.phone.as_deref(), &[Rule::Max(50)])
            .finish()?;
        Ok(BranchInput {
            code,
            name,
            address: self.address.as_deref(),
            phone: self.phone.as_deref(),
            is_active: self.is_active.unwrap_or(false),
        })
    }
}

#[derive(Serialize)]
struct BranchResponse {
    id: i64,
    code: String,
    name: String,
    address: Option<String>,
    phone: Option<String>,
    is_active: bool,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<Branch> for BranchResponse {
    fn from(b: Branch) -> Self {
        Self {
            id: b.id,
            code: b.code,
            name: b.name,
            address: b.address,
            phone: b.phone,
            is_active: b.is_active,
            created_at: b.created_at,
            updated_at: b.updated_at,
        }
    }
}

/// Go passes "Branch not found" to `apperror.NotFound`, which appends
/// " not found" again; kept for wire compatibility.
fn not_found() -> AppError {
    AppError::not_found("Branch not found")
}

fn code_taken() -> AppError {
    AppError::bad_request("Branch code already exists")
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<BranchResponse>>> {
    let (page, per_page) = query.page();
    let filter = ListFilter {
        search: query.get("search"),
        is_active: query.bool("is_active"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list(&state.db, user.company_id, &filter).await?;
    // The Go handler never sets `count`, so it is always 0.
    let pagination = Pagination::new(total, page, per_page, 0);
    let data = rows.into_iter().map(BranchResponse::from).collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Branch list retrieved",
        data,
        pagination,
        None,
    ))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<BranchRequest>,
) -> AppResult<Reply<BranchResponse>> {
    let input = req.validate()?;
    if db::code_exists(&state.db, user.company_id, input.code, 0).await? {
        return Err(code_taken());
    }
    let branch = db::create(&state.db, user.company_id, &input).await?;
    Ok(success(
        StatusCode::CREATED,
        "Branch created successfully",
        branch.into(),
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<BranchResponse>> {
    let id = parse_id(&id, "branch")?;
    let branch = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Branch retrieved successfully",
        branch.into(),
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<BranchRequest>,
) -> AppResult<Reply<BranchResponse>> {
    let id = parse_id(&id, "branch")?;
    let input = req.validate()?;
    if db::get(&state.db, user.company_id, id).await?.is_none() {
        return Err(not_found());
    }
    if db::code_exists(&state.db, user.company_id, input.code, id).await? {
        return Err(code_taken());
    }
    let branch = db::update(&state.db, user.company_id, id, &input)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Branch updated successfully",
        branch.into(),
    ))
}

async fn destroy(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_id(&id, "branch")?;
    if !db::delete(&state.db, user.company_id, id).await? {
        return Err(not_found());
    }
    Ok(success_message(
        StatusCode::OK,
        "Branch deleted successfully",
    ))
}
