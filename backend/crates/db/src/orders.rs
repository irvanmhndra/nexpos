//! `orders`, `order_items`, and `payments`, like the Go repositories. Items,
//! payments, and names load in one query per page rather than one per order.

use std::collections::HashMap;

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use serde_json::Value;
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder, types::Json};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Order {
    pub id: i64,
    pub company_id: i64,
    pub branch_id: i64,
    pub order_no: String,
    pub customer_id: Option<i64>,
    pub cashier_id: i64,
    pub status: String,
    pub total_amount: Decimal,
    pub total_discount: Decimal,
    pub total_tax: Decimal,
    pub grand_total: Decimal,
    pub applied_promotions: Option<Json<Value>>,
    pub notes: Option<String>,
    pub payment_status: String,
    pub fulfillment_type: String,
    pub fulfillment_status: Option<String>,
    pub shipping_address: Option<String>,
    pub confirmed_at: Option<DateTime<Utc>>,
    pub paid_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
    pub cancelled_at: Option<DateTime<Utc>>,
    pub voided_at: Option<DateTime<Utc>>,
    pub cancel_reason: Option<String>,
    pub void_reason: Option<String>,
    pub offline_id: Option<String>,
    pub synced_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct OrderItem {
    pub id: i64,
    pub order_id: i64,
    pub product_id: Option<i64>,
    pub product_variant_id: Option<i64>,
    pub sku: String,
    pub product_name: String,
    pub variant_name: Option<String>,
    pub variant_attributes: Option<Json<Value>>,
    pub unit_price: Decimal,
    pub unit_cost: Decimal,
    pub quantity: i32,
    pub discount_amount: Decimal,
    pub tax_amount: Decimal,
    pub subtotal: Decimal,
    pub cogs_amount: Decimal,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Payment {
    pub id: i64,
    pub order_id: i64,
    pub method: String,
    pub amount: Decimal,
    pub reference_no: Option<String>,
    pub status: String,
    pub refunded_amount: Decimal,
    pub refunded_at: Option<DateTime<Utc>>,
    pub refund_reason: Option<String>,
    pub paid_at: DateTime<Utc>,
}

pub struct NewOrder<'a> {
    pub company_id: i64,
    pub branch_id: i64,
    pub order_no: &'a str,
    pub customer_id: Option<i64>,
    pub cashier_id: i64,
    pub total_amount: Decimal,
    pub total_discount: Decimal,
    pub total_tax: Decimal,
    pub grand_total: Decimal,
    /// JSON text for the `applied_promotions` column.
    pub applied_promotions: Option<&'a str>,
    pub notes: Option<&'a str>,
    pub fulfillment_type: &'a str,
    pub fulfillment_status: Option<&'a str>,
    pub shipping_address: Option<&'a str>,
    /// Canonical UUID text.
    pub offline_id: Option<&'a str>,
}

pub struct NewItem<'a> {
    pub order_id: i64,
    pub product_id: i64,
    pub product_variant_id: i64,
    pub sku: &'a str,
    pub product_name: &'a str,
    pub variant_name: &'a str,
    pub variant_attributes: Option<&'a Json<Value>>,
    pub unit_price: Decimal,
    pub unit_cost: Decimal,
    pub quantity: i64,
    pub discount_amount: Decimal,
    pub tax_amount: Decimal,
    pub subtotal: Decimal,
    pub cogs_amount: Decimal,
}

#[derive(Default)]
pub struct ListFilter<'a> {
    pub search: &'a str,
    pub status: Option<&'a str>,
    pub payment_status: Option<&'a str>,
    pub fulfillment_type: Option<&'a str>,
    pub fulfillment_status: Option<&'a str>,
    pub customer_id: Option<i64>,
    pub date_from: Option<&'a str>,
    pub date_to: Option<&'a str>,
    pub limit: i64,
    pub offset: i64,
}

