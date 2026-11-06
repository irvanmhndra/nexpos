//! `/api/v1/stock-opnames`: same requests, responses, and errors as the Go
//! `StockOpnameHandler` and `StockOpnameService`.

use axum::{
    Router,
    extract::State,
    http::StatusCode,
    routing::{get, patch, post},
};
use chrono::{DateTime, Local, Utc};
use nexpos_db::{
    stock::{self, NewMovement},
    stock_opnames::{self as db, StockOpname, StockOpnameItem},
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use sqlx::PgPool;

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, RawItemIds, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/stock-opnames", get(list).post(create))
        .route("/api/v1/stock-opnames/{id}", get(show))
        .route("/api/v1/stock-opnames/{id}/items", patch(bulk_update_items))
        .route(
            "/api/v1/stock-opnames/{id}/items/{itemId}",
            patch(update_item),
        )
        .route("/api/v1/stock-opnames/{id}/complete", post(complete))
        .route("/api/v1/stock-opnames/{id}/cancel", post(cancel))
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CreateRequest {
    branch_id: Option<i64>,
    notes: Option<String>,
    category_id: Option<i64>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct UpdateItemRequest {
    counted_stock: Option<i64>,
    notes: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct BulkItem {
    item_id: Option<i64>,
    counted_stock: Option<i64>,
    notes: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct BulkRequest {
    items: Option<Vec<BulkItem>>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CompleteRequest {
    notes: Option<String>,
}

#[derive(Serialize)]
struct ItemResponse {
    id: i64,
    stock_opname_id: i64,
    product_variant_id: i64,
    sku: String,
    product_name: String,
    variant_name: String,
    system_stock: i32,
    counted_stock: Option<i32>,
    variance_qty: i32,
    #[serde(serialize_with = "crate::money::serialize")]
    unit_cost: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    variance_value: Decimal,
    notes: Option<String>,
    #[serde(serialize_with = "crate::time::serialize_some")]
    counted_at: Option<DateTime<Utc>>,
}

impl From<StockOpnameItem> for ItemResponse {
    fn from(i: StockOpnameItem) -> Self {
        Self {
            id: i.id,
            stock_opname_id: i.stock_opname_id,
            product_variant_id: i.product_variant_id,
            sku: i.sku,
            product_name: i.product_name,
            variant_name: i.variant_name,
            system_stock: i.system_stock,
            counted_stock: i.counted_stock,
            variance_qty: i.variance_qty,
            unit_cost: i.unit_cost,
            variance_value: i.variance_value,
            notes: i.notes,
            counted_at: i.counted_at,
        }
    }
}

#[derive(Serialize)]
struct OpnameResponse {
    id: i64,
    branch_id: i64,
    opname_number: String,
    status: String,
    notes: Option<String>,
    total_items: i64,
    counted_items: i64,
    total_variance_qty: i32,
    #[serde(serialize_with = "crate::money::serialize")]
    total_variance_value: Decimal,
    started_by: Option<i64>,
    completed_by: Option<i64>,
    #[serde(serialize_with = "crate::time::serialize")]
    started_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize_some")]
    completed_at: Option<DateTime<Utc>>,
    #[serde(serialize_with = "crate::time::serialize_some")]
    cancelled_at: Option<DateTime<Utc>>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
    #[serde(skip_serializing_if = "Option::is_none")]
    items: Option<Vec<ItemResponse>>,
}

fn response(o: StockOpname, stats: (i64, i64), items: Option<Vec<ItemResponse>>) -> OpnameResponse {
    OpnameResponse {
        id: o.id,
        branch_id: o.branch_id,
        opname_number: o.opname_number,
        status: o.status,
        notes: o.notes,
        total_items: stats.0,
        counted_items: stats.1,
        total_variance_qty: o.total_variance_qty,
        total_variance_value: o.total_variance_value,
        started_by: o.started_by,
        completed_by: o.completed_by,
        started_at: o.started_at,
        completed_at: o.completed_at,
        cancelled_at: o.cancelled_at,
        created_at: o.created_at,
        updated_at: o.updated_at,
        items,
    }
}

/// Go's message is "Stock opname not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Stock opname not found")
}

fn not_editable() -> AppError {
    AppError::bad_request("Cannot edit items on a completed or cancelled opname")
}

async fn load(db: &PgPool, company_id: i64, id: i64) -> AppResult<OpnameResponse> {
    let opname = db::get(db, company_id, id).await?.ok_or_else(not_found)?;
    let items = db::items_of(db, opname.id).await?;
    let stats = db::item_stats(db, &[opname.id]).await?;
    Ok(response(
        opname,
        stats.get(&id).copied().unwrap_or_default(),
        Some(items.into_iter().map(ItemResponse::from).collect()),
    ))
}

fn apply_count(
    item: &mut StockOpnameItem,
    counted: i64,
    notes: Option<String>,
    now: DateTime<Utc>,
) {
    let variance = counted - i64::from(item.system_stock);
    item.counted_stock = Some(i32::try_from(counted).unwrap_or(i32::MAX));
    item.variance_qty = i32::try_from(variance).unwrap_or(i32::MAX);
    item.variance_value = item.unit_cost * Decimal::from(variance);
    item.notes = notes;
    item.counted_at = Some(now);
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<CreateRequest>,
) -> AppResult<Reply<OpnameResponse>> {
    let branch_id = req.branch_id.unwrap_or(0);
    Validator::new()
        .int("branch_id", branch_id, &[Rule::Required])
        .finish()?;
    crate::branches::check_owned(&state.db, user.company_id, branch_id).await?;

    let date = Local::now().format("%Y%m%d").to_string();
    let number = db::next_number(&state.db, user.company_id, &date).await?;

    // In one transaction, so a refused snapshot leaves no empty opname behind.
    let mut tx = state.db.begin().await?;
    let id = db::create(
        &mut *tx,
        user.company_id,
        branch_id,
        &number,
        req.notes.as_deref(),
        user.user_id,
        Utc::now(),
    )
    .await?;
    let taken =
        db::snapshot_items(&mut *tx, id, user.company_id, branch_id, req.category_id).await?;
    if taken == 0 {
        return Err(AppError::bad_request(
            "No active products found to snapshot for this branch",
        ));
    }
    tx.commit().await?;

    let opname = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::CREATED,
        "Stock opname created successfully",
        opname,
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<OpnameResponse>> {
    let id = parse_id(&id, "stock opname")?;
    let opname = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Stock opname retrieved successfully",
        opname,
    ))
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<OpnameResponse>>> {
    let (page, per_page) = query.page();
    let (rows, total) = db::list(
        &state.db,
        user.company_id,
        query.get("status"),
        query.int("branch_id"),
        per_page,
        (page - 1) * per_page,
    )
    .await?;
    let ids: Vec<i64> = rows.iter().map(|o| o.id).collect();
    let stats = db::item_stats(&state.db, &ids).await?;
    let data = rows
        .into_iter()
        .map(|o| {
            let counts = stats.get(&o.id).copied().unwrap_or_default();
            response(o, counts, None)
        })
        .collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Stock opnames retrieved",
        data,
        Pagination::new(total, page, per_page, 0),
        None,
    ))
}

async fn update_item(
    State(state): State<AppState>,
    user: AuthUser,
    RawItemIds(id, item_id): RawItemIds,
    JsonBody(req): JsonBody<UpdateItemRequest>,
) -> AppResult<Reply<ItemResponse>> {
    let id = parse_id(&id, "stock opname")?;
    let item_id = parse_id(&item_id, "item")?;
    let counted = req.counted_stock.unwrap_or(0);
    Validator::new()
        .int("counted_stock", counted, &[Rule::Min(0)])
        .finish()?;

    let opname = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if opname.status != "in_progress" {
        return Err(not_editable());
    }
    let mut item = db::item(&state.db, id, item_id)
        .await?
        .ok_or_else(|| AppError::not_found("Stock opname item not found"))?;
    apply_count(&mut item, counted, req.notes, Utc::now());
    db::update_item(&state.db, &item, user.user_id).await?;
    Ok(success(
        StatusCode::OK,
        "Stock opname item updated successfully",
        item.into(),
    ))
}

async fn bulk_update_items(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<BulkRequest>,
) -> AppResult<Reply<OpnameResponse>> {
    let id = parse_id(&id, "stock opname")?;
    let mut v = Validator::new();
    v.list("items", req.items.as_ref().map(Vec::len));
    let lines = req.items.unwrap_or_default();
    for line in &lines {
        v.int("item_id", line.item_id.unwrap_or(0), &[Rule::Required])
            .int(
                "counted_stock",
                line.counted_stock.unwrap_or(0),
                &[Rule::Min(0)],
            );
    }
    v.finish()?;

    let opname = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if opname.status != "in_progress" {
        return Err(not_editable());
    }
    let now = Utc::now();
    let mut tx = state.db.begin().await?;
    for line in lines {
        let mut item = db::item(&mut *tx, id, line.item_id.unwrap_or(0))
            .await?
            .ok_or_else(|| AppError::bad_request("Stock opname item not found"))?;
        apply_count(&mut item, line.counted_stock.unwrap_or(0), line.notes, now);
        db::update_item(&mut *tx, &item, user.user_id).await?;
    }
    tx.commit().await?;

    let opname = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Stock opname items updated successfully",
        opname,
    ))
}

