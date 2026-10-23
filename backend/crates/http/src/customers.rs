//! `/api/v1/customers`: same requests, responses, and errors as the Go
//! `CustomerHandler`/`CustomerService`.

use axum::{Router, extract::State, http::StatusCode, routing::get};
use chrono::{DateTime, Utc};
use nexpos_db::customers::{self as db, Customer, CustomerInput, ListFilter};
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/customers", get(list).post(create))
        .route(
            "/api/v1/customers/{id}",
            get(show).put(update).delete(destroy),
        )
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CustomerRequest {
    code: Option<String>,
    name: Option<String>,
    phone: Option<String>,
    email: Option<String>,
    is_member: Option<bool>,
}

impl CustomerRequest {
    /// `CreateCustomerRequest` / `UpdateCustomerRequest` tags.
    fn validate(&self) -> AppResult<CustomerInput<'_>> {
        let code = self.code.as_deref().unwrap_or_default();
        let name = self.name.as_deref().unwrap_or_default();
        Validator::new()
            .field("code", code, &[Rule::Required, Rule::Min(1), Rule::Max(50)])
            .field(
                "name",
                name,
                &[Rule::Required, Rule::Min(1), Rule::Max(255)],
            )
            .optional("phone", self.phone.as_deref(), &[Rule::Max(50)])
            .optional(
                "email",
                self.email.as_deref(),
                &[Rule::Email, Rule::Max(255)],
            )
            .finish()?;
        Ok(CustomerInput {
            code,
            name,
            phone: self.phone.as_deref(),
            email: self.email.as_deref(),
            is_member: self.is_member.unwrap_or(false),
        })
    }
}

#[derive(Serialize)]
struct CustomerResponse {
    id: i64,
    code: String,
    name: String,
    phone: Option<String>,
    email: Option<String>,
    is_member: bool,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<Customer> for CustomerResponse {
    fn from(c: Customer) -> Self {
        Self {
            id: c.id,
            code: c.code,
            name: c.name,
            phone: c.phone,
            email: c.email,
            is_member: c.is_member,
            created_at: c.created_at,
            updated_at: c.updated_at,
        }
    }
}

/// Go's message is "Customer not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Customer not found")
}

fn code_taken() -> AppError {
    AppError::bad_request("Customer code already exists")
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<CustomerResponse>>> {
    let (page, per_page) = query.page();
    let filter = ListFilter {
        search: query.get("search"),
        is_member: query.bool("is_member"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list(&state.db, user.company_id, &filter).await?;
    // The Go handler never sets `count`, so it is always 0.
    let pagination = Pagination::new(total, page, per_page, 0);
    let data = rows.into_iter().map(CustomerResponse::from).collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Customer list retrieved",
        data,
        pagination,
        None,
    ))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<CustomerRequest>,
) -> AppResult<Reply<CustomerResponse>> {
    let input = req.validate()?;
    if db::code_exists(&state.db, user.company_id, input.code, 0).await? {
        return Err(code_taken());
    }
    let customer = db::create(&state.db, user.company_id, &input).await?;
    Ok(success(
        StatusCode::CREATED,
        "Customer created successfully",
        customer.into(),
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<CustomerResponse>> {
    let id = parse_id(&id, "customer")?;
    let customer = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Customer retrieved successfully",
        customer.into(),
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<CustomerRequest>,
) -> AppResult<Reply<CustomerResponse>> {
    let id = parse_id(&id, "customer")?;
    let input = req.validate()?;
    if db::get(&state.db, user.company_id, id).await?.is_none() {
        return Err(not_found());
    }
    if db::code_exists(&state.db, user.company_id, input.code, id).await? {
        return Err(code_taken());
    }
    let customer = db::update(&state.db, user.company_id, id, &input)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Customer updated successfully",
        customer.into(),
    ))
}

async fn destroy(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_id(&id, "customer")?;
    if !db::delete(&state.db, user.company_id, id).await? {
        return Err(not_found());
    }
    Ok(success_message(
        StatusCode::OK,
        "Customer deleted successfully",
    ))
}