const COLUMNS: &str = "id, company_id, branch_id, order_no, customer_id, cashier_id, status, \
     total_amount, total_discount, total_tax, grand_total, applied_promotions, notes, \
     payment_status, fulfillment_type, fulfillment_status, shipping_address, \
     confirmed_at, paid_at, completed_at, cancelled_at, voided_at, \
     cancel_reason, void_reason, offline_id::text AS offline_id, synced_at, created_at, updated_at";

const ITEM_COLUMNS: &str = "id, order_id, product_id, product_variant_id, sku, product_name, variant_name, \
     variant_attributes, unit_price, unit_cost, quantity, discount_amount, tax_amount, subtotal, cogs_amount";

const PAYMENT_COLUMNS: &str = "id, order_id, method, amount, reference_no, status, refunded_amount, \
     refunded_at, refund_reason, paid_at";

pub async fn get(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<Option<Order>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM orders WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// Locks the row until the transaction ends, so concurrent state changes on an
/// order run one after another.
pub async fn lock(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<Order>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM orders WHERE id = $1 AND company_id = $2 FOR UPDATE"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// Same filters, order (newest first), and count as the Go repository. Dates are
/// bound as text and cast, so an unparseable one is an error from Postgres.
pub async fn list(
    db: &PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<Order>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !f.search.is_empty() {
            q.push(" AND order_no ILIKE ")
                .push_bind(format!("%{}%", f.search));
        }
        for (column, value) in [
            ("status", f.status),
            ("payment_status", f.payment_status),
            ("fulfillment_type", f.fulfillment_type),
            ("fulfillment_status", f.fulfillment_status),
        ] {
            if let Some(value) = value {
                q.push(format!(" AND {column} = "))
                    .push_bind(value.to_owned());
            }
        }
        if let Some(customer) = f.customer_id {
            q.push(" AND customer_id = ").push_bind(customer);
        }
        if let Some(from) = f.date_from {
            q.push(" AND created_at >= ")
                .push_bind(from.to_owned())
                .push("::timestamptz");
        }
        if let Some(to) = f.date_to {
            q.push(" AND created_at <= ")
                .push_bind(to.to_owned())
                .push("::date + interval '1 day'");
        }
    };
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM orders");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM orders"));
    filter(&mut rows);
    rows.push(" ORDER BY created_at DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

pub async fn create(db: impl PgExecutor<'_>, o: &NewOrder<'_>) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "INSERT INTO orders (company_id, branch_id, order_no, customer_id, cashier_id, status, \
             total_amount, total_discount, total_tax, grand_total, applied_promotions, notes, \
             payment_status, fulfillment_type, fulfillment_status, shipping_address, offline_id) \
         VALUES ($1, $2, $3, $4, $5, 'draft', $6, $7, $8, $9, $10::jsonb, $11, 'unpaid', $12, $13, $14, $15::uuid) \
         RETURNING id",
    )
    .bind(o.company_id)
    .bind(o.branch_id)
    .bind(o.order_no)
    .bind(o.customer_id)
    .bind(o.cashier_id)
    .bind(o.total_amount)
    .bind(o.total_discount)
    .bind(o.total_tax)
    .bind(o.grand_total)
    .bind(o.applied_promotions)
    .bind(o.notes)
    .bind(o.fulfillment_type)
    .bind(o.fulfillment_status)
    .bind(o.shipping_address)
    .bind(o.offline_id)
    .fetch_one(db)
    .await
}

