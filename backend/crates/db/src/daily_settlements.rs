//! `daily_settlements` and `daily_settlement_items`, like the Go repository.
//! Dates are bound as text and cast, so an unparseable one is an error from Postgres.

use std::collections::HashMap;

use chrono::{DateTime, NaiveDate, Utc};
use rust_decimal::Decimal;
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct DailySettlement {
    pub id: i64,
    pub company_id: i64,
    pub branch_id: i64,
    pub settlement_date: NaiveDate,
    pub status: String,
    pub total_sales: Decimal,
    pub total_refunds: Decimal,
    pub total_expenses: Decimal,
    pub notes: Option<String>,
    pub recorded_by: Option<i64>,
    pub finalized_by: Option<i64>,
    pub finalized_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct DailySettlementItem {
    pub id: i64,
    pub daily_settlement_id: i64,
    pub payment_method: String,
    pub expected_amount: Decimal,
    pub actual_amount: Option<Decimal>,
    pub variance_amount: Decimal,
    pub notes: Option<String>,
}

#[derive(Debug, sqlx::FromRow)]
pub struct MethodTotals {
    pub method: String,
    pub gross_sales: Decimal,
    pub refunds: Decimal,
}

pub struct NewSettlement<'a> {
    pub company_id: i64,
    pub branch_id: i64,
    pub settlement_date: &'a str,
    pub total_sales: Decimal,
    pub total_refunds: Decimal,
    pub total_expenses: Decimal,
    pub notes: Option<&'a str>,
    pub recorded_by: i64,
}

pub struct ListFilter<'a> {
    pub status: &'a str,
    pub branch_id: Option<i64>,
    pub date_from: &'a str,
    pub date_to: &'a str,
    pub limit: i64,
    pub offset: i64,
}

const COLUMNS: &str = "id, company_id, branch_id, settlement_date, status, total_sales, total_refunds, \
     total_expenses, notes, recorded_by, finalized_by, finalized_at, created_at, updated_at";

const ITEM_COLUMNS: &str = "id, daily_settlement_id, payment_method, expected_amount, actual_amount, variance_amount, notes";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<DailySettlement>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM daily_settlements WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn exists_for_date(
    db: impl PgExecutor<'_>,
    company_id: i64,
    branch_id: i64,
    date: &str,
) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS (SELECT 1 FROM daily_settlements \
         WHERE company_id = $1 AND branch_id = $2 AND settlement_date = $3::date)",
    )
    .bind(company_id)
    .bind(branch_id)
    .bind(date)
    .fetch_one(db)
    .await
}

pub async fn list(
    db: &PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<DailySettlement>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !f.status.is_empty() {
            q.push(" AND status = ").push_bind(f.status.to_owned());
        }
        if let Some(branch) = f.branch_id {
            q.push(" AND branch_id = ").push_bind(branch);
        }
        if !f.date_from.is_empty() {
            q.push(" AND settlement_date >= ")
                .push_bind(f.date_from.to_owned())
                .push("::date");
        }
        if !f.date_to.is_empty() {
            q.push(" AND settlement_date <= ")
                .push_bind(f.date_to.to_owned())
                .push("::date");
        }
    };
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM daily_settlements");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM daily_settlements"));
    filter(&mut rows);
    rows.push(" ORDER BY settlement_date DESC, id DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

pub async fn create(db: impl PgExecutor<'_>, s: &NewSettlement<'_>) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "INSERT INTO daily_settlements (company_id, branch_id, settlement_date, status, total_sales, \
             total_refunds, total_expenses, notes, recorded_by) \
         VALUES ($1, $2, $3::date, 'draft', $4, $5, $6, $7, $8) RETURNING id",
    )
    .bind(s.company_id)
    .bind(s.branch_id)
    .bind(s.settlement_date)
    .bind(s.total_sales)
    .bind(s.total_refunds)
    .bind(s.total_expenses)
    .bind(s.notes)
    .bind(s.recorded_by)
    .fetch_one(db)
    .await
}