async fn complete(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<CompleteRequest>,
) -> AppResult<Reply<OpnameResponse>> {
    let id = parse_id(&id, "stock opname")?;

    // The opname and each stock row are locked: completing twice must not apply
    // the count twice, and a sale during completion must not be lost.
    let mut tx = state.db.begin().await?;
    let mut opname = db::lock(&mut *tx, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if opname.status != "in_progress" {
        return Err(AppError::bad_request("Stock opname is not in progress"));
    }

    let mut variance_qty = 0;
    let mut variance_value = Decimal::ZERO;
    for item in db::items_of(&mut *tx, opname.id).await? {
        let Some(counted) = item.counted_stock else {
            continue;
        };
        let current = stock::lock(&mut tx, item.product_variant_id, opname.branch_id)
            .await?
            .quantity;
        if counted != current {
            stock::record_movement(
                &mut *tx,
                &NewMovement {
                    variant_id: item.product_variant_id,
                    branch_id: opname.branch_id,
                    kind: "ADJUST",
                    quantity: (i64::from(counted) - i64::from(current)).abs(),
                    stock_before: i64::from(current),
                    stock_after: i64::from(counted),
                    unit_cost: Some(item.unit_cost),
                    reference_type: Some("stock_opname"),
                    reference_id: Some(opname.id),
                    note: None,
                    created_by: Some(user.user_id),
                },
            )
            .await?;
            stock::set_quantity(
                &mut *tx,
                item.product_variant_id,
                opname.branch_id,
                i64::from(counted),
            )
            .await?;
        }
        variance_qty += item.variance_qty;
        variance_value += item.variance_value;
    }

    opname.status = "completed".to_owned();
    opname.completed_by = Some(user.user_id);
    opname.completed_at = Some(Utc::now());
    opname.total_variance_qty = variance_qty;
    opname.total_variance_value = variance_value;
    if req.notes.is_some() {
        opname.notes = req.notes;
    }
    db::update(&mut *tx, &opname).await?;
    tx.commit().await?;

    let opname = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Stock opname completed successfully",
        opname,
    ))
}

async fn cancel(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<OpnameResponse>> {
    let id = parse_id(&id, "stock opname")?;
    let mut opname = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if opname.status != "in_progress" {
        return Err(AppError::bad_request(
            "Only in-progress opnames can be cancelled",
        ));
    }
    opname.status = "cancelled".to_owned();
    opname.cancelled_at = Some(Utc::now());
    db::update(&state.db, &opname).await?;

    let opname = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Stock opname cancelled successfully",
        opname,
    ))
}
