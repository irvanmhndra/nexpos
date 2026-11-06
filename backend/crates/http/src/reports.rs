//! `/api/v1/reports/*`: same responses as the Go `ReportHandler`/`ReportService`.
//! List reports are wrapped twice (`{"data":{"data":[...]}}`), as in Go.

use axum::{Router, extract::State, http::StatusCode, routing::get};
use nexpos_db::reports::{self as db, Range};
use nexpos_domain::report as rules;
use rust_decimal::Decimal;
use serde::Serialize;

use crate::{
    AppError, AppResult, AppState,
    envelope::{Reply, success},
    extract::AuthUser,
    params::QueryParams,
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/reports/summary", get(summary))
        .route("/api/v1/reports/sales-trend", get(sales_trend))
        .route("/api/v1/reports/top-products", get(top_products))
        .route("/api/v1/reports/category-revenue", get(category_revenue))
        .route("/api/v1/reports/payment-methods", get(payment_methods))
        .route("/api/v1/reports/hourly-sales", get(hourly_sales))
}

const DEFAULT_TOP_PRODUCTS: i64 = 10;

/// `date_from` and `date_to` are required but not parsed here; Postgres does that.
fn range<'a>(user: &AuthUser, query: &'a QueryParams) -> AppResult<Range<'a>> {
    let (from, to) = (query.get("date_from"), query.get("date_to"));
    if from.is_empty() || to.is_empty() {
        return Err(AppError::bad_request("date_from and date_to are required"));
    }
    Ok(Range {
        company_id: user.company_id,
        from,
        to,
    })
}

#[derive(Serialize)]
struct List<T: Serialize> {
    data: Vec<T>,
}

#[derive(Serialize)]
struct SummaryResponse {
    #[serde(serialize_with = "crate::money::serialize")]
    total_revenue: Decimal,
    total_orders: i64,
    total_items_sold: i64,
    #[serde(serialize_with = "crate::money::serialize")]
    avg_order_value: Decimal,
    total_customers: i64,
    new_customers: i64,
    completed_orders: i64,
    cancelled_orders: i64,
    pending_orders: i64,
    #[serde(serialize_with = "crate::money::serialize")]
    total_discount: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    total_tax: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    gross_profit: Decimal,
    #[serde(serialize_with = "crate::money::serialize_float")]
    gross_profit_margin: f64,
}

async fn summary(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<SummaryResponse>> {
    let r = range(&user, &query)?;
    let orders = db::order_stats(&state.db, r).await?;
    let items = db::item_stats(&state.db, r).await?;
    // Non-critical, like in Go: a failed count shows as zero.
    let new_customers = db::new_customers(&state.db, r).await.unwrap_or(0);

    let gross_profit = orders.total_revenue - items.total_cogs;
    let body = SummaryResponse {
        total_revenue: orders.total_revenue,
        total_orders: orders.total_orders,
        total_items_sold: items.total_items,
        avg_order_value: rules::avg_order_value(orders.total_revenue, orders.total_orders),
        total_customers: orders.total_customers,
        new_customers,
        completed_orders: orders.completed_orders,
        cancelled_orders: orders.cancelled_orders,
        pending_orders: orders.pending_orders,
        total_discount: orders.total_discount,
        total_tax: orders.total_tax,
        gross_profit,
        gross_profit_margin: rules::gross_margin(orders.total_revenue, gross_profit),
    };
    Ok(success(
        StatusCode::OK,
        "Report summary retrieved successfully",
        body,
    ))
}

#[derive(Serialize)]
struct SalesTrendItem {
    /// Go scans the date into a string, which formats it as a midnight UTC time.
    date: String,
    #[serde(serialize_with = "crate::money::serialize")]
    sales: Decimal,
    orders: i64,
    items: i64,
}

async fn sales_trend(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<List<SalesTrendItem>>> {
    let r = range(&user, &query)?;
    let data = db::sales_trend(&state.db, r)
        .await?
        .into_iter()
        .map(|row| SalesTrendItem {
            date: row.date.format("%Y-%m-%dT00:00:00Z").to_string(),
            sales: row.sales,
            orders: row.orders,
            items: row.items,
        })
        .collect();
    Ok(success(
        StatusCode::OK,
        "Sales trend retrieved successfully",
        List { data },
    ))
}

#[derive(Serialize)]
struct TopProductItem {
    product_id: i64,
    product_name: String,
    total_sold: i64,
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
}

async fn top_products(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<List<TopProductItem>>> {
    let r = range(&user, &query)?;
    let limit = query
        .int("limit")
        .filter(|l| *l > 0)
        .unwrap_or(DEFAULT_TOP_PRODUCTS);
    let data = db::top_products(&state.db, r, limit)
        .await?
        .into_iter()
        .map(|row| TopProductItem {
            product_id: row.product_id,
            product_name: row.product_name,
            total_sold: row.total_sold,
            total_amount: row.total_amount,
        })
        .collect();
    Ok(success(
        StatusCode::OK,
        "Top products retrieved successfully",
        List { data },
    ))
}

#[derive(Serialize)]
struct CategoryRevenueItem {
    category_id: i64,
    category_name: String,
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
    order_count: i64,
}

async fn category_revenue(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<List<CategoryRevenueItem>>> {
    let r = range(&user, &query)?;
    let data = db::category_revenue(&state.db, r)
        .await?
        .into_iter()
        .map(|row| CategoryRevenueItem {
            category_id: row.category_id,
            category_name: row.category_name,
            total_amount: row.total_amount,
            order_count: row.order_count,
        })
        .collect();
    Ok(success(
        StatusCode::OK,
        "Category revenue retrieved successfully",
        List { data },
    ))
}

#[derive(Serialize)]
struct PaymentMethodItem {
    method: String,
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
    count: i64,
}

async fn payment_methods(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<List<PaymentMethodItem>>> {
    let r = range(&user, &query)?;
    let data = db::payment_methods(&state.db, r)
        .await?
        .into_iter()
        .map(|row| PaymentMethodItem {
            method: row.method,
            total_amount: row.total_amount,
            count: row.count,
        })
        .collect();
    Ok(success(
        StatusCode::OK,
        "Payment methods retrieved successfully",
        List { data },
    ))
}

#[derive(Serialize)]
struct HourlySalesItem {
    hour: i32,
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
    order_count: i64,
}

async fn hourly_sales(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<List<HourlySalesItem>>> {
    let r = range(&user, &query)?;
    let data = db::hourly_sales(&state.db, r)
        .await?
        .into_iter()
        .map(|row| HourlySalesItem {
            hour: row.hour,
            total_amount: row.total_amount,
            order_count: row.order_count,
        })
        .collect();
    Ok(success(
        StatusCode::OK,
        "Hourly sales retrieved successfully",
        List { data },
    ))
}
