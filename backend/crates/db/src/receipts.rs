//! `receipts`: an immutable snapshot of an order at completion, with its items
//! and payments as JSONB. Handled here as JSON text; the HTTP layer owns the shape.

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::PgExecutor;

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Receipt {
    pub id: i64,
    pub company_id: i64,
    pub order_id: i64,
    pub order_no: String,
    pub cashier_id: i64,
    pub cashier_name: String,
    pub customer_id: Option<i64>,
    pub customer_name: Option<String>,
    pub items: String,
    pub payments: String,
    pub total_amount: Decimal,
    pub total_discount: Decimal,
    pub total_tax: Decimal,
    pub grand_total: Decimal,
    pub notes: Option<String>,
    pub completed_at: DateTime<Utc>,
    pub created_at: DateTime<Utc>,
}

pub struct NewReceipt<'a> {
    pub company_id: i64,
    pub order_id: i64,
    pub order_no: &'a str,
    pub cashier_id: i64,
    pub cashier_name: &'a str,
    pub customer_id: Option<i64>,
    pub customer_name: Option<&'a str>,
    pub items: &'a str,
    pub payments: &'a str,
    pub total_amount: Decimal,
    pub total_discount: Decimal,
    pub total_tax: Decimal,
    pub grand_total: Decimal,
    pub notes: Option<&'a str>,
    pub completed_at: DateTime<Utc>,
}

/// Write-once: saving an order's receipt again does nothing.
pub async fn save(db: impl PgExecutor<'_>, r: &NewReceipt<'_>) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO receipts (company_id, order_id, order_no, cashier_id, cashier_name, customer_id, \
             customer_name, items, payments, total_amount, total_discount, total_tax, grand_total, \
             notes, completed_at) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9::jsonb, $10, $11, $12, $13, $14, $15) \
         ON CONFLICT (company_id, order_id) DO NOTHING",
    )
    .bind(r.company_id)
    .bind(r.order_id)
    .bind(r.order_no)
    .bind(r.cashier_id)
    .bind(r.cashier_name)
    .bind(r.customer_id)
    .bind(r.customer_name)
    .bind(r.items)
    .bind(r.payments)
    .bind(r.total_amount)
    .bind(r.total_discount)
    .bind(r.total_tax)
    .bind(r.grand_total)
    .bind(r.notes)
    .bind(r.completed_at)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn by_order(
    db: impl PgExecutor<'_>,
    company_id: i64,
    order_id: i64,
) -> sqlx::Result<Option<Receipt>> {
    sqlx::query_as(
        "SELECT id, company_id, order_id, order_no, cashier_id, cashier_name, customer_id, customer_name, \
             items::text AS items, payments::text AS payments, total_amount, total_discount, total_tax, \
             grand_total, notes, completed_at, created_at \
         FROM receipts WHERE company_id = $1 AND order_id = $2",
    )
    .bind(company_id)
    .bind(order_id)
    .fetch_optional(db)
    .await
}
