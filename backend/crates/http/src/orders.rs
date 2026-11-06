//! `/api/v1/orders`: same requests, responses, and errors as the Go `OrderHandler`,
//! `OrderService`, and `ReceiptService`. Each transition locks the order and runs
//! in one transaction with its stock movements.
//!
//! Go behavior kept for the web app: a refund never changes the order's
//! `payment_status`, an update doesn't re-apply a promotion, and only the complete
//! endpoint saves a receipt, not auto-completion.

use std::str::FromStr;

use axum::{
    Router,
    extract::State,
    http::StatusCode,
    routing::{get, post},
};
use chrono::{DateTime, Local, Utc};
use nexpos_db::{
    company_settings::{self, CompanySettings},
    customers,
    orders::{self as db, ListFilter, NewItem, NewOrder, Order, OrderItem, Payment},
    products::{self, SaleVariant},
    promotions::{self, Promotion},
    receipts::{self, NewReceipt, Receipt},
    stock::{self, NewMovement},
};
use nexpos_domain::pricing::{self, PriceRules, PromoTerms};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use sqlx::{PgConnection, PgExecutor, PgPool};

use crate::{
    AppError, AppResult, AppState,
    envelope::{FieldError, Pagination, Reply, success, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
    products::go_numbers,
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/orders", get(list).post(create))
        // Echo matches these against `/orders/:id`, where "preview" isn't a number.
        .route(
            "/api/v1/orders/preview",
            post(preview).get(preview_as_id).put(preview_as_id),
        )
        .route("/api/v1/orders/{id}", get(show).put(update))
        .route("/api/v1/orders/{id}/confirm", post(confirm))
        .route("/api/v1/orders/{id}/payments", post(add_payment))
        .route("/api/v1/orders/{id}/complete", post(complete))
        .route("/api/v1/orders/{id}/cancel", post(cancel))
        .route("/api/v1/orders/{id}/void", post(void))
        .route("/api/v1/orders/{id}/refund", post(refund))
        .route("/api/v1/orders/{id}/receipt", get(receipt))
}

const FULFILLMENT_TYPES: &str = "counter delivery";
const PAYMENT_METHODS: &str = "cash debit_card credit_card e_wallet bank_transfer qris";

// ============== Requests ==============

#[derive(Default, Deserialize)]
#[serde(default)]
struct ItemInput {
    product_variant_id: Option<i64>,
    quantity: Option<i64>,
    #[serde(deserialize_with = "crate::money::deserialize")]
    discount_amount: Decimal,
}

impl ItemInput {
    fn quantity(&self) -> i64 {
        self.quantity.unwrap_or(0)
    }
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct PaymentInput {
    method: Option<String>,
    #[serde(deserialize_with = "crate::money::deserialize")]
    amount: Decimal,
    reference_no: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CreateRequest {
    customer_id: Option<i64>,
    fulfillment_type: Option<String>,
    shipping_address: Option<String>,
    items: Option<Vec<ItemInput>>,
    payments: Option<Vec<PaymentInput>>,
    notes: Option<String>,
    #[serde(deserialize_with = "deserialize_uuid")]
    offline_id: Option<String>,
    promo_code: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct UpdateRequest {
    customer_id: Option<i64>,
    fulfillment_type: Option<String>,
    shipping_address: Option<String>,
    items: Option<Vec<ItemInput>>,
    notes: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct PreviewRequest {
    items: Option<Vec<ItemInput>>,
    promo_code: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct AddPaymentRequest {
    payments: Option<Vec<PaymentInput>>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CompleteRequest {
    fulfillment_status: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct ReasonRequest {
    reason: Option<String>,
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct RefundRequest {
    payment_id: Option<i64>,
    #[serde(deserialize_with = "crate::money::deserialize")]
    amount: Decimal,
    refund_reason: Option<String>,
}

/// `*uuid.UUID`: canonical text, or Go's bad-body error for anything else.
fn deserialize_uuid<'de, D: serde::Deserializer<'de>>(d: D) -> Result<Option<String>, D::Error> {
    let raw: Option<String> = Option::deserialize(d)?;
    raw.map(|text| {
        uuid::Uuid::parse_str(&text)
            .map(|id| id.to_string())
            .map_err(|_| serde::de::Error::custom(crate::money::INVALID_DECIMAL))
    })
    .transpose()
}

fn check_items(v: &mut Validator, items: &[ItemInput]) {
    for item in items {
        v.int("quantity", item.quantity(), &[Rule::Required, Rule::Min(1)]);
        v.decimal(
            "discount_amount",
            item.discount_amount,
            &[Rule::NotNegative],
        );
    }
}

fn check_payments(v: &mut Validator, payments: &[PaymentInput]) {
    for payment in payments {
        v.field(
            "method",
            payment.method.as_deref().unwrap_or_default(),
            &[Rule::Required, Rule::OneOf(PAYMENT_METHODS)],
        );
        v.decimal("amount", payment.amount, &[Rule::Required, Rule::Positive]);
    }
}

fn check_reason(req: &ReasonRequest) -> AppResult<&str> {
    let reason = req.reason.as_deref().unwrap_or_default();
    Validator::new()
        .field("reason", reason, &[Rule::Required])
        .finish()?;
    Ok(reason)
}

// ============== Responses ==============

#[derive(Serialize)]
struct ItemResponse {
    id: i64,
    product_id: Option<i64>,
    product_variant_id: Option<i64>,
    sku: String,
    product_name: String,
    variant_name: Option<String>,
    variant_attributes: Option<Value>,
    #[serde(serialize_with = "crate::money::serialize")]
    unit_price: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    unit_cost: Decimal,
    quantity: i32,
    #[serde(serialize_with = "crate::money::serialize")]
    discount_amount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    tax_amount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    subtotal: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    cogs_amount: Decimal,
}

impl From<OrderItem> for ItemResponse {
    fn from(i: OrderItem) -> Self {
        Self {
            id: i.id,
            product_id: i.product_id,
            product_variant_id: i.product_variant_id,
            sku: i.sku,
            product_name: i.product_name,
            variant_name: i.variant_name,
            variant_attributes: i.variant_attributes.map(|a| go_numbers(a.0)),
            unit_price: i.unit_price,
            unit_cost: i.unit_cost,
            quantity: i.quantity,
            discount_amount: i.discount_amount,
            tax_amount: i.tax_amount,
            subtotal: i.subtotal,
            cogs_amount: i.cogs_amount,
        }
    }
}

#[derive(Serialize)]
struct PaymentResponse {
    id: i64,
    method: String,
    #[serde(serialize_with = "crate::money::serialize")]
    amount: Decimal,
    reference_no: Option<String>,
    status: String,
    #[serde(serialize_with = "crate::money::serialize")]
    refunded_amount: Decimal,
    #[serde(
        serialize_with = "crate::time::serialize_some",
        skip_serializing_if = "Option::is_none"
    )]
    refunded_at: Option<DateTime<Utc>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    refund_reason: Option<String>,
    #[serde(serialize_with = "crate::time::serialize")]
    paid_at: DateTime<Utc>,
}

impl From<Payment> for PaymentResponse {
    fn from(p: Payment) -> Self {
        Self {
            id: p.id,
            method: p.method,
            amount: p.amount,
            reference_no: p.reference_no,
            status: p.status,
            refunded_amount: p.refunded_amount,
            refunded_at: p.refunded_at,
            refund_reason: p.refund_reason,
            paid_at: p.paid_at,
        }
    }
}

#[derive(Serialize)]
struct AppliedPromotion {
    id: i64,
    code: String,
    name: String,
    #[serde(serialize_with = "crate::money::serialize")]
    discount_amount: Decimal,
}

#[derive(Serialize)]
struct OrderResponse {
    id: i64,
    order_no: String,
    customer_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    customer_name: Option<String>,
    cashier_id: i64,
    #[serde(skip_serializing_if = "String::is_empty")]
    cashier_name: String,
    status: String,
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_discount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_tax: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    grand_total: Decimal,
    notes: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    applied_promotion: Option<AppliedPromotion>,
    payment_status: String,
    fulfillment_type: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    fulfillment_status: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    shipping_address: Option<String>,
    #[serde(
        serialize_with = "crate::time::serialize_some",
        skip_serializing_if = "Option::is_none"
    )]
    confirmed_at: Option<DateTime<Utc>>,
    #[serde(
        serialize_with = "crate::time::serialize_some",
        skip_serializing_if = "Option::is_none"
    )]
    paid_at: Option<DateTime<Utc>>,
    #[serde(
        serialize_with = "crate::time::serialize_some",
        skip_serializing_if = "Option::is_none"
    )]
    completed_at: Option<DateTime<Utc>>,
    #[serde(
        serialize_with = "crate::time::serialize_some",
        skip_serializing_if = "Option::is_none"
    )]
    cancelled_at: Option<DateTime<Utc>>,
    #[serde(
        serialize_with = "crate::time::serialize_some",
        skip_serializing_if = "Option::is_none"
    )]
    voided_at: Option<DateTime<Utc>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    cancel_reason: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    void_reason: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    offline_id: Option<String>,
    #[serde(
        serialize_with = "crate::time::serialize_some",
        skip_serializing_if = "Option::is_none"
    )]
    synced_at: Option<DateTime<Utc>>,
    #[serde(serialize_with = "crate::money::serialize")]
    paid_amount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    balance_due: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    refunded_total: Decimal,
    items: Vec<ItemResponse>,
    payments: Vec<PaymentResponse>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

