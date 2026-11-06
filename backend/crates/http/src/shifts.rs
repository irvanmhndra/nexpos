//! `/api/v1/shifts`: same requests, responses, and errors as the Go `ShiftHandler`
//! and `ShiftService`.

use axum::{
    Router,
    extract::State,
    http::StatusCode,
    routing::{get, post},
};
use chrono::{DateTime, Utc};
use nexpos_db::{
    orders,
    shifts::{self as db, ListFilter, Shift},
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/shifts", get(list).post(open))
        .route("/api/v1/shifts/current", get(current))
        .route("/api/v1/shifts/{id}", get(show))
        .route("/api/v1/shifts/{id}/close", post(close))
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct OpenRequest {
    branch_id: Option<i64>,
    #[serde(deserialize_with = "crate::money::deserialize")]
    opening_float: Decimal,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CloseRequest {
    #[serde(deserialize_with = "crate::money::deserialize")]
    actual_cash: Decimal,
    notes: Option<String>,
}

#[derive(Serialize)]
struct ShiftResponse {
    id: i64,
    branch_id: i64,
    cashier_id: i64,
    status: String,
    #[serde(serialize_with = "crate::money::serialize")]
    opening_float: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    closing_float: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    expected_cash: Decimal,
    #[serde(serialize_with = "crate::money::serialize_some")]
    actual_cash: Option<Decimal>,
    #[serde(serialize_with = "crate::money::serialize_some")]
    cash_difference: Option<Decimal>,
    notes: Option<String>,
    #[serde(serialize_with = "crate::time::serialize")]
    opened_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize_some")]
    closed_at: Option<DateTime<Utc>>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<Shift> for ShiftResponse {
    fn from(s: Shift) -> Self {
        Self {
            id: s.id,
            branch_id: s.branch_id,
            cashier_id: s.cashier_id,
            status: s.status,
            opening_float: s.opening_float,
            closing_float: s.closing_float,
            expected_cash: s.expected_cash,
            actual_cash: s.actual_cash,
            cash_difference: s.cash_difference,
            notes: s.notes,
            opened_at: s.opened_at,
            closed_at: s.closed_at,
            created_at: s.created_at,
            updated_at: s.updated_at,
        }
    }
}

/// Go's message is "Shift not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Shift not found")
}

async fn open(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<OpenRequest>,
) -> AppResult<Reply<ShiftResponse>> {
    let branch_id = req.branch_id.unwrap_or(0);
    Validator::new()
        .int("branch_id", branch_id, &[Rule::Required])
        .decimal("opening_float", req.opening_float, &[Rule::Min(0)])
        .finish()?;
    crate::branches::check_owned(&state.db, user.company_id, branch_id).await?;
    if db::open_shift(&state.db, user.company_id, branch_id, user.user_id)
        .await?
        .is_some()
    {
        return Err(AppError::bad_request(
            "There is already an open shift for this cashier at this branch",
        ));
    }
    let shift = db::create(
        &state.db,
        user.company_id,
        branch_id,
        user.user_id,
        req.opening_float,
        Utc::now(),
    )
    .await?;
    Ok(success(
        StatusCode::CREATED,
        "Shift opened successfully",
        shift.into(),
    ))
}

async fn close(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<CloseRequest>,
) -> AppResult<Reply<ShiftResponse>> {
    let id = parse_id(&id, "shift")?;
    Validator::new()
        .decimal("actual_cash", req.actual_cash, &[Rule::Min(0)])
        .finish()?;

    let mut shift = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if shift.status != "open" {
        return Err(AppError::bad_request("Shift is not open"));
    }
    let now = Utc::now();
    let cash = orders::cash_total(&state.db, shift.branch_id, shift.opened_at, now).await?;
    let expected = shift.opening_float + cash;
    shift.status = "closed".to_owned();
    shift.closing_float = req.actual_cash;
    shift.expected_cash = expected;
    shift.actual_cash = Some(req.actual_cash);
    shift.cash_difference = Some(req.actual_cash - expected);
    shift.notes = req.notes;
    shift.closed_at = Some(now);
    db::close(&state.db, &shift).await?;

    let shift = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Shift closed successfully",
        shift.into(),
    ))
}

async fn current(State(state): State<AppState>, user: AuthUser) -> AppResult<Reply<ShiftResponse>> {
    let shift = db::open_shift(
        &state.db,
        user.company_id,
        user.branch_id.unwrap_or(1),
        user.user_id,
    )
    .await?
    .ok_or_else(|| AppError::not_found("No open shift found"))?;
    Ok(success(
        StatusCode::OK,
        "Current shift retrieved successfully",
        shift.into(),
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<ShiftResponse>> {
    let id = parse_id(&id, "shift")?;
    let shift = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Shift retrieved successfully",
        shift.into(),
    ))
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<ShiftResponse>>> {
    let (page, per_page) = query.page();
    let filter = ListFilter {
        branch_id: query.int("branch_id"),
        cashier_id: query.int("cashier_id"),
        status: query.get("status"),
        date_from: query.get("date_from"),
        date_to: query.get("date_to"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list(&state.db, user.company_id, &filter).await?;
    Ok(success_with_pagination(
        StatusCode::OK,
        "Shifts retrieved",
        rows.into_iter().map(ShiftResponse::from).collect(),
        Pagination::new(total, page, per_page, 0),
        None,
    ))
}
