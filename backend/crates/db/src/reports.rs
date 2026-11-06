//! Sales reports, the same SQL as the Go `ReportRepository`. Dates are compared
//! in the session's TimeZone, so the pool sets it to the database's zone.
//!
//! `created_at >= $2::date` takes the dates as text; an unparseable one is an
//! error from Postgres, as in Go.

use chrono::NaiveDate;
use rust_decimal::Decimal;
use sqlx::PgExecutor;

#[derive(Debug, Default, sqlx::FromRow)]
pub struct OrderStats {
    pub total_revenue: Decimal,
    pub total_orders: i64,
    pub total_discount: Decimal,
    pub total_tax: Decimal,
    pub completed_orders: i64,
    pub cancelled_orders: i64,
    pub pending_orders: i64,
    pub total_customers: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct ItemStats {
    pub total_items: i64,
    pub total_cogs: Decimal,
}

#[derive(Debug, sqlx::FromRow)]
pub struct SalesTrendRow {
    pub date: NaiveDate,
    pub sales: Decimal,
    pub orders: i64,
    pub items: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct TopProductRow {
    pub product_id: i64,
    pub product_name: String,
    pub total_sold: i64,
    pub total_amount: Decimal,
}

#[derive(Debug, sqlx::FromRow)]
pub struct CategoryRevenueRow {
    pub category_id: i64,
    pub category_name: String,
    pub total_amount: Decimal,
    pub order_count: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct PaymentMethodRow {
    pub method: String,
    pub total_amount: Decimal,
    pub count: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct HourlySalesRow {
    pub hour: i32,
    pub total_amount: Decimal,
    pub order_count: i64,
}

/// Which `company` and date range a report covers.
#[derive(Clone, Copy)]
pub struct Range<'a> {
    pub company_id: i64,
    pub from: &'a str,
    pub to: &'a str,
}

pub async fn order_stats(db: impl PgExecutor<'_>, r: Range<'_>) -> sqlx::Result<OrderStats> {
    sqlx::query_as(
        "SELECT \
            COALESCE(SUM(grand_total) FILTER (WHERE status = 'completed'), 0) AS total_revenue, \
            COUNT(*) FILTER (WHERE status = 'completed') AS total_orders, \
            COALESCE(SUM(total_discount) FILTER (WHERE status = 'completed'), 0) AS total_discount, \
            COALESCE(SUM(total_tax) FILTER (WHERE status = 'completed'), 0) AS total_tax, \
            COUNT(*) FILTER (WHERE status = 'completed') AS completed_orders, \
            COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_orders, \
            COUNT(*) FILTER (WHERE status = 'pending') AS pending_orders, \
            COUNT(DISTINCT customer_id) FILTER (WHERE status = 'completed') AS total_customers \
         FROM orders \
         WHERE company_id = $1 AND created_at >= $2::date AND created_at < ($3::date + interval '1 day')",
    )
    .bind(r.company_id)
    .bind(r.from)
    .bind(r.to)
    .fetch_one(db)
    .await
}

pub async fn item_stats(db: impl PgExecutor<'_>, r: Range<'_>) -> sqlx::Result<ItemStats> {
    sqlx::query_as(
        "SELECT COALESCE(SUM(oi.quantity), 0)::bigint AS total_items, \
            COALESCE(SUM(oi.cogs_amount), 0) AS total_cogs \
         FROM order_items oi JOIN orders o ON o.id = oi.order_id \
         WHERE o.company_id = $1 AND o.created_at >= $2::date \
           AND o.created_at < ($3::date + interval '1 day') AND o.status = 'completed'",
    )
    .bind(r.company_id)
    .bind(r.from)
    .bind(r.to)
    .fetch_one(db)
    .await
}

pub async fn new_customers(db: impl PgExecutor<'_>, r: Range<'_>) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "SELECT COUNT(*) FROM customers \
         WHERE company_id = $1 AND created_at >= $2::date AND created_at < ($3::date + interval '1 day')",
    )
    .bind(r.company_id)
    .bind(r.from)
    .bind(r.to)
    .fetch_one(db)
    .await
}

pub async fn sales_trend(
    db: impl PgExecutor<'_>,
    r: Range<'_>,
) -> sqlx::Result<Vec<SalesTrendRow>> {
    sqlx::query_as(
        "SELECT DATE(created_at) AS date, COALESCE(SUM(grand_total), 0) AS sales, COUNT(*) AS orders, \
            COALESCE(SUM((SELECT SUM(quantity) FROM order_items WHERE order_id = orders.id)), 0)::bigint AS items \
         FROM orders \
         WHERE company_id = $1 AND created_at >= $2::date \
           AND created_at < ($3::date + interval '1 day') AND status = 'completed' \
         GROUP BY DATE(created_at) ORDER BY date ASC",
    )
    .bind(r.company_id)
    .bind(r.from)
    .bind(r.to)
    .fetch_all(db)
    .await
}

pub async fn top_products(
    db: impl PgExecutor<'_>,
    r: Range<'_>,
    limit: i64,
) -> sqlx::Result<Vec<TopProductRow>> {
    sqlx::query_as(
        "SELECT COALESCE(oi.product_id, 0) AS product_id, oi.product_name, \
            SUM(oi.quantity)::bigint AS total_sold, SUM(oi.subtotal) AS total_amount \
         FROM order_items oi JOIN orders o ON o.id = oi.order_id \
         WHERE o.company_id = $1 AND o.created_at >= $2::date \
           AND o.created_at < ($3::date + interval '1 day') AND o.status = 'completed' \
         GROUP BY oi.product_id, oi.product_name ORDER BY total_sold DESC LIMIT $4",
    )
    .bind(r.company_id)
    .bind(r.from)
    .bind(r.to)
    .bind(limit)
    .fetch_all(db)
    .await
}

pub async fn category_revenue(
    db: impl PgExecutor<'_>,
    r: Range<'_>,
) -> sqlx::Result<Vec<CategoryRevenueRow>> {
    sqlx::query_as(
        "SELECT COALESCE(p.product_category_id, 0) AS category_id, \
            COALESCE(pc.name, 'Tanpa Kategori') AS category_name, \
            SUM(oi.subtotal) AS total_amount, COUNT(DISTINCT o.id) AS order_count \
         FROM order_items oi JOIN orders o ON o.id = oi.order_id \
         LEFT JOIN products p ON p.id = oi.product_id \
         LEFT JOIN product_categories pc ON pc.id = p.product_category_id \
         WHERE o.company_id = $1 AND o.created_at >= $2::date \
           AND o.created_at < ($3::date + interval '1 day') AND o.status = 'completed' \
         GROUP BY p.product_category_id, pc.name ORDER BY total_amount DESC",
    )
    .bind(r.company_id)
    .bind(r.from)
    .bind(r.to)
    .fetch_all(db)
    .await
}

pub async fn payment_methods(
    db: impl PgExecutor<'_>,
    r: Range<'_>,
) -> sqlx::Result<Vec<PaymentMethodRow>> {
    sqlx::query_as(
        "SELECT p.method, SUM(p.amount) AS total_amount, COUNT(*) AS count \
         FROM payments p JOIN orders o ON o.id = p.order_id \
         WHERE o.company_id = $1 AND o.created_at >= $2::date \
           AND o.created_at < ($3::date + interval '1 day') AND o.status = 'completed' \
         GROUP BY p.method ORDER BY total_amount DESC",
    )
    .bind(r.company_id)
    .bind(r.from)
    .bind(r.to)
    .fetch_all(db)
    .await
}

pub async fn hourly_sales(
    db: impl PgExecutor<'_>,
    r: Range<'_>,
) -> sqlx::Result<Vec<HourlySalesRow>> {
    sqlx::query_as(
        "SELECT EXTRACT(HOUR FROM created_at)::int AS hour, SUM(grand_total) AS total_amount, \
            COUNT(*) AS order_count \
         FROM orders \
         WHERE company_id = $1 AND created_at >= $2::date \
           AND created_at < ($3::date + interval '1 day') AND status = 'completed' \
         GROUP BY EXTRACT(HOUR FROM created_at) ORDER BY hour ASC",
    )
    .bind(r.company_id)
    .bind(r.from)
    .bind(r.to)
    .fetch_all(db)
    .await
}