/// The promotion saved on an order, as Go reads it back: nothing without a code.
fn applied_promotion(saved: Option<&Value>) -> Option<AppliedPromotion> {
    let saved = saved?.as_object()?;
    let code = saved.get("code")?.as_str().filter(|c| !c.is_empty())?;
    let discount_amount = saved
        .get("discount_amount")
        .and_then(|value| match value {
            Value::Number(n) => Decimal::from_str(&n.to_string())
                .or_else(|_| Decimal::from_scientific(&n.to_string()))
                .ok(),
            Value::String(s) => Decimal::from_str(s).ok(),
            _ => None,
        })
        .unwrap_or_default();
    Some(AppliedPromotion {
        id: saved.get("id").and_then(Value::as_i64).unwrap_or(0),
        code: code.to_owned(),
        name: saved
            .get("name")
            .and_then(Value::as_str)
            .unwrap_or_default()
            .to_owned(),
        discount_amount,
    })
}

/// Responses for a page of orders, with their items, payments, and names in one
/// query each.
async fn responses(
    db: &PgPool,
    company_id: i64,
    orders: Vec<Order>,
) -> AppResult<Vec<OrderResponse>> {
    let ids: Vec<i64> = orders.iter().map(|o| o.id).collect();
    let customer_ids: Vec<i64> = orders.iter().filter_map(|o| o.customer_id).collect();
    let cashier_ids: Vec<i64> = orders.iter().map(|o| o.cashier_id).collect();
    let mut items = db::items_by_order(db, &ids).await?;
    let mut payments = db::payments_by_order(db, &ids).await?;
    let customers = db::customer_names(db, company_id, &customer_ids).await?;
    let cashiers = db::user_names(db, company_id, &cashier_ids).await?;

    Ok(orders
        .into_iter()
        .map(|o| {
            let payments = payments.remove(&o.id).unwrap_or_default();
            let mut paid_amount = Decimal::ZERO;
            let mut refunded_total = Decimal::ZERO;
            for p in payments.iter().filter(|p| p.status != "failed") {
                paid_amount += p.amount - p.refunded_amount;
                refunded_total += p.refunded_amount;
            }
            OrderResponse {
                applied_promotion: applied_promotion(o.applied_promotions.as_ref().map(|a| &a.0)),
                customer_name: o.customer_id.and_then(|id| customers.get(&id).cloned()),
                cashier_name: cashiers.get(&o.cashier_id).cloned().unwrap_or_default(),
                balance_due: o.grand_total - paid_amount,
                paid_amount,
                refunded_total,
                items: items
                    .remove(&o.id)
                    .unwrap_or_default()
                    .into_iter()
                    .map(ItemResponse::from)
                    .collect(),
                payments: payments.into_iter().map(PaymentResponse::from).collect(),
                id: o.id,
                order_no: o.order_no,
                customer_id: o.customer_id,
                cashier_id: o.cashier_id,
                status: o.status,
                total_amount: o.total_amount,
                total_discount: o.total_discount,
                total_tax: o.total_tax,
                grand_total: o.grand_total,
                notes: o.notes,
                payment_status: o.payment_status,
                fulfillment_type: o.fulfillment_type,
                fulfillment_status: o.fulfillment_status,
                shipping_address: o.shipping_address,
                confirmed_at: o.confirmed_at,
                paid_at: o.paid_at,
                completed_at: o.completed_at,
                cancelled_at: o.cancelled_at,
                voided_at: o.voided_at,
                cancel_reason: o.cancel_reason,
                void_reason: o.void_reason,
                offline_id: o.offline_id,
                synced_at: o.synced_at,
                created_at: o.created_at,
                updated_at: o.updated_at,
            }
        })
        .collect())
}