/// Saves everything a transition can change, like the Go `Update`.
pub async fn update(db: impl PgExecutor<'_>, o: &Order) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE orders SET customer_id = $1, status = $2, total_amount = $3, total_discount = $4, \
             total_tax = $5, grand_total = $6, notes = $7, payment_status = $8, fulfillment_type = $9, \
             fulfillment_status = $10, shipping_address = $11, confirmed_at = $12, paid_at = $13, \
             completed_at = $14, cancelled_at = $15, voided_at = $16, cancel_reason = $17, \
             void_reason = $18, synced_at = $19, updated_at = NOW() \
         WHERE id = $20 AND company_id = $21",
    )
    .bind(o.customer_id)
    .bind(&o.status)
    .bind(o.total_amount)
    .bind(o.total_discount)
    .bind(o.total_tax)
    .bind(o.grand_total)
    .bind(&o.notes)
    .bind(&o.payment_status)
    .bind(&o.fulfillment_type)
    .bind(&o.fulfillment_status)
    .bind(&o.shipping_address)
    .bind(o.confirmed_at)
    .bind(o.paid_at)
    .bind(o.completed_at)
    .bind(o.cancelled_at)
    .bind(o.voided_at)
    .bind(&o.cancel_reason)
    .bind(&o.void_reason)
    .bind(o.synced_at)
    .bind(o.id)
    .bind(o.company_id)
    .execute(db)
    .await?;
    Ok(())
}

/// `ORD-<prefix>-NNNN`: one past the highest number of the day for the branch.
pub async fn next_order_no(
    db: impl PgExecutor<'_>,
    company_id: i64,
    branch_id: i64,
    date_prefix: &str,
) -> String {
    let next: i32 = sqlx::query_scalar(
        "SELECT COALESCE(MAX(CAST(SUBSTRING(order_no FROM '[0-9]+$') AS INTEGER)), 0) + 1 \
         FROM orders WHERE company_id = $1 AND branch_id = $2 AND order_no LIKE $3",
    )
    .bind(company_id)
    .bind(branch_id)
    .bind(format!("ORD-{date_prefix}-%"))
    .fetch_one(db)
    .await
    .unwrap_or(1);
    format!("ORD-{date_prefix}-{next:04}")
}

pub async fn create_item(db: impl PgExecutor<'_>, i: &NewItem<'_>) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO order_items (order_id, product_id, product_variant_id, sku, product_name, \
             variant_name, variant_attributes, unit_price, unit_cost, quantity, discount_amount, \
             tax_amount, subtotal, cogs_amount) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
    )
    .bind(i.order_id)
    .bind(i.product_id)
    .bind(i.product_variant_id)
    .bind(i.sku)
    .bind(i.product_name)
    .bind(i.variant_name)
    .bind(i.variant_attributes)
    .bind(i.unit_price)
    .bind(i.unit_cost)
    .bind(i.quantity)
    .bind(i.discount_amount)
    .bind(i.tax_amount)
    .bind(i.subtotal)
    .bind(i.cogs_amount)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn delete_items(db: impl PgExecutor<'_>, order_id: i64) -> sqlx::Result<()> {
    sqlx::query("DELETE FROM order_items WHERE order_id = $1")
        .bind(order_id)
        .execute(db)
        .await?;
    Ok(())
}

pub async fn items_of(db: impl PgExecutor<'_>, order_id: i64) -> sqlx::Result<Vec<OrderItem>> {
    sqlx::query_as(&format!(
        "SELECT {ITEM_COLUMNS} FROM order_items WHERE order_id = $1 ORDER BY id"
    ))
    .bind(order_id)
    .fetch_all(db)
    .await
}

/// Items of many orders, each list in id order.
pub async fn items_by_order(
    db: impl PgExecutor<'_>,
    order_ids: &[i64],
) -> sqlx::Result<HashMap<i64, Vec<OrderItem>>> {
    let rows: Vec<OrderItem> = sqlx::query_as(&format!(
        "SELECT {ITEM_COLUMNS} FROM order_items WHERE order_id = ANY($1) ORDER BY id"
    ))
    .bind(order_ids)
    .fetch_all(db)
    .await?;
    let mut by_order: HashMap<i64, Vec<OrderItem>> = HashMap::new();
    for row in rows {
        by_order.entry(row.order_id).or_default().push(row);
    }
    Ok(by_order)
}

