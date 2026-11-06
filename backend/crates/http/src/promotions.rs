//! `/api/v1/promotions`: same requests, responses, and errors as the Go
//! `PromotionHandler`/`PromotionService`.
//!
//! Go quirk kept for the web app: create and update echo `start_at`/`end_at` with
//! the offset they were sent in, and decimals as sent; reads show the stored values.

use axum::{Router, extract::State, http::StatusCode, routing::get};
use chrono::{DateTime, FixedOffset, Utc};
use nexpos_db::promotions::{self as db, ListFilter, Promotion, PromotionInput};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
    time,
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/promotions", get(list).post(create))
        .route(
            "/api/v1/promotions/{id}",
            get(show).put(update).delete(destroy),
        )
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct PromotionRequest {
    code: Option<String>,
    name: Option<String>,
    description: Option<String>,
    #[serde(rename = "type")]
    kind: Option<String>,
    discount_type: Option<String>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    discount_value: Option<Decimal>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    min_purchase: Option<Decimal>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    max_discount: Option<Decimal>,
    start_at: Option<String>,
    end_at: Option<String>,
    priority: Option<i32>,
    is_active: Option<bool>,
}

/// A request that passed validation, with its times parsed.
struct Parsed<'a> {
    input: PromotionInput<'a>,
    start_at: DateTime<FixedOffset>,
    end_at: Option<DateTime<FixedOffset>>,
}

impl PromotionRequest {
    fn code(&self) -> &str {
        self.code.as_deref().unwrap_or_default()
    }

    fn validate(&self) -> AppResult<()> {
        let non_negative = |v: Option<Decimal>| v.is_none_or(|d| d >= Decimal::ZERO);
        Validator::new()
            .field(
                "code",
                self.code(),
                &[Rule::Required, Rule::Min(1), Rule::Max(50)],
            )
            .field(
                "name",
                self.name.as_deref().unwrap_or_default(),
                &[Rule::Required, Rule::Min(1), Rule::Max(255)],
            )
            .optional(
                "description",
                self.description.as_deref(),
                &[Rule::Max(1000)],
            )
            .field(
                "type",
                self.kind.as_deref().unwrap_or_default(),
                &[Rule::Required, Rule::OneOf("discount bundle conditional")],
            )
            .optional(
                "discount_type",
                self.discount_type.as_deref(),
                &[Rule::OneOf("percentage fixed")],
            )
            .invalid_unless("discount_value", non_negative(self.discount_value))
            .invalid_unless("min_purchase", non_negative(self.min_purchase))
            .invalid_unless("max_discount", non_negative(self.max_discount))
            .field(
                "start_at",
                self.start_at.as_deref().unwrap_or_default(),
                &[Rule::Required],
            )
            .invalid_unless("priority", self.priority.unwrap_or(0) >= 0)
            .finish()
    }

    /// Times must be RFC 3339; Go checks them after the code.
    fn parse(&self) -> AppResult<Parsed<'_>> {
        let start_at = DateTime::parse_from_rfc3339(self.start_at.as_deref().unwrap_or_default())
            .map_err(|_| AppError::bad_request("Invalid start_at format"))?;
        let end_at = self
            .end_at
            .as_deref()
            .map(DateTime::parse_from_rfc3339)
            .transpose()
            .map_err(|_| AppError::bad_request("Invalid end_at format"))?;
        Ok(Parsed {
            input: PromotionInput {
                code: self.code(),
                name: self.name.as_deref().unwrap_or_default(),
                description: self.description.as_deref(),
                kind: self.kind.as_deref().unwrap_or_default(),
                discount_type: self.discount_type.as_deref(),
                discount_value: self.discount_value,
                min_purchase: self.min_purchase,
                max_discount: self.max_discount,
                start_at: start_at.with_timezone(&Utc),
                end_at: end_at.map(|t| t.with_timezone(&Utc)),
                priority: self.priority.unwrap_or(0),
                is_active: self.is_active.unwrap_or(false),
            },
            start_at,
            end_at,
        })
    }
}