async fn load(db: &PgPool, company_id: i64, id: i64) -> AppResult<OrderResponse> {
    let order = db::get(db, company_id, id).await?.ok_or_else(not_found)?;
    responses(db, company_id, vec![order])
        .await?
        .pop()
        .ok_or_else(not_found)
}

/// Go's message is "Order not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Order not found")
}

fn customer_not_found() -> AppError {
    AppError::not_found("Customer not found")
}

// ============== Pricing ==============

async fn settings_of(db: impl PgExecutor<'_>, company_id: i64) -> AppResult<CompanySettings> {
    Ok(company_settings::get(db, company_id)
        .await?
        .unwrap_or_else(|| crate::company_settings::defaults(company_id)))
}

fn price_rules(s: &CompanySettings) -> PriceRules {
    PriceRules {
        tax_enabled: s.tax_enabled,
        tax_rate: s.tax_rate,
        tax_inclusive: s.tax_inclusive,
        rounding_enabled: s.rounding_enabled,
        rounding_amount: s.rounding_amount,
    }
}

async fn sale_variant(
    db: impl PgExecutor<'_>,
    company_id: i64,
    item: &ItemInput,
) -> AppResult<SaleVariant> {
    let id = item
        .product_variant_id
        .ok_or_else(|| AppError::bad_request("Product variant is required"))?;
    products::variant_for_company(db, company_id, id)
        .await?
        .ok_or_else(|| AppError::not_found("Product variant not found"))
}

fn unit_price(v: &SaleVariant) -> Decimal {
    pricing::effective_price(v.price, v.sale_price, v.sale_start, v.sale_end, Utc::now())
}

/// A line ready to insert, and what it adds to the order's totals.
struct PricedLine {
    variant: SaleVariant,
    quantity: i64,
    unit_price: Decimal,
    discount: Decimal,
    subtotal: Decimal,
    gross: Decimal,
    tax: Decimal,
}

fn price_line(variant: SaleVariant, item: &ItemInput, rules: &PriceRules) -> PricedLine {
    let unit_price = unit_price(&variant);
    let quantity = item.quantity();
    let amounts = pricing::line_amounts(unit_price, quantity, item.discount_amount, rules);
    PricedLine {
        variant,
        quantity,
        unit_price,
        discount: item.discount_amount,
        subtotal: amounts.subtotal,
        gross: amounts.gross,
        tax: amounts.tax,
    }
}

