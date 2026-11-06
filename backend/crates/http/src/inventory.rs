//! `/api/v1/inventory` and `/api/v1/inventory/movements`: same requests,
//! responses, and errors as the Go `InventoryHandler` and `InventoryService`.

use axum::{
    Router,
    extract::State,
    http::StatusCode,
    routing::{get, post, put},
};
use chrono::{DateTime, Datelike, Local, TimeZone, Utc};
use nexpos_db::{
    products,
    stock::{self as db, InventoryFilter, InventoryRow, MovementFilter, MovementRow, NewMovement},
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/inventory", get(list))
        .route("/api/v1/inventory/stats", get(stats))
        .route("/api/v1/inventory/adjust", post(adjust))
        .route(
            "/api/v1/inventory/{variantId}/min-stock",
            put(update_min_stock),
        )
        .route("/api/v1/inventory/movements", get(movements))
        .route("/api/v1/inventory/movements/stats", get(movement_stats))
}

fn variant_not_found() -> AppError {
    AppError::not_found("Product variant not found")
}

#[derive(Serialize)]
struct InventoryItem {
    product_variant_id: i64,
    product_id: i64,
    product_name: String,
    variant_name: String,
    sku: String,
    category_name: String,
    image_data: Option<String>,
    branch_id: i64,
    branch_name: String,
    current_stock: i32,
    min_stock: i32,
    #[serde(serialize_with = "crate::money::serialize")]
    stock_value: Decimal,
    status: &'static str,
    has_variants: bool,
}

fn stock_status(current: i32, min: i32) -> &'static str {
    if current <= 0 {
        "out_of_stock"
    } else if current <= min {
        "low"
    } else {
        "normal"
    }
}

