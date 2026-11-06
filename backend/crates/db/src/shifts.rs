//! `shifts`, like the Go repository.

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Shift {
    pub id: i64,
    pub company_id: i64,
    pub branch_id: i64,
    pub cashier_id: i64,
    pub status: String,
    pub opening_float: Decimal,
    pub closing_float: Decimal,
    pub expected_cash: Decimal,
    pub actual_cash: Option<Decimal>,
    pub cash_difference: Option<Decimal>,
    pub notes: Option<String>,
    pub opened_at: DateTime<Utc>,
    pub closed_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub struct ListFilter<'a> {
    pub branch_id: Option<i64>,
    pub cashier_id: Option<i64>,
    pub status: &'a str,
    pub date_from: &'a str,
    pub date_to: &'a str,
    pub limit: i64,
    pub offset: i64,
}

const COLUMNS: &str = "id, company_id, branch_id, cashier_id, status, opening_float, closing_float, \
     expected_cash, actual_cash, cash_difference, notes, opened_at, closed_at, created_at, updated_at";

pub async fn get(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<Option<Shift>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM shifts WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn open_shift(
    db: impl PgExecutor<'_>,
    company_id: i64,
    branch_id: i64,
    cashier_id: i64,
) -> sqlx::Result<Option<Shift>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM shifts \
         WHERE company_id = $1 AND branch_id = $2 AND cashier_id = $3 AND status = 'open' \
         ORDER BY opened_at DESC LIMIT 1"
    ))
    .bind(company_id)
    .bind(branch_id)
    .bind(cashier_id)
    .fetch_optional(db)
    .await
}

pub async fn create(
    db: impl PgExecutor<'_>,
    company_id: i64,
    branch_id: i64,
    cashier_id: i64,
    opening_float: Decimal,
    opened_at: DateTime<Utc>,
) -> sqlx::Result<Shift> {
    sqlx::query_as(&format!(
        "INSERT INTO shifts (company_id, branch_id, cashier_id, status, opening_float, closing_float, expected_cash, opened_at) \
         VALUES ($1, $2, $3, 'open', $4, 0, 0, $5) RETURNING {COLUMNS}"
    ))
    .bind(company_id)
    .bind(branch_id)
    .bind(cashier_id)
    .bind(opening_float)
    .bind(opened_at)
    .fetch_one(db)
    .await
}

pub async fn close(db: impl PgExecutor<'_>, s: &Shift) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE shifts SET status = $1, closing_float = $2, expected_cash = $3, actual_cash = $4, \
             cash_difference = $5, notes = $6, closed_at = $7, updated_at = NOW() \
         WHERE id = $8 AND company_id = $9",
    )
    .bind(&s.status)
    .bind(s.closing_float)
    .bind(s.expected_cash)
    .bind(s.actual_cash)
    .bind(s.cash_difference)
    .bind(&s.notes)
    .bind(s.closed_at)
    .bind(s.id)
    .bind(s.company_id)
    .execute(db)
    .await?;
    Ok(())
}

/// Dates are bound as text and cast, so an unparseable one is an error from Postgres.
pub async fn list(
    db: &PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<Shift>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if let Some(branch) = f.branch_id {
            q.push(" AND branch_id = ").push_bind(branch);
        }
        if let Some(cashier) = f.cashier_id {
            q.push(" AND cashier_id = ").push_bind(cashier);
        }
        if !f.status.is_empty() {
            q.push(" AND status = ").push_bind(f.status.to_owned());
        }
        if !f.date_from.is_empty() {
            q.push(" AND opened_at >= ")
                .push_bind(f.date_from.to_owned())
                .push("::timestamptz");
        }
        if !f.date_to.is_empty() {
            q.push(" AND opened_at < ")
                .push_bind(f.date_to.to_owned())
                .push("::timestamptz");
        }
    };
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM shifts");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM shifts"));
    filter(&mut rows);
    rows.push(" ORDER BY opened_at DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}