async fn insert_line(
    db: impl PgExecutor<'_>,
    order_id: i64,
    line: &PricedLine,
) -> sqlx::Result<()> {
    let v = &line.variant;
    db::create_item(
        db,
        &NewItem {
            order_id,
            product_id: v.product_id,
            product_variant_id: v.id,
            sku: &v.sku,
            product_name: &v.product_name,
            variant_name: &v.name,
            variant_attributes: v.attributes.as_ref(),
            unit_price: line.unit_price,
            unit_cost: v.standard_cost,
            quantity: line.quantity,
            discount_amount: line.discount,
            tax_amount: line.tax,
            subtotal: line.subtotal,
            cogs_amount: v.standard_cost * Decimal::from(line.quantity),
        },
    )
    .await
}

/// The best promotion for this subtotal: the named one, or the best active one.
/// A failed lookup means no promotion, as in Go.
async fn best_promotion(
    db: &PgPool,
    company_id: i64,
    code: Option<&str>,
    subtotal: Decimal,
) -> (Option<Promotion>, Decimal) {
    let now = Utc::now();
    let candidates = match code.filter(|c| !c.is_empty()) {
        Some(code) => promotions::by_code(db, company_id, code)
            .await
            .ok()
            .flatten()
            .filter(|p| p.is_active && p.start_at <= now && p.end_at.is_none_or(|end| end >= now))
            .into_iter()
            .collect(),
        None => promotions::active(db, company_id, now)
            .await
            .unwrap_or_default(),
    };

    let mut best = None;
    let mut best_discount = Decimal::ZERO;
    for promo in candidates {
        let terms = PromoTerms {
            discount_type: promo.discount_type.as_deref(),
            discount_value: promo.discount_value,
            min_purchase: promo.min_purchase,
            max_discount: promo.max_discount,
        };
        if let Some(discount) = pricing::promotion_discount(&terms, subtotal)
            && discount > best_discount
        {
            best_discount = discount;
            best = Some(promo);
        }
    }
    (best, best_discount)
}

fn applied_response(promo: &Promotion, discount: Decimal) -> AppliedPromotion {
    AppliedPromotion {
        id: promo.id,
        code: promo.code.clone(),
        name: promo.name.clone(),
        discount_amount: discount,
    }
}

// ============== Stock ==============

/// Moves stock for every tracked line, locking in variant order so two orders
/// sharing variants cannot deadlock. A sale below zero floors at zero and leaves a
/// trail for the next stock opname.
async fn move_stock(
    conn: &mut PgConnection,
    order: &Order,
    kind: &str,
    reference_type: &str,
) -> sqlx::Result<()> {
    let mut items = db::items_of(&mut *conn, order.id).await?;
    items.sort_by_key(|item| item.product_variant_id.unwrap_or(0));

    for item in items {
        let Some(variant_id) = item.product_variant_id else {
            continue;
        };
        let current = stock::lock(conn, variant_id, order.branch_id).await?;
        let before = i64::from(current.quantity);
        let quantity = i64::from(item.quantity);
        let mut after = if kind == "OUT" {
            before - quantity
        } else {
            before + quantity
        };
        if after < 0 {
            tracing::warn!(
                variant_id,
                branch_id = order.branch_id,
                order_id = order.id,
                stock = before,
                sold = quantity,
                "stock below zero on sale, clamping to 0"
            );
            after = 0;
        }
        stock::record_movement(
            &mut *conn,
            &NewMovement {
                variant_id,
                branch_id: order.branch_id,
                kind,
                quantity,
                stock_before: before,
                stock_after: after,
                unit_cost: None,
                reference_type: Some(reference_type),
                reference_id: Some(order.id),
                note: None,
                created_by: None,
            },
        )
        .await?;
        stock::set_quantity(&mut *conn, variant_id, order.branch_id, after).await?;
    }
    Ok(())
}

// ============== Payments ==============

fn apply_paid(order: &mut Order, total_paid: Decimal) {
    let status = pricing::payment_status(total_paid, order.grand_total);
    order.payment_status = status.to_owned();
    if status == "paid" {
        order.paid_at.get_or_insert_with(Utc::now);
    } else {
        order.paid_at = None;
    }
}

/// Records payments, then completes a fully paid counter order when the company
/// wants that, deducting its stock in the same transaction.
async fn process_payments(
    conn: &mut PgConnection,
    order: &mut Order,
    payments: &[PaymentInput],
    settings: &CompanySettings,
) -> AppResult<()> {
    let now = Utc::now();
    for p in payments {
        db::create_payment(
            &mut *conn,
            order.id,
            p.method.as_deref().unwrap_or_default(),
            p.amount,
            p.reference_no.as_deref(),
            now,
        )
        .await?;
    }
    apply_paid(order, db::total_paid(&mut *conn, order.id).await?);

    let auto_complete = settings.auto_complete_counter_orders
        && order.fulfillment_type == "counter"
        && order.payment_status == "paid"
        && (order.status == "draft" || order.status == "confirmed");
    if auto_complete {
        if order.status == "draft" {
            order.confirmed_at = Some(now);
        }
        order.status = "completed".to_owned();
        order.completed_at = Some(now);
        order.paid_at = Some(now);
    }
    db::update(&mut *conn, order).await?;
    if auto_complete {
        move_stock(conn, order, "OUT", "order").await?;
    }
    Ok(())
}