impl From<InventoryRow> for InventoryItem {
    fn from(r: InventoryRow) -> Self {
        Self {
            status: stock_status(r.current_stock, r.min_stock),
            product_variant_id: r.product_variant_id,
            product_id: r.product_id,
            product_name: r.product_name,
            variant_name: r.variant_name,
            sku: r.sku,
            category_name: r.category_name,
            image_data: r.image_data,
            branch_id: r.branch_id,
            branch_name: r.branch_name,
            current_stock: r.current_stock,
            min_stock: r.min_stock,
            stock_value: r.stock_value,
            has_variants: r.has_variants,
        }
    }
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<InventoryItem>>> {
    let (page, per_page) = query.page();
    let filter = InventoryFilter {
        branch_id: query.int("branch_id").unwrap_or(0),
        search: query.get("search"),
        category: query.get("category"),
        status: query.get("status"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list_inventory(&state.db, user.company_id, &filter).await?;
    Ok(success_with_pagination(
        StatusCode::OK,
        "Inventory list retrieved",
        rows.into_iter().map(InventoryItem::from).collect(),
        Pagination::new(total, page, per_page, 0),
        None,
    ))
}

#[derive(Serialize)]
struct StatsResponse {
    total_sku: i64,
    low_stock: i64,
    out_of_stock: i64,
    #[serde(serialize_with = "crate::money::serialize")]
    total_stock_value: Decimal,
}

async fn stats(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<StatsResponse>> {
    let branch_id = query.int("branch_id").unwrap_or(0);
    let s = db::inventory_stats(&state.db, user.company_id, branch_id).await?;
    Ok(success(
        StatusCode::OK,
        "Inventory stats retrieved",
        StatsResponse {
            total_sku: s.total_sku,
            low_stock: s.low_stock,
            out_of_stock: s.out_of_stock,
            total_stock_value: s.total_stock_value,
        },
    ))
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct AdjustRequest {
    variant_id: Option<i64>,
    branch_id: Option<i64>,
    #[serde(rename = "type")]
    kind: Option<String>,
    quantity: Option<i64>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    unit_cost: Option<Decimal>,
    note: Option<String>,
}

async fn adjust(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<AdjustRequest>,
) -> AppResult<Reply<()>> {
    let variant_id = req.variant_id.unwrap_or(0);
    let branch_id = req.branch_id.unwrap_or(0);
    let kind = req.kind.as_deref().unwrap_or_default();
    let quantity = req.quantity.unwrap_or(0);
    Validator::new()
        .int("variant_id", variant_id, &[Rule::Required])
        .int("branch_id", branch_id, &[Rule::Required])
        .field(
            "type",
            kind,
            &[Rule::Required, Rule::OneOf("IN OUT ADJUST")],
        )
        .int("quantity", quantity, &[Rule::Required, Rule::Min(1)])
        .finish()?;

    if products::variant_for_company(&state.db, user.company_id, variant_id)
        .await?
        .is_none()
    {
        return Err(variant_not_found());
    }
    crate::branches::check_owned(&state.db, user.company_id, branch_id).await?;

    // The new quantity is computed from the locked row, so a concurrent change
    // waits rather than being overwritten.
    let mut tx = state.db.begin().await?;
    let current = i64::from(db::lock(&mut tx, variant_id, branch_id).await?.quantity);
    let after = match kind {
        "IN" => current + quantity,
        "OUT" if current < quantity => return Err(AppError::bad_request("Insufficient stock")),
        "OUT" => current - quantity,
        _ => quantity,
    };
    db::record_movement(
        &mut *tx,
        &NewMovement {
            variant_id,
            branch_id,
            kind,
            quantity,
            stock_before: current,
            stock_after: after,
            unit_cost: req.unit_cost,
            reference_type: None,
            reference_id: None,
            note: req.note.as_deref(),
            created_by: Some(user.user_id),
        },
    )
    .await?;
    db::set_quantity(&mut *tx, variant_id, branch_id, after).await?;
    tx.commit().await?;

    Ok(success_message(
        StatusCode::OK,
        "Stock adjusted successfully",
    ))
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct MinStockRequest {
    min_quantity: Option<i64>,
}

async fn update_min_stock(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(variant): RawId,
    query: QueryParams,
    JsonBody(req): JsonBody<MinStockRequest>,
) -> AppResult<Reply<()>> {
    let variant_id = parse_id(&variant, "variant")?;
    let branch_id = query.int("branch_id").unwrap_or(0);
    if branch_id == 0 {
        return Err(AppError::bad_request("branch_id is required"));
    }
    let min_quantity = req.min_quantity.unwrap_or(0);
    Validator::new()
        .int("min_quantity", min_quantity, &[Rule::Min(0)])
        .finish()?;

    if products::variant_for_company(&state.db, user.company_id, variant_id)
        .await?
        .is_none()
    {
        return Err(variant_not_found());
    }
    crate::branches::check_owned(&state.db, user.company_id, branch_id).await?;
    db::set_min_quantity(&state.db, variant_id, branch_id, min_quantity).await?;
    Ok(success_message(
        StatusCode::OK,
        "Min stock updated successfully",
    ))
}

#[derive(Serialize)]
struct MovementResponse {
    id: i64,
    product_variant_id: i64,
    product_name: String,
    variant_name: String,
    sku: String,
    branch_id: i64,
    branch_name: String,
    #[serde(rename = "type")]
    kind: String,
    quantity: i32,
    stock_before: i32,
    stock_after: i32,
    #[serde(serialize_with = "crate::money::serialize_some")]
    unit_cost: Option<Decimal>,
    reference_type: Option<String>,
    reference_id: Option<i64>,
    note: Option<String>,
    created_by: Option<i64>,
    created_by_name: String,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
}

impl From<MovementRow> for MovementResponse {
    fn from(r: MovementRow) -> Self {
        Self {
            id: r.id,
            product_variant_id: r.product_variant_id,
            product_name: r.product_name,
            variant_name: r.variant_name,
            sku: r.sku,
            branch_id: r.branch_id,
            branch_name: r.branch_name,
            kind: r.kind,
            quantity: r.quantity,
            stock_before: r.stock_before,
            stock_after: r.stock_after,
            unit_cost: r.unit_cost,
            reference_type: r.reference_type,
            reference_id: r.reference_id,
            note: r.note,
            created_by: r.created_by,
            created_by_name: r.created_by_name,
            created_at: r.created_at,
        }
    }
}

async fn movements(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<MovementResponse>>> {
    let (page, per_page) = query.page();
    let filter = MovementFilter {
        branch_id: query.int("branch_id").unwrap_or(0),
        kind: query.get("type"),
        search: query.get("search"),
        start_date: query.get("start_date"),
        end_date: query.get("end_date"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list_movements(&state.db, user.company_id, &filter).await?;
    Ok(success_with_pagination(
        StatusCode::OK,
        "Stock movements retrieved",
        rows.into_iter().map(MovementResponse::from).collect(),
        Pagination::new(total, page, per_page, 0),
        None,
    ))
}

#[derive(Serialize)]
struct MovementStatsResponse {
    total_movements: i64,
    stock_in: i64,
    stock_out: i64,
    adjustments: i64,
}

async fn movement_stats(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<MovementStatsResponse>> {
    let branch_id = query.int("branch_id").unwrap_or(0);
    let now = Local::now();
    let since = Utc
        .with_ymd_and_hms(now.year(), now.month(), 1, 0, 0, 0)
        .single()
        .unwrap_or_default();
    let s = db::movement_stats(&state.db, user.company_id, branch_id, since).await?;
    Ok(success(
        StatusCode::OK,
        "Movement stats retrieved",
        MovementStatsResponse {
            total_movements: s.total_movements,
            stock_in: s.stock_in,
            stock_out: s.stock_out,
            adjustments: s.adjustments,
        },
    ))
}