#[derive(Serialize)]
struct PromotionResponse {
    id: i64,
    code: String,
    name: String,
    description: Option<String>,
    #[serde(rename = "type")]
    kind: String,
    discount_type: Option<String>,
    #[serde(serialize_with = "crate::money::serialize_some")]
    discount_value: Option<Decimal>,
    #[serde(serialize_with = "crate::money::serialize_some")]
    min_purchase: Option<Decimal>,
    #[serde(serialize_with = "crate::money::serialize_some")]
    max_discount: Option<Decimal>,
    start_at: String,
    end_at: Option<String>,
    priority: i32,
    is_active: bool,
    status: &'static str,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

/// Derived from the flag and the schedule, never stored.
fn status(is_active: bool, start_at: DateTime<Utc>, end_at: Option<DateTime<Utc>>) -> &'static str {
    let now = Utc::now();
    if !is_active {
        "inactive"
    } else if start_at > now {
        "scheduled"
    } else if end_at.is_some_and(|end| end < now) {
        "expired"
    } else {
        "active"
    }
}

impl From<Promotion> for PromotionResponse {
    fn from(p: Promotion) -> Self {
        Self {
            status: status(p.is_active, p.start_at, p.end_at),
            id: p.id,
            code: p.code,
            name: p.name,
            description: p.description,
            kind: p.kind,
            discount_type: p.discount_type,
            discount_value: p.discount_value,
            min_purchase: p.min_purchase,
            max_discount: p.max_discount,
            start_at: time::render(&p.start_at),
            end_at: p.end_at.as_ref().map(time::render),
            priority: p.priority,
            is_active: p.is_active,
            created_at: p.created_at,
            updated_at: p.updated_at,
        }
    }
}

fn echoed(
    id: i64,
    parsed: &Parsed<'_>,
    created_at: DateTime<Utc>,
    updated_at: DateTime<Utc>,
) -> PromotionResponse {
    let p = &parsed.input;
    PromotionResponse {
        id,
        code: p.code.to_owned(),
        name: p.name.to_owned(),
        description: p.description.map(str::to_owned),
        kind: p.kind.to_owned(),
        discount_type: p.discount_type.map(str::to_owned),
        discount_value: p.discount_value,
        min_purchase: p.min_purchase,
        max_discount: p.max_discount,
        start_at: time::go_rfc3339_fixed(&parsed.start_at),
        end_at: parsed.end_at.as_ref().map(time::go_rfc3339_fixed),
        priority: p.priority,
        is_active: p.is_active,
        status: status(p.is_active, p.start_at, p.end_at),
        created_at,
        updated_at,
    }
}

/// Go's message is "Promotion not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Promotion not found")
}

fn code_taken() -> AppError {
    AppError::bad_request("Promotion code already exists")
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<PromotionResponse>>> {
    let (page, per_page) = query.page();
    let filter = ListFilter {
        search: query.get("search"),
        is_active: query.bool("is_active"),
        kind: query.get("type"),
        status: query.get("status"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list(&state.db, user.company_id, &filter).await?;
    let pagination = Pagination::new(total, page, per_page, 0);
    let data = rows.into_iter().map(PromotionResponse::from).collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Promotion list retrieved",
        data,
        pagination,
        None,
    ))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<PromotionRequest>,
) -> AppResult<Reply<PromotionResponse>> {
    req.validate()?;
    if db::code_exists(&state.db, user.company_id, req.code(), 0).await? {
        return Err(code_taken());
    }
    let parsed = req.parse()?;
    let (id, created_at, updated_at) =
        db::create(&state.db, user.company_id, &parsed.input).await?;
    Ok(success(
        StatusCode::CREATED,
        "Promotion created successfully",
        echoed(id, &parsed, created_at, updated_at),
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<PromotionResponse>> {
    let id = parse_id(&id, "promotion")?;
    let row = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Promotion retrieved successfully",
        row.into(),
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<PromotionRequest>,
) -> AppResult<Reply<PromotionResponse>> {
    let id = parse_id(&id, "promotion")?;
    req.validate()?;
    let current = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if db::code_exists(&state.db, user.company_id, req.code(), id).await? {
        return Err(code_taken());
    }
    let parsed = req.parse()?;
    let updated_at = db::update(&state.db, user.company_id, id, &parsed.input)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Promotion updated successfully",
        echoed(id, &parsed, current.created_at, updated_at),
    ))
}

async fn destroy(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_id(&id, "promotion")?;
    if !db::delete(&state.db, user.company_id, id).await? {
        return Err(not_found());
    }
    Ok(success_message(
        StatusCode::OK,
        "Promotion deleted successfully",
    ))
}