// ============== Handlers ==============

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<OrderResponse>>> {
    let invalid = |_| AppError::bad_request("invalid query parameters");
    // Echo binds the whole query strictly here, unlike the other lists.
    let page = query.strict_int("page").map_err(invalid)?.unwrap_or(0);
    let per_page = query.strict_int("per_page").map_err(invalid)?.unwrap_or(0);
    let customer_id = query.strict_opt_int("customer_id").map_err(invalid)?;
    let page = page.max(1);
    let per_page = match per_page {
        n if n < 1 => 20,
        n => n.min(100),
    };
    let filter = ListFilter {
        search: query.get("search"),
        status: query.opt("status"),
        payment_status: query.opt("payment_status"),
        fulfillment_type: query.opt("fulfillment_type"),
        fulfillment_status: query.opt("fulfillment_status"),
        customer_id,
        date_from: query.opt("date_from"),
        date_to: query.opt("date_to"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list(&state.db, user.company_id, &filter).await?;
    let data = responses(&state.db, user.company_id, rows).await?;
    let pagination = Pagination::new(total, page, per_page, data.len() as i64);
    Ok(success_with_pagination(
        StatusCode::OK,
        "Orders retrieved successfully",
        data,
        pagination,
        None,
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<OrderResponse>> {
    let id = parse_id(&id, "order")?;
    let order = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Order retrieved successfully",
        order,
    ))
}

#[derive(Serialize)]
struct PreviewResponse {
    #[serde(serialize_with = "crate::money::serialize")]
    subtotal: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    item_discount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    promo_discount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    tax: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    grand_total: Decimal,
    #[serde(skip_serializing_if = "Option::is_none")]
    applied_promotion: Option<AppliedPromotion>,
}

async fn preview_as_id(_user: AuthUser) -> AppError {
    AppError::bad_request("invalid order ID")
}

async fn preview(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<PreviewRequest>,
) -> AppResult<Reply<PreviewResponse>> {
    let mut v = Validator::new();
    v.list("items", req.items.as_ref().map(Vec::len));
    check_items(&mut v, req.items.as_deref().unwrap_or_default());
    v.finish()?;
    let items = req.items.as_deref().unwrap_or_default();

    let settings = settings_of(&state.db, user.company_id).await?;
    let rules = price_rules(&settings);
    let (mut gross, mut item_discount, mut tax) = (Decimal::ZERO, Decimal::ZERO, Decimal::ZERO);
    for item in items {
        let variant = sale_variant(&state.db, user.company_id, item).await?;
        let line = price_line(variant, item, &rules);
        gross += line.gross;
        item_discount += line.discount;
        tax += line.tax;
    }

    let subtotal = gross - item_discount;
    let (promo, promo_discount) = best_promotion(
        &state.db,
        user.company_id,
        req.promo_code.as_deref(),
        subtotal,
    )
    .await;
    Ok(success(
        StatusCode::OK,
        "Order preview calculated",
        PreviewResponse {
            subtotal,
            item_discount,
            promo_discount,
            tax,
            grand_total: pricing::grand_total(gross, item_discount + promo_discount, tax, &rules),
            applied_promotion: promo.as_ref().map(|p| applied_response(p, promo_discount)),
        },
    ))
}

/// The JSON saved on an order: what `applied_promotion` shows later.
fn promotion_json(promo: &Promotion, discount: Decimal) -> AppResult<String> {
    serde_json::to_string(&applied_response(promo, discount))
        .map_err(|e| AppError::Internal(e.into()))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<CreateRequest>,
) -> AppResult<Reply<OrderResponse>> {
    let mut v = Validator::new();
    if let Some(kind) = req.fulfillment_type.as_deref().filter(|k| !k.is_empty()) {
        v.field("fulfillment_type", kind, &[Rule::OneOf(FULFILLMENT_TYPES)]);
    }
    v.list("items", req.items.as_ref().map(Vec::len));
    check_items(&mut v, req.items.as_deref().unwrap_or_default());
    v.finish()?;
    let items = req.items.as_deref().unwrap_or_default();
    let branch_id = user.branch_id.unwrap_or(1);

    let settings = settings_of(&state.db, user.company_id).await?;
    if let Some(customer) = req.customer_id
        && customers::get(&state.db, user.company_id, customer)
            .await?
            .is_none()
    {
        return Err(customer_not_found());
    }
    let fulfillment_type = req
        .fulfillment_type
        .as_deref()
        .filter(|k| !k.is_empty())
        .unwrap_or("counter");
    let delivery = fulfillment_type == "delivery";
    if delivery && settings.require_customer_for_delivery && req.customer_id.is_none() {
        return Err(AppError::bad_request(
            "Customer is required for delivery orders",
        ));
    }

    let date = Local::now().format("%Y%m%d").to_string();
    let order_no = db::next_order_no(&state.db, user.company_id, branch_id, &date).await;

    let rules = price_rules(&settings);
    let (mut gross, mut discount, mut tax) = (Decimal::ZERO, Decimal::ZERO, Decimal::ZERO);
    let mut lines = Vec::with_capacity(items.len());
    for item in items {
        let variant = sale_variant(&state.db, user.company_id, item).await?;
        let available = i64::from(stock::on_hand(&state.db, variant.id, branch_id).await?);
        if available < item.quantity() {
            return Err(AppError::validation(
                format!(
                    "Stok {} tidak cukup (tersedia: {available}, diminta: {})",
                    variant.sku,
                    item.quantity()
                ),
                Vec::<FieldError>::new(),
            ));
        }
        let line = price_line(variant, item, &rules);
        gross += line.gross;
        discount += line.discount;
        tax += line.tax;
        lines.push(line);
    }

    let (promo, promo_discount) = best_promotion(
        &state.db,
        user.company_id,
        req.promo_code.as_deref(),
        gross - discount,
    )
    .await;
    if promo.is_some() {
        discount += promo_discount;
    }
    let saved_promotion = promo
        .as_ref()
        .map(|p| promotion_json(p, promo_discount))
        .transpose()?;

    // The order, its items, payments, and any stock deduction commit together.
    let mut tx = state.db.begin().await?;
    let id = db::create(
        &mut *tx,
        &NewOrder {
            company_id: user.company_id,
            branch_id,
            order_no: &order_no,
            customer_id: req.customer_id,
            cashier_id: user.user_id,
            total_amount: gross,
            total_discount: discount,
            total_tax: tax,
            grand_total: pricing::grand_total(gross, discount, tax, &rules),
            applied_promotions: saved_promotion.as_deref(),
            notes: req.notes.as_deref(),
            fulfillment_type,
            fulfillment_status: delivery.then_some("pending"),
            shipping_address: req.shipping_address.as_deref(),
            offline_id: req.offline_id.as_deref(),
        },
    )
    .await?;
    for line in &lines {
        insert_line(&mut *tx, id, line).await?;
    }
    if let Some(payments) = req.payments.as_deref().filter(|p| !p.is_empty()) {
        let mut order = db::lock(&mut *tx, user.company_id, id)
            .await?
            .ok_or_else(not_found)?;
        process_payments(&mut tx, &mut order, payments, &settings).await?;
    }
    tx.commit().await?;

    let order = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::CREATED,
        "Order created successfully",
        order,
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<UpdateRequest>,
) -> AppResult<Reply<OrderResponse>> {
    let id = parse_id(&id, "order")?;
    let mut v = Validator::new();
    v.optional(
        "fulfillment_type",
        req.fulfillment_type.as_deref(),
        &[Rule::OneOf(FULFILLMENT_TYPES)],
    );
    if let Some(items) = &req.items {
        v.list("items", Some(items.len()));
        check_items(&mut v, items);
    }
    v.finish()?;

    // Locked, with items and totals replaced in one transaction, so a failure
    // midway cannot leave the draft without its items.
    let mut tx = state.db.begin().await?;
    let mut order = db::lock(&mut *tx, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if order.status != "draft" {
        return Err(AppError::bad_request("Only draft orders can be updated"));
    }
    if let Some(customer) = req.customer_id {
        if customers::get(&mut *tx, user.company_id, customer)
            .await?
            .is_none()
        {
            return Err(customer_not_found());
        }
        order.customer_id = Some(customer);
    }
    if let Some(kind) = req.fulfillment_type {
        order.fulfillment_type = kind;
    }
    if req.shipping_address.is_some() {
        order.shipping_address = req.shipping_address;
    }
    if req.notes.is_some() {
        order.notes = req.notes;
    }

    if let Some(items) = req.items.as_deref().filter(|i| !i.is_empty()) {
        db::delete_items(&mut *tx, order.id).await?;
        let settings = settings_of(&mut *tx, user.company_id).await?;
        let rules = price_rules(&settings);
        let (mut gross, mut discount, mut tax) = (Decimal::ZERO, Decimal::ZERO, Decimal::ZERO);
        for item in items {
            let variant = sale_variant(&mut *tx, user.company_id, item).await?;
            let line = price_line(variant, item, &rules);
            gross += line.gross;
            discount += line.discount;
            tax += line.tax;
            insert_line(&mut *tx, order.id, &line).await?;
        }
        order.total_amount = gross;
        order.total_discount = discount;
        order.total_tax = tax;
        order.grand_total = pricing::grand_total(gross, discount, tax, &rules);
    }
    db::update(&mut *tx, &order).await?;
    tx.commit().await?;

    let order = load(&state.db, user.company_id, id).await?;
    Ok(success(StatusCode::OK, "Order updated successfully", order))
}

async fn confirm(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<OrderResponse>> {
    let id = parse_id(&id, "order")?;
    let mut tx = state.db.begin().await?;
    let mut order = db::lock(&mut *tx, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if order.status != "draft" {
        return Err(AppError::bad_request("Only draft orders can be confirmed"));
    }
    order.status = "confirmed".to_owned();
    order.confirmed_at = Some(Utc::now());
    db::update(&mut *tx, &order).await?;
    tx.commit().await?;

    let order = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Order confirmed successfully",
        order,
    ))
}

async fn add_payment(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<AddPaymentRequest>,
) -> AppResult<Reply<OrderResponse>> {
    let id = parse_id(&id, "order")?;
    let mut v = Validator::new();
    v.list("payments", req.payments.as_ref().map(Vec::len));
    check_payments(&mut v, req.payments.as_deref().unwrap_or_default());
    v.finish()?;

    let mut tx = state.db.begin().await?;
    let mut order = db::lock(&mut *tx, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if order.status != "draft" && order.status != "confirmed" {
        return Err(AppError::bad_request("Cannot add payment to this order"));
    }
    let settings = settings_of(&mut *tx, user.company_id).await?;
    process_payments(
        &mut tx,
        &mut order,
        req.payments.as_deref().unwrap_or_default(),
        &settings,
    )
    .await?;
    tx.commit().await?;

    let order = load(&state.db, user.company_id, id).await?;
    Ok(success(StatusCode::OK, "Payment added successfully", order))
}

async fn complete(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<CompleteRequest>,
) -> AppResult<Reply<OrderResponse>> {
    let id = parse_id(&id, "order")?;
    let mut tx = state.db.begin().await?;
    let mut order = db::lock(&mut *tx, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if order.status != "confirmed" {
        return Err(AppError::bad_request(
            "Only confirmed orders can be completed",
        ));
    }
    if order.payment_status != "paid" {
        return Err(AppError::bad_request(
            "Order must be fully paid before completion",
        ));
    }
    if order.fulfillment_type == "delivery" && req.fulfillment_status.is_some() {
        order.fulfillment_status = req.fulfillment_status;
    }
    order.status = "completed".to_owned();
    order.completed_at = Some(Utc::now());
    db::update(&mut *tx, &order).await?;
    // Stock moves in the same transaction: either both happen or neither.
    move_stock(&mut tx, &order, "OUT", "order").await?;
    tx.commit().await?;

    let response = load(&state.db, user.company_id, id).await?;
    // Best-effort, like Go: the sale stands even if the snapshot fails.
    if let Err(err) = save_receipt(&state.db, user.company_id, &response).await {
        tracing::error!(order_id = id, error = ?err, "failed to save receipt snapshot");
    }
    Ok(success(
        StatusCode::OK,
        "Order completed successfully",
        response,
    ))
}

async fn cancel(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<ReasonRequest>,
) -> AppResult<Reply<OrderResponse>> {
    let id = parse_id(&id, "order")?;
    let reason = check_reason(&req)?;
    let mut tx = state.db.begin().await?;
    let mut order = db::lock(&mut *tx, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if order.status != "draft" && order.status != "confirmed" {
        return Err(AppError::bad_request("Cannot cancel this order"));
    }
    if order.payment_status == "paid" {
        return Err(AppError::bad_request(
            "Cannot cancel a fully paid order. Use void with refund instead.",
        ));
    }
    order.status = "cancelled".to_owned();
    order.cancelled_at = Some(Utc::now());
    order.cancel_reason = Some(reason.to_owned());
    db::update(&mut *tx, &order).await?;
    tx.commit().await?;

    let order = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Order cancelled successfully",
        order,
    ))
}

async fn void(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<ReasonRequest>,
) -> AppResult<Reply<OrderResponse>> {
    let id = parse_id(&id, "order")?;
    let reason = check_reason(&req)?;
    let mut tx = state.db.begin().await?;
    let mut order = db::lock(&mut *tx, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if order.status != "confirmed" && order.status != "completed" {
        return Err(AppError::bad_request("Cannot void this order"));
    }
    let was_completed = order.status == "completed";
    order.status = "voided".to_owned();
    order.voided_at = Some(Utc::now());
    order.void_reason = Some(reason.to_owned());
    if order.payment_status == "paid" || order.payment_status == "partial" {
        order.payment_status = "refunded".to_owned();
    }
    db::update(&mut *tx, &order).await?;
    // Stock was only deducted if the order was completed.
    if was_completed {
        move_stock(&mut tx, &order, "IN", "order_void").await?;
    }
    tx.commit().await?;

    let order = load(&state.db, user.company_id, id).await?;
    Ok(success(StatusCode::OK, "Order voided successfully", order))
}

async fn refund(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<RefundRequest>,
) -> AppResult<Reply<OrderResponse>> {
    let id = parse_id(&id, "order")?;
    let reason = req.refund_reason.as_deref().unwrap_or_default();
    let payment_id = req.payment_id.unwrap_or(0);
    Validator::new()
        .int("payment_id", payment_id, &[Rule::Required])
        .decimal("amount", req.amount, &[Rule::Required, Rule::Positive])
        .field("refund_reason", reason, &[Rule::Required])
        .finish()?;

    let mut tx = state.db.begin().await?;
    if db::get(&mut *tx, user.company_id, id).await?.is_none() {
        return Err(not_found());
    }
    let mut payment = db::lock_payment(&mut *tx, payment_id)
        .await?
        .ok_or_else(|| AppError::not_found("Payment not found"))?;
    if payment.order_id != id {
        return Err(AppError::bad_request(
            "Payment does not belong to this order",
        ));
    }
    if req.amount > payment.amount - payment.refunded_amount {
        return Err(AppError::bad_request(
            "Refund amount exceeds available amount",
        ));
    }
    payment.refunded_amount += req.amount;
    payment.refunded_at = Some(Utc::now());
    payment.refund_reason = Some(reason.to_owned());
    payment.status = if payment.refunded_amount >= payment.amount {
        "refunded"
    } else {
        "partially_refunded"
    }
    .to_owned();
    db::update_payment(&mut *tx, &payment).await?;
    tx.commit().await?;

    let order = load(&state.db, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Payment refunded successfully",
        order,
    ))
}

// ============== Receipts ==============

#[derive(Serialize, Deserialize)]
struct ReceiptItem {
    product_name: String,
    variant_name: Option<String>,
    sku: String,
    quantity: i64,
    #[serde(
        serialize_with = "crate::money::serialize",
        deserialize_with = "crate::money::deserialize"
    )]
    unit_price: Decimal,
    #[serde(
        serialize_with = "crate::money::serialize",
        deserialize_with = "crate::money::deserialize"
    )]
    discount: Decimal,
    #[serde(
        serialize_with = "crate::money::serialize",
        deserialize_with = "crate::money::deserialize"
    )]
    tax: Decimal,
    #[serde(
        serialize_with = "crate::money::serialize",
        deserialize_with = "crate::money::deserialize"
    )]
    subtotal: Decimal,
}

