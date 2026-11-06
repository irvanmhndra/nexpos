//! `/api/v1/purchase-orders`: same requests, responses, and errors as the Go
//! `PurchaseOrderHandler` and `PurchaseOrderService`.
//!
//! Go behavior kept for the web app: the order goes to the session's branch, not
//! `branch_id` in the body; create echoes costs as sent, while reads show the
//! columns (rounded to 2 places); receive answers with the `updated_at` from
//! before the write.

use std::collections::HashMap;

use axum::{
    Router,
    extract::State,
    http::StatusCode,
    routing::{get, post},
};
use chrono::{DateTime, Local, Utc};
use nexpos_db::{
    products,
    purchase_orders::{self as db, NewItem, NewPurchaseOrder, PurchaseOrder, PurchaseOrderItem},
    stock::{self, NewMovement},
    suppliers,
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
        .route("/api/v1/purchase-orders", get(list).post(create))
        .route("/api/v1/purchase-orders/{id}", get(show))
        .route("/api/v1/purchase-orders/{id}/receive", post(receive))
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CreateItem {
    product_variant_id: Option<i64>,
    sku: Option<String>,
    variant_name: Option<String>,
    quantity: Option<i64>,
    #[serde(deserialize_with = "crate::money::deserialize")]
    unit_cost: Decimal,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CreateRequest {
    branch_id: Option<i64>,
    supplier_id: Option<i64>,
    notes: Option<String>,
    items: Option<Vec<CreateItem>>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct ReceiveItem {
    item_id: Option<i64>,
    received_quantity: Option<i64>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct ReceiveRequest {
    items: Option<Vec<ReceiveItem>>,
}

#[derive(Serialize)]
struct ItemResponse {
    id: i64,
    product_variant_id: Option<i64>,
    sku: String,
    variant_name: String,
    quantity: i32,
    received_quantity: i32,
    #[serde(serialize_with = "crate::money::serialize")]
    unit_cost: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    subtotal: Decimal,
}

impl From<PurchaseOrderItem> for ItemResponse {
    fn from(i: PurchaseOrderItem) -> Self {
        Self {
            id: i.id,
            product_variant_id: i.product_variant_id,
            sku: i.sku,
            variant_name: i.variant_name,
            quantity: i.quantity,
            received_quantity: i.received_quantity,
            unit_cost: i.unit_cost,
            subtotal: i.subtotal,
        }
    }
}

#[derive(Serialize)]
struct PurchaseOrderResponse {
    id: i64,
    branch_id: i64,
    supplier_id: i64,
    po_number: String,
    status: String,
    notes: Option<String>,
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
    #[serde(serialize_with = "crate::time::serialize_some")]
    ordered_at: Option<DateTime<Utc>>,
    #[serde(serialize_with = "crate::time::serialize_some")]
    received_at: Option<DateTime<Utc>>,
    #[serde(serialize_with = "crate::time::serialize_some")]
    cancelled_at: Option<DateTime<Utc>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    items: Option<Vec<ItemResponse>>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

fn response(po: PurchaseOrder, items: Option<Vec<ItemResponse>>) -> PurchaseOrderResponse {
    PurchaseOrderResponse {
        id: po.id,
        branch_id: po.branch_id,
        supplier_id: po.supplier_id,
        po_number: po.po_number,
        status: po.status,
        notes: po.notes,
        total_amount: po.total_amount,
        ordered_at: po.ordered_at,
        received_at: po.received_at,
        cancelled_at: po.cancelled_at,
        items,
        created_at: po.created_at,
        updated_at: po.updated_at,
    }
}

/// Go's message is "Purchase order not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Purchase order not found")
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<PurchaseOrderResponse>>> {
    let (page, per_page) = query.page();
    let (rows, total) = db::list(
        &state.db,
        user.company_id,
        query.get("status"),
        query.int("supplier_id"),
        per_page,
        (page - 1) * per_page,
    )
    .await?;
    Ok(success_with_pagination(
        StatusCode::OK,
        "Purchase orders retrieved",
        rows.into_iter().map(|po| response(po, None)).collect(),
        Pagination::new(total, page, per_page, 0),
        None,
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<PurchaseOrderResponse>> {
    let id = parse_id(&id, "purchase order")?;
    let po = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    let items = db::items_of(&state.db, po.id).await?;
    Ok(success(
        StatusCode::OK,
        "Purchase order retrieved successfully",
        response(
            po,
            Some(items.into_iter().map(ItemResponse::from).collect()),
        ),
    ))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<CreateRequest>,
) -> AppResult<Reply<PurchaseOrderResponse>> {
    let supplier_id = req.supplier_id.unwrap_or(0);
    let mut v = Validator::new();
    v.int("branch_id", req.branch_id.unwrap_or(0), &[Rule::Required])
        .int("supplier_id", supplier_id, &[Rule::Required])
        .list("items", req.items.as_ref().map(Vec::len));
    let items = req.items.as_deref().unwrap_or_default();
    for item in items {
        v.field(
            "sku",
            item.sku.as_deref().unwrap_or_default(),
            &[Rule::Required, Rule::Max(100)],
        )
        .field(
            "variant_name",
            item.variant_name.as_deref().unwrap_or_default(),
            &[Rule::Required, Rule::Max(255)],
        )
        .int(
            "quantity",
            item.quantity.unwrap_or(0),
            &[Rule::Required, Rule::Min(1)],
        )
        .decimal("unit_cost", item.unit_cost, &[Rule::Min(0)]);
    }
    v.finish()?;

    if suppliers::get(&state.db, user.company_id, supplier_id)
        .await?
        .is_none()
    {
        return Err(AppError::not_found("Supplier not found"));
    }

    let date = Local::now().format("%Y%m%d").to_string();
    let number = db::next_number(&state.db, user.company_id, &date).await?;

    let mut total = Decimal::ZERO;
    let mut subtotals = Vec::with_capacity(items.len());
    for item in items {
        // Receiving stocks the variant and updates its cost, so it must be the company's.
        if let Some(variant) = item.product_variant_id
            && products::variant_for_company(&state.db, user.company_id, variant)
                .await?
                .is_none()
        {
            return Err(AppError::not_found("Product variant not found"));
        }
        let subtotal = item.unit_cost * Decimal::from(item.quantity.unwrap_or(0));
        total += subtotal;
        subtotals.push(subtotal);
    }

    let mut tx = state.db.begin().await?;
    let mut po = db::create(
        &mut *tx,
        &NewPurchaseOrder {
            company_id: user.company_id,
            branch_id: user.branch_id.unwrap_or(1),
            supplier_id,
            po_number: &number,
            notes: req.notes.as_deref(),
            total_amount: total,
            ordered_at: Utc::now(),
        },
    )
    .await?;
    let mut responses = Vec::with_capacity(items.len());
    for (item, subtotal) in items.iter().zip(subtotals) {
        let sku = item.sku.as_deref().unwrap_or_default();
        let variant_name = item.variant_name.as_deref().unwrap_or_default();
        let quantity = item.quantity.unwrap_or(0);
        let id = db::create_item(
            &mut *tx,
            &NewItem {
                purchase_order_id: po.id,
                product_variant_id: item.product_variant_id,
                sku,
                variant_name,
                quantity,
                unit_cost: item.unit_cost,
                subtotal,
            },
        )
        .await?;
        responses.push(ItemResponse {
            id,
            product_variant_id: item.product_variant_id,
            sku: sku.to_owned(),
            variant_name: variant_name.to_owned(),
            quantity: i32::try_from(quantity).unwrap_or_default(),
            received_quantity: 0,
            unit_cost: item.unit_cost,
            subtotal,
        });
    }
    tx.commit().await?;

    po.total_amount = total;
    Ok(success(
        StatusCode::CREATED,
        "Purchase order created successfully",
        response(po, Some(responses)),
    ))
}

async fn receive(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<ReceiveRequest>,
) -> AppResult<Reply<PurchaseOrderResponse>> {
    let id = parse_id(&id, "purchase order")?;
    let mut v = Validator::new();
    v.list("items", req.items.as_ref().map(Vec::len));
    let lines = req.items.as_deref().unwrap_or_default();
    for line in lines {
        v.int("item_id", line.item_id.unwrap_or(0), &[Rule::Required])
            .int(
                "received_quantity",
                line.received_quantity.unwrap_or(0),
                &[Rule::Min(0)],
            );
    }
    v.finish()?;

    // The order and each stock row are locked: a double receipt, or a receipt
    // racing a sale, must not double-count or lose stock.
    let mut tx = state.db.begin().await?;
    let mut po = db::lock(&mut *tx, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if po.status == "cancelled" || po.status == "received" {
        return Err(AppError::bad_request(
            "Purchase order cannot be received in its current status",
        ));
    }

    let mut items = db::items_of(&mut *tx, po.id).await?;
    let index: HashMap<i64, usize> = items.iter().enumerate().map(|(i, it)| (it.id, i)).collect();

    for line in lines {
        let item = index
            .get(&line.item_id.unwrap_or(0))
            .map(|i| &mut items[*i])
            .ok_or_else(|| AppError::bad_request("Purchase order item not found"))?;
        let received = line.received_quantity.unwrap_or(0);
        let total = (i64::from(item.received_quantity) + received).min(i64::from(item.quantity));

        if let Some(variant_id) = item.product_variant_id.filter(|_| received > 0) {
            let before = i64::from(
                stock::lock(&mut tx, variant_id, po.branch_id)
                    .await?
                    .quantity,
            );
            let after = before + received;
            stock::record_movement(
                &mut *tx,
                &NewMovement {
                    variant_id,
                    branch_id: po.branch_id,
                    kind: "IN",
                    quantity: received,
                    stock_before: before,
                    stock_after: after,
                    unit_cost: Some(item.unit_cost),
                    reference_type: Some("purchase_order"),
                    reference_id: Some(po.id),
                    note: None,
                    created_by: None,
                },
            )
            .await?;
            stock::set_quantity(&mut *tx, variant_id, po.branch_id, after).await?;
            if item.unit_cost > Decimal::ZERO
                && products::variant_for_company(&mut *tx, user.company_id, variant_id)
                    .await?
                    .is_some()
            {
                products::set_last_purchase_cost(&mut *tx, variant_id, item.unit_cost).await?;
            }
        }

        db::set_received(&mut *tx, item.id, total).await?;
        item.received_quantity = i32::try_from(total).unwrap_or_default();
    }

    if items.iter().all(|i| i.received_quantity >= i.quantity) {
        po.status = "received".to_owned();
        po.received_at = Some(Utc::now());
    } else if items.iter().any(|i| i.received_quantity > 0) {
        po.status = "partial".to_owned();
    }
    db::set_status(&mut *tx, &po).await?;
    tx.commit().await?;

    let items = db::items_of(&state.db, po.id).await?;
    Ok(success(
        StatusCode::OK,
        "Purchase order received successfully",
        response(
            po,
            Some(items.into_iter().map(ItemResponse::from).collect()),
        ),
    ))
}