/// Payments of many orders, each list in id order.
pub async fn payments_by_order(
    db: impl PgExecutor<'_>,
    order_ids: &[i64],
) -> sqlx::Result<HashMap<i64, Vec<Payment>>> {
    let rows: Vec<Payment> = sqlx::query_as(&format!(
        "SELECT {PAYMENT_COLUMNS} FROM payments WHERE order_id = ANY($1) ORDER BY id"
    ))
    .bind(order_ids)
    .fetch_all(db)
    .await?;
    let mut by_order: HashMap<i64, Vec<Payment>> = HashMap::new();
    for row in rows {
        by_order.entry(row.order_id).or_default().push(row);
    }
    Ok(by_order)
}

pub async fn customer_names(
    db: impl PgExecutor<'_>,
    company_id: i64,
    ids: &[i64],
) -> sqlx::Result<HashMap<i64, String>> {
    let rows: Vec<(i64, String)> =
        sqlx::query_as("SELECT id, name FROM customers WHERE company_id = $1 AND id = ANY($2)")
            .bind(company_id)
            .bind(ids)
            .fetch_all(db)
            .await?;
    Ok(rows.into_iter().collect())
}

pub async fn user_names(
    db: impl PgExecutor<'_>,
    company_id: i64,
    ids: &[i64],
) -> sqlx::Result<HashMap<i64, String>> {
    let rows: Vec<(i64, String)> =
        sqlx::query_as("SELECT id, name FROM users WHERE company_id = $1 AND id = ANY($2)")
            .bind(company_id)
            .bind(ids)
            .fetch_all(db)
            .await?;
    Ok(rows.into_iter().collect())
}

pub async fn create_payment(
    db: impl PgExecutor<'_>,
    order_id: i64,
    method: &str,
    amount: Decimal,
    reference_no: Option<&str>,
    paid_at: DateTime<Utc>,
) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO payments (order_id, method, amount, reference_no, status, paid_at) \
         VALUES ($1, $2, $3, $4, 'completed', $5)",
    )
    .bind(order_id)
    .bind(method)
    .bind(amount)
    .bind(reference_no)
    .bind(paid_at)
    .execute(db)
    .await?;
    Ok(())
}

/// Locks a payment: the refundable amount is checked and updated under the lock,
/// so concurrent refunds cannot exceed it.
pub async fn lock_payment(db: impl PgExecutor<'_>, id: i64) -> sqlx::Result<Option<Payment>> {
    sqlx::query_as(&format!(
        "SELECT {PAYMENT_COLUMNS} FROM payments WHERE id = $1 FOR UPDATE"
    ))
    .bind(id)
    .fetch_optional(db)
    .await
}

pub async fn update_payment(db: impl PgExecutor<'_>, p: &Payment) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE payments SET status = $1, refunded_amount = $2, refunded_at = $3, refund_reason = $4 \
         WHERE id = $5",
    )
    .bind(&p.status)
    .bind(p.refunded_amount)
    .bind(p.refunded_at)
    .bind(&p.refund_reason)
    .bind(p.id)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn total_paid(db: impl PgExecutor<'_>, order_id: i64) -> sqlx::Result<Decimal> {
    sqlx::query_scalar(
        "SELECT COALESCE(SUM(amount - refunded_amount), 0) FROM payments \
         WHERE order_id = $1 AND status != 'failed'",
    )
    .bind(order_id)
    .fetch_one(db)
    .await
}

/// Cash taken at a branch in `[from, to)`, net of refunds.
pub async fn cash_total(
    db: impl PgExecutor<'_>,
    branch_id: i64,
    from: DateTime<Utc>,
    to: DateTime<Utc>,
) -> sqlx::Result<Decimal> {
    sqlx::query_scalar(
        "SELECT COALESCE(SUM(p.amount - p.refunded_amount), 0) \
         FROM payments p JOIN orders o ON o.id = p.order_id \
         WHERE o.branch_id = $1 AND p.method = 'cash' AND p.status != 'failed' \
           AND p.paid_at >= $2 AND p.paid_at < $3",
    )
    .bind(branch_id)
    .bind(from)
    .bind(to)
    .fetch_one(db)
    .await
}