#[derive(Serialize, Deserialize)]
struct ReceiptPayment {
    method: String,
    #[serde(
        serialize_with = "crate::money::serialize",
        deserialize_with = "crate::money::deserialize"
    )]
    amount: Decimal,
    reference_no: Option<String>,
}

#[derive(Serialize)]
struct ReceiptResponse {
    id: i64,
    company_id: i64,
    order_id: i64,
    order_no: String,
    cashier_id: i64,
    cashier_name: String,
    customer_id: Option<i64>,
    customer_name: Option<String>,
    items: Vec<ReceiptItem>,
    payments: Vec<ReceiptPayment>,
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_discount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_tax: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    grand_total: Decimal,
    notes: Option<String>,
    #[serde(serialize_with = "crate::time::serialize")]
    completed_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
}

impl TryFrom<Receipt> for ReceiptResponse {
    type Error = serde_json::Error;

    fn try_from(r: Receipt) -> Result<Self, Self::Error> {
        Ok(Self {
            items: serde_json::from_str(&r.items)?,
            payments: serde_json::from_str(&r.payments)?,
            id: r.id,
            company_id: r.company_id,
            order_id: r.order_id,
            order_no: r.order_no,
            cashier_id: r.cashier_id,
            cashier_name: r.cashier_name,
            customer_id: r.customer_id,
            customer_name: r.customer_name,
            total_amount: r.total_amount,
            total_discount: r.total_discount,
            total_tax: r.total_tax,
            grand_total: r.grand_total,
            notes: r.notes,
            completed_at: r.completed_at,
            created_at: r.created_at,
        })
    }
}