pub async fn finalize(db: impl PgExecutor<'_>, s: &DailySettlement) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE daily_settlements SET status = $1, notes = $2, finalized_by = $3, finalized_at = $4, \
             updated_at = NOW() \
         WHERE id = $5 AND company_id = $6",
    )
    .bind(&s.status)
    .bind(&s.notes)
    .bind(s.finalized_by)
    .bind(s.finalized_at)
    .bind(s.id)
    .bind(s.company_id)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn create_item(
    db: impl PgExecutor<'_>,
    settlement_id: i64,
    method: &str,
    expected: Decimal,
    variance: Decimal,
) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO daily_settlement_items (daily_settlement_id, payment_method, expected_amount, variance_amount) \
         VALUES ($1, $2, $3, $4)",
    )
    .bind(settlement_id)
    .bind(method)
    .bind(expected)
    .bind(variance)
    .execute(db)
    .await?;
    Ok(())
}

/// Items of many settlements, each list in id order.
pub async fn items_by_settlement(
    db: impl PgExecutor<'_>,
    ids: &[i64],
) -> sqlx::Result<HashMap<i64, Vec<DailySettlementItem>>> {
    let rows: Vec<DailySettlementItem> = sqlx::query_as(&format!(
        "SELECT {ITEM_COLUMNS} FROM daily_settlement_items WHERE daily_settlement_id = ANY($1) ORDER BY id"
    ))
    .bind(ids)
    .fetch_all(db)
    .await?;
    let mut by_settlement: HashMap<i64, Vec<DailySettlementItem>> = HashMap::new();
    for row in rows {
        by_settlement
            .entry(row.daily_settlement_id)
            .or_default()
            .push(row);
    }
    Ok(by_settlement)
}

pub async fn item(
    db: impl PgExecutor<'_>,
    settlement_id: i64,
    item_id: i64,
) -> sqlx::Result<Option<DailySettlementItem>> {
    sqlx::query_as(&format!(
        "SELECT {ITEM_COLUMNS} FROM daily_settlement_items WHERE id = $1 AND daily_settlement_id = $2"
    ))
    .bind(item_id)
    .bind(settlement_id)
    .fetch_optional(db)
    .await
}

pub async fn update_item(db: impl PgExecutor<'_>, i: &DailySettlementItem) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE daily_settlement_items SET actual_amount = $1, variance_amount = $2, notes = $3, updated_at = NOW() \
         WHERE id = $4",
    )
    .bind(i.actual_amount)
    .bind(i.variance_amount)
    .bind(&i.notes)
    .bind(i.id)
    .execute(db)
    .await?;
    Ok(())
}

/// Sales and refunds per method: a sale counts on the day it was paid, a refund
/// on the day it was issued.
pub async fn payment_breakdown(
    db: impl PgExecutor<'_>,
    company_id: i64,
    branch_id: i64,
    date: &str,
) -> sqlx::Result<Vec<MethodTotals>> {
    sqlx::query_as(
        "SELECT method, COALESCE(SUM(gross_sales), 0) AS gross_sales, COALESCE(SUM(refunds), 0) AS refunds \
         FROM ( \
             SELECT p.method, p.amount AS gross_sales, 0::numeric AS refunds \
             FROM payments p JOIN orders o ON o.id = p.order_id \
             WHERE o.company_id = $1 AND o.branch_id = $2 AND p.status != 'failed' \
               AND p.paid_at::date = $3::date \
             UNION ALL \
             SELECT p.method, 0::numeric AS gross_sales, p.refunded_amount AS refunds \
             FROM payments p JOIN orders o ON o.id = p.order_id \
             WHERE o.company_id = $1 AND o.branch_id = $2 AND p.refunded_at IS NOT NULL \
               AND p.refunded_at::date = $3::date \
         ) t GROUP BY method ORDER BY method",
    )
    .bind(company_id)
    .bind(branch_id)
    .bind(date)
    .fetch_all(db)
    .await
}

pub async fn expenses_total(
    db: impl PgExecutor<'_>,
    company_id: i64,
    branch_id: i64,
    date: &str,
) -> sqlx::Result<Decimal> {
    sqlx::query_scalar(
        "SELECT COALESCE(SUM(amount), 0) FROM expenses \
         WHERE company_id = $1 AND branch_id = $2 AND expense_date = $3::date",
    )
    .bind(company_id)
    .bind(branch_id)
    .bind(date)
    .fetch_one(db)
    .await
}
