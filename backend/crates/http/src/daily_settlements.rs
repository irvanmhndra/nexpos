//! `/api/v1/daily-settlements`: same requests, responses, and errors as the Go
//! `DailySettlementHandler` and `DailySettlementService`.
//!
//! Go behavior kept for the web app: `settlement_date` reads as a midnight UTC
//! time on a settlement but echoes the query as sent on the report.

use std::collections::HashMap;

use axum::{
    Router,
    extract::State,
    http::StatusCode,
    routing::{get, patch, post},
};
use chrono::{DateTime, NaiveDate, Utc};
use nexpos_db::daily_settlements::{
    self as db, DailySettlement, DailySettlementItem, ListFilter, NewSettlement,
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use sqlx::PgPool;

use crate::{
    AppError, AppResult, AppState,
    branches::check_owned,
    envelope::{Pagination, Reply, success, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, RawItemIds, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/daily-settlements", get(list).post(create))
        .route("/api/v1/daily-settlements/report", get(report))
        .route("/api/v1/daily-settlements/{id}", get(show))
        .route(
            "/api/v1/daily-settlements/{id}/items",
            patch(bulk_update_items),
        )
        .route(
            "/api/v1/daily-settlements/{id}/items/{itemId}",
            patch(update_item),
        )
        .route("/api/v1/daily-settlements/{id}/finalize", post(finalize))
}

/// Always shown, even with no activity that day, so the screen stays consistent.
const METHODS: [&str; 6] = [
    "cash",
    "debit_card",
    "credit_card",
    "e_wallet",
    "bank_transfer",
    "qris",
];

#[derive(Default, Deserialize)]
#[serde(default)]
struct CreateRequest {
    branch_id: Option<i64>,
    settlement_date: Option<String>,
    notes: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct UpdateItemRequest {
    #[serde(deserialize_with = "crate::money::deserialize")]
    actual_amount: Decimal,
    notes: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct BulkItem {
    item_id: Option<i64>,
    #[serde(deserialize_with = "crate::money::deserialize")]
    actual_amount: Decimal,
    notes: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct BulkRequest {
    items: Option<Vec<BulkItem>>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct FinalizeRequest {
    notes: Option<String>,
}

#[derive(Serialize)]
struct ItemResponse {
    id: i64,
    payment_method: String,
    #[serde(serialize_with = "crate::money::serialize")]
    expected_amount: Decimal,
    #[serde(serialize_with = "crate::money::serialize_some")]
    actual_amount: Option<Decimal>,
    #[serde(serialize_with = "crate::money::serialize")]
    variance_amount: Decimal,
    notes: Option<String>,
}

impl From<DailySettlementItem> for ItemResponse {
    fn from(i: DailySettlementItem) -> Self {
        Self {
            id: i.id,
            payment_method: i.payment_method,
            expected_amount: i.expected_amount,
            actual_amount: i.actual_amount,
            variance_amount: i.variance_amount,
            notes: i.notes,
        }
    }
}

#[derive(Serialize)]
struct SettlementResponse {
    id: i64,
    branch_id: i64,
    settlement_date: String,
    status: String,
    #[serde(serialize_with = "crate::money::serialize")]
    total_sales: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_refunds: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_expenses: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_expected: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_actual: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_variance: Decimal,
    notes: Option<String>,
    recorded_by: Option<i64>,
    finalized_by: Option<i64>,
    #[serde(serialize_with = "crate::time::serialize_some")]
    finalized_at: Option<DateTime<Utc>>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
    items: Vec<ItemResponse>,
}

fn response(s: DailySettlement, items: Vec<DailySettlementItem>) -> SettlementResponse {
    let mut total_expected = Decimal::ZERO;
    let mut total_actual = Decimal::ZERO;
    let mut total_variance = Decimal::ZERO;
    for item in &items {
        total_expected += item.expected_amount;
        total_actual += item.actual_amount.unwrap_or_default();
        total_variance += item.variance_amount;
    }
    SettlementResponse {
        id: s.id,
        branch_id: s.branch_id,
        settlement_date: s.settlement_date.format("%Y-%m-%dT00:00:00Z").to_string(),
        status: s.status,
        total_sales: s.total_sales,
        total_refunds: s.total_refunds,
        total_expenses: s.total_expenses,
        total_expected,
        total_actual,
        total_variance,
        notes: s.notes,
        recorded_by: s.recorded_by,
        finalized_by: s.finalized_by,
        finalized_at: s.finalized_at,
        created_at: s.created_at,
        updated_at: s.updated_at,
        items: items.into_iter().map(ItemResponse::from).collect(),
    }
}

/// Go's message is "Daily settlement not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Daily settlement not found")
}

fn finalized() -> AppError {
    AppError::bad_request("Cannot edit a finalized settlement")
}

async fn load(db: &PgPool, company_id: i64, id: i64) -> AppResult<SettlementResponse> {
    let settlement = db::get(db, company_id, id).await?.ok_or_else(not_found)?;
    let mut items = db::items_by_settlement(db, &[settlement.id]).await?;
    Ok(response(settlement, items.remove(&id).unwrap_or_default()))
}

// ============== Expected amounts ==============

#[derive(Serialize)]
struct MethodBreakdown {
    payment_method: &'static str,
    #[serde(serialize_with = "crate::money::serialize")]
    gross_sales: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    refunds: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    expenses_out: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    expected_amount: Decimal,
}

struct Breakdown {
    by_method: Vec<MethodBreakdown>,
    total_sales: Decimal,
    total_refunds: Decimal,
    total_expenses: Decimal,
    total_expected: Decimal,
}

/// Expenses are paid from the till, so they only reduce what's expected in cash.
async fn breakdown(
    db: &PgPool,
    company_id: i64,
    branch_id: i64,
    date: &str,
) -> AppResult<Breakdown> {
    let totals = db::payment_breakdown(db, company_id, branch_id, date).await?;
    let expenses = db::expenses_total(db, company_id, branch_id, date).await?;
    let totals: HashMap<String, (Decimal, Decimal)> = totals
        .into_iter()
        .map(|t| (t.method, (t.gross_sales, t.refunds)))
        .collect();

    let mut result = Breakdown {
        by_method: Vec::with_capacity(METHODS.len()),
        total_sales: Decimal::ZERO,
        total_refunds: Decimal::ZERO,
        total_expenses: expenses,
        total_expected: Decimal::ZERO,
    };
    for method in METHODS {
        let (gross, refunds) = totals.get(method).copied().unwrap_or_default();
        let expenses_out = if method == "cash" {
            expenses
        } else {
            Decimal::ZERO
        };
        let expected = gross - refunds - expenses_out;
        result.by_method.push(MethodBreakdown {
            payment_method: method,
            gross_sales: gross,
            refunds,
            expenses_out,
            expected_amount: expected,
        });
        result.total_sales += gross;
        result.total_refunds += refunds;
        result.total_expected += expected;
    }
    Ok(result)
}

#[derive(Serialize)]
struct ReportResponse {
    branch_id: i64,
    settlement_date: String,
    #[serde(serialize_with = "crate::money::serialize")]
    total_sales: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_refunds: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_expenses: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_expected: Decimal,
    by_method: Vec<MethodBreakdown>,
}

async fn report(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<ReportResponse>> {
    let branch_id = query.int("branch_id").unwrap_or(0);
    if branch_id == 0 {
        return Err(AppError::bad_request("branch_id is required"));
    }
    let date = query.get("date");
    if date.is_empty() {
        return Err(AppError::bad_request("date is required"));
    }
    check_owned(&state.db, user.company_id, branch_id).await?;
    let b = breakdown(&state.db, user.company_id, branch_id, date).await?;
    Ok(success(
        StatusCode::OK,
        "Daily settlement report retrieved",
        ReportResponse {
            branch_id,
            settlement_date: date.to_owned(),
            total_sales: b.total_sales,
            total_refunds: b.total_refunds,
            total_expenses: b.total_expenses,
            total_expected: b.total_expected,
            by_method: b.by_method,
        },
    ))
}

/// `time.Parse("2006-01-02", …)`: exactly this format, and a real date.
fn is_iso_date(text: &str) -> bool {
    text.len() == 10 && NaiveDate::parse_from_str(text, "%Y-%m-%d").is_ok()
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<CreateRequest>,
) -> AppResult<Reply<SettlementResponse>> {
    let branch_id = req.branch_id.unwrap_or(0);
    let date = req.settlement_date.as_deref().unwrap_or_default();
    Validator::new()
        .int("branch_id", branch_id, &[Rule::Required])
        .field("settlement_date", date, &[Rule::Required])
        .finish()?;
    check_owned(&state.db, user.company_id, branch_id).await?;
    if !is_iso_date(date) {
        return Err(AppError::bad_request("settlement_date must be YYYY-MM-DD"));
    }
    if db::exists_for_date(&state.db, user.company_id, branch_id, date).await? {
        return Err(AppError::bad_request(
            "Settlement already exists for this branch and date",
        ));
    }

    let b = breakdown(&state.db, user.company_id, branch_id, date).await?;
    let mut tx = state.db.begin().await?;
    let id = db::create(
        &mut *tx,
        &NewSettlement {
            company_id: user.company_id,
            branch_id,
            settlement_date: date,
            total_sales: b.total_sales,
            total_refunds: b.total_refunds,
            total_expenses: b.total_expenses,
            notes: req.notes.as_deref(),
            recorded_by: user.user_id,
        },
    )
    .await?;
    for m in &b.by_method {
        // Nothing has been counted yet, so everything expected is missing.
        db::create_item(
            &mut *tx,
            id,
            m.payment_method,
            m.expected_amount,
            -m.expected_amount,
        )
        .await?;
    }
    tx.commit().await?;

    let settlement = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::CREATED,
        "Daily settlement created successfully",
        settlement,
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<SettlementResponse>> {
    let id = parse_id(&id, "settlement")?;
    let settlement = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Daily settlement retrieved successfully",
        settlement,
    ))
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<SettlementResponse>>> {
    let (page, per_page) = query.page();
    let filter = ListFilter {
        status: query.get("status"),
        branch_id: query.int("branch_id"),
        date_from: query.get("date_from"),
        date_to: query.get("date_to"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list(&state.db, user.company_id, &filter).await?;
    let ids: Vec<i64> = rows.iter().map(|s| s.id).collect();
    let mut items = db::items_by_settlement(&state.db, &ids).await?;
    let data = rows
        .into_iter()
        .map(|s| {
            let lines = items.remove(&s.id).unwrap_or_default();
            response(s, lines)
        })
        .collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Daily settlements retrieved",
        data,
        Pagination::new(total, page, per_page, 0),
        None,
    ))
}

fn apply_actual(item: &mut DailySettlementItem, actual: Decimal, notes: Option<String>) {
    item.actual_amount = Some(actual);
    item.variance_amount = actual - item.expected_amount;
    item.notes = notes;
}

async fn editable(db: &PgPool, company_id: i64, id: i64) -> AppResult<()> {
    let settlement = db::get(db, company_id, id).await?.ok_or_else(not_found)?;
    if settlement.status != "draft" {
        return Err(finalized());
    }
    Ok(())
}

async fn update_item(
    State(state): State<AppState>,
    user: AuthUser,
    RawItemIds(id, item_id): RawItemIds,
    JsonBody(req): JsonBody<UpdateItemRequest>,
) -> AppResult<Reply<ItemResponse>> {
    let id = parse_id(&id, "settlement")?;
    let item_id = parse_id(&item_id, "item")?;
    Validator::new()
        .decimal("actual_amount", req.actual_amount, &[Rule::NotNegative])
        .finish()?;
    editable(&state.db, user.company_id, id).await?;

    let mut item = db::item(&state.db, id, item_id)
        .await?
        .ok_or_else(|| AppError::not_found("Settlement item not found"))?;
    apply_actual(&mut item, req.actual_amount, req.notes);
    db::update_item(&state.db, &item).await?;
    Ok(success(
        StatusCode::OK,
        "Settlement item updated successfully",
        item.into(),
    ))
}

async fn bulk_update_items(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<BulkRequest>,
) -> AppResult<Reply<SettlementResponse>> {
    let id = parse_id(&id, "settlement")?;
    let mut v = Validator::new();
    v.list("items", req.items.as_ref().map(Vec::len));
    let lines = req.items.unwrap_or_default();
    for line in &lines {
        v.int("item_id", line.item_id.unwrap_or(0), &[Rule::Required])
            .decimal("actual_amount", line.actual_amount, &[Rule::NotNegative]);
    }
    v.finish()?;
    editable(&state.db, user.company_id, id).await?;

    let mut tx = state.db.begin().await?;
    for line in lines {
        let mut item = db::item(&mut *tx, id, line.item_id.unwrap_or(0))
            .await?
            .ok_or_else(|| AppError::bad_request("Settlement item not found"))?;
        apply_actual(&mut item, line.actual_amount, line.notes);
        db::update_item(&mut *tx, &item).await?;
    }
    tx.commit().await?;

    let settlement = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Settlement items updated successfully",
        settlement,
    ))
}

async fn finalize(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<FinalizeRequest>,
) -> AppResult<Reply<SettlementResponse>> {
    let id = parse_id(&id, "settlement")?;
    let mut settlement = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if settlement.status != "draft" {
        return Err(AppError::bad_request("Settlement is not in draft status"));
    }
    settlement.status = "finalized".to_owned();
    settlement.finalized_by = Some(user.user_id);
    settlement.finalized_at = Some(Utc::now());
    if req.notes.is_some() {
        settlement.notes = req.notes;
    }
    db::finalize(&state.db, &settlement).await?;

    let settlement = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Daily settlement finalized successfully",
        settlement,
    ))
}