async fn save_receipt(db: &PgPool, company_id: i64, o: &OrderResponse) -> anyhow::Result<()> {
    let items: Vec<ReceiptItem> = o
        .items
        .iter()
        .map(|i| ReceiptItem {
            product_name: i.product_name.clone(),
            variant_name: i.variant_name.clone(),
            sku: i.sku.clone(),
            quantity: i64::from(i.quantity),
            unit_price: i.unit_price,
            discount: i.discount_amount,
            tax: i.tax_amount,
            subtotal: i.subtotal,
        })
        .collect();
    let payments: Vec<ReceiptPayment> = o
        .payments
        .iter()
        .map(|p| ReceiptPayment {
            method: p.method.clone(),
            amount: p.amount,
            reference_no: p.reference_no.clone(),
        })
        .collect();
    receipts::save(
        db,
        &NewReceipt {
            company_id,
            order_id: o.id,
            order_no: &o.order_no,
            cashier_id: o.cashier_id,
            cashier_name: &o.cashier_name,
            customer_id: o.customer_id,
            customer_name: o.customer_name.as_deref(),
            items: &serde_json::to_string(&items)?,
            payments: &serde_json::to_string(&payments)?,
            total_amount: o.total_amount,
            total_discount: o.total_discount,
            total_tax: o.total_tax,
            grand_total: o.grand_total,
            notes: o.notes.as_deref(),
            completed_at: o.completed_at.unwrap_or_else(Utc::now),
        },
    )
    .await?;
    Ok(())
}

async fn receipt(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<ReceiptResponse>> {
    let id = parse_id(&id, "order")?;
    let row = receipts::by_order(&state.db, user.company_id, id)
        .await?
        .ok_or_else(|| AppError::not_found("receipt"))?;
    let response = ReceiptResponse::try_from(row).map_err(|e| AppError::Internal(e.into()))?;
    Ok(success(
        StatusCode::OK,
        "Receipt retrieved successfully",
        response,
    ))
}
