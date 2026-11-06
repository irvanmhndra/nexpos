//! `/api/v1/suppliers`: same requests, responses, and errors as the Go
//! `SupplierHandler` and the supplier part of `PurchaseOrderService`.

use axum::{Router, extract::State, http::StatusCode, routing::get};
use chrono::{DateTime, Utc};
use nexpos_db::suppliers::{self as db, ListFilter, Supplier, SupplierInput};
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/suppliers", get(list).post(create))
        .route(
            "/api/v1/suppliers/{id}",
            get(show).put(update).delete(destroy),
        )
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct SupplierRequest {
    name: Option<String>,
    code: Option<String>,
    contact_name: Option<String>,
    phone: Option<String>,
    email: Option<String>,
    address: Option<String>,
    notes: Option<String>,
    is_active: Option<bool>,
}

impl SupplierRequest {
    fn validate(&self) -> AppResult<SupplierInput<'_>> {
        let name = self.name.as_deref().unwrap_or_default();
        let code = self.code.as_deref().unwrap_or_default();
        Validator::new()
            .field(
                "name",
                name,
                &[Rule::Required, Rule::Min(1), Rule::Max(200)],
            )
            .field("code", code, &[Rule::Required, Rule::Min(1), Rule::Max(50)])
            .optional(
                "contact_name",
                self.contact_name.as_deref(),
                &[Rule::Max(100)],
            )
            .optional("phone", self.phone.as_deref(), &[Rule::Max(50)])
            .optional(
                "email",
                self.email.as_deref(),
                &[Rule::Email, Rule::Max(150)],
            )
            .finish()?;
        Ok(SupplierInput {
            name,
            code,
            contact_name: self.contact_name.as_deref(),
            phone: self.phone.as_deref(),
            email: self.email.as_deref(),
            address: self.address.as_deref(),
            notes: self.notes.as_deref(),
            is_active: self.is_active.unwrap_or(false),
        })
    }
}

#[derive(Serialize)]
struct SupplierResponse {
    id: i64,
    name: String,
    code: String,
    contact_name: Option<String>,
    phone: Option<String>,
    email: Option<String>,
    address: Option<String>,
    notes: Option<String>,
    is_active: bool,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<Supplier> for SupplierResponse {
    fn from(s: Supplier) -> Self {
        Self {
            id: s.id,
            name: s.name,
            code: s.code,
            contact_name: s.contact_name,
            phone: s.phone,
            email: s.email,
            address: s.address,
            notes: s.notes,
            is_active: s.is_active,
            created_at: s.created_at,
            updated_at: s.updated_at,
        }
    }
}

/// Go's message is "Supplier not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Supplier not found")
}

fn code_taken() -> AppError {
    AppError::bad_request("Supplier code already exists")
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<SupplierResponse>>> {
    let (page, per_page) = query.page();
    let filter = ListFilter {
        search: query.get("search"),
        is_active: query.bool("is_active"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list(&state.db, user.company_id, &filter).await?;
    // Go's `buildPagination` never sets `count`, so it is always 0.
    let pagination = Pagination::new(total, page, per_page, 0);
    let data = rows.into_iter().map(SupplierResponse::from).collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Supplier list retrieved",
        data,
        pagination,
        None,
    ))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<SupplierRequest>,
) -> AppResult<Reply<SupplierResponse>> {
    let input = req.validate()?;
    if db::code_taken(&state.db, user.company_id, input.code, 0).await? {
        return Err(code_taken());
    }
    let supplier = db::create(&state.db, user.company_id, &input).await?;
    Ok(success(
        StatusCode::CREATED,
        "Supplier created successfully",
        supplier.into(),
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<SupplierResponse>> {
    let id = parse_id(&id, "supplier")?;
    let supplier = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Supplier retrieved successfully",
        supplier.into(),
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<SupplierRequest>,
) -> AppResult<Reply<SupplierResponse>> {
    let id = parse_id(&id, "supplier")?;
    let input = req.validate()?;
    if db::get(&state.db, user.company_id, id).await?.is_none() {
        return Err(not_found());
    }
    if db::code_taken(&state.db, user.company_id, input.code, id).await? {
        return Err(code_taken());
    }
    let supplier = db::update(&state.db, user.company_id, id, &input)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Supplier updated successfully",
        supplier.into(),
    ))
}

async fn destroy(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_id(&id, "supplier")?;
    if !db::delete(&state.db, user.company_id, id).await? {
        return Err(not_found());
    }
    Ok(success_message(
        StatusCode::OK,
        "Supplier deleted successfully",
    ))
}
