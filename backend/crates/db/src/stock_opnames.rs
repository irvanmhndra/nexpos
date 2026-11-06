//! `stock_opnames` and `stock_opname_items`, like the Go repository.

use std::collections::HashMap;

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct StockOpname {
    pub id: i64,
    pub company_id: i64,
    pub branch_id: i64,
    pub opname_number: String,
    pub status: String,
    pub notes: Option<String>,
    pub total_variance_qty: i32,
    pub total_variance_value: Decimal,
    pub started_by: Option<i64>,
    pub completed_by: Option<i64>,
    pub started_at: DateTime<Utc>,
    pub completed_at: Option<DateTime<Utc>>,
    pub cancelled_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct StockOpnameItem {
    pub id: i64,
    pub stock_opname_id: i64,
    pub product_variant_id: i64,
    pub sku: String,
    pub product_name: String,
    pub variant_name: String,
    pub system_stock: i32,
    pub counted_stock: Option<i32>,
    pub variance_qty: i32,
    pub unit_cost: Decimal,
    pub variance_value: Decimal,
    pub notes: Option<String>,
    pub counted_at: Option<DateTime<Utc>>,
}

const COLUMNS: &str = "id, company_id, branch_id, opname_number, status, notes, total_variance_qty, \
     total_variance_value, started_by, completed_by, started_at, completed_at, cancelled_at, created_at, updated_at";

const ITEM_COLUMNS: &str = "id, stock_opname_id, product_variant_id, sku, product_name, variant_name, \
     system_stock, counted_stock, variance_qty, unit_cost, variance_value, notes, counted_at";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<StockOpname>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM stock_opnames WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// Locks the row: completing twice at once must not apply the count twice.
pub async fn lock(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<StockOpname>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM stock_opnames WHERE id = $1 AND company_id = $2 FOR UPDATE"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// `OPN-<prefix>-NNNN`: the company's count for the day, plus one.
pub async fn next_number(
    db: impl PgExecutor<'_>,
    company_id: i64,
    date_prefix: &str,
) -> sqlx::Result<String> {
    let prefix = format!("OPN-{date_prefix}-");
    let count: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM stock_opnames WHERE company_id = $1 AND opname_number LIKE $2",
    )
    .bind(company_id)
    .bind(format!("{prefix}%"))
    .fetch_one(db)
    .await?;
    Ok(format!("{prefix}{:04}", count + 1))
}

pub async fn list(
    db: &PgPool,
    company_id: i64,
    status: &str,
    branch_id: Option<i64>,
    limit: i64,
    offset: i64,
) -> sqlx::Result<(Vec<StockOpname>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !status.is_empty() {
            q.push(" AND status = ").push_bind(status.to_owned());
        }
        if let Some(branch) = branch_id {
            q.push(" AND branch_id = ").push_bind(branch);
        }
    };
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM stock_opnames");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM stock_opnames"));
    filter(&mut rows);
    rows.push(" ORDER BY created_at DESC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

pub async fn create(
    db: impl PgExecutor<'_>,
    company_id: i64,
    branch_id: i64,
    number: &str,
    notes: Option<&str>,
    started_by: i64,
    started_at: DateTime<Utc>,
) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "INSERT INTO stock_opnames (company_id, branch_id, opname_number, status, notes, started_by, started_at) \
         VALUES ($1, $2, $3, 'in_progress', $4, $5, $6) RETURNING id",
    )
    .bind(company_id)
    .bind(branch_id)
    .bind(number)
    .bind(notes)
    .bind(started_by)
    .bind(started_at)
    .fetch_one(db)
    .await
}

pub async fn update(db: impl PgExecutor<'_>, o: &StockOpname) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE stock_opnames SET status = $1, notes = $2, total_variance_qty = $3, total_variance_value = $4, \
             completed_by = $5, completed_at = $6, cancelled_at = $7, updated_at = NOW() \
         WHERE id = $8 AND company_id = $9",
    )
    .bind(&o.status)
    .bind(&o.notes)
    .bind(o.total_variance_qty)
    .bind(o.total_variance_value)
    .bind(o.completed_by)
    .bind(o.completed_at)
    .bind(o.cancelled_at)
    .bind(o.id)
    .bind(o.company_id)
    .execute(db)
    .await?;
    Ok(())
}

/// One item per active variant, with the stock and cost at this moment. Returns
/// how many were taken.
pub async fn snapshot_items(
    db: impl PgExecutor<'_>,
    opname_id: i64,
    company_id: i64,
    branch_id: i64,
    category_id: Option<i64>,
) -> sqlx::Result<u64> {
    let result = sqlx::query(
        "INSERT INTO stock_opname_items (stock_opname_id, product_variant_id, sku, product_name, variant_name, system_stock, unit_cost) \
         SELECT $1, pv.id, pv.sku, p.name, pv.name, COALESCE(s.quantity, 0), \
             COALESCE(NULLIF(pv.last_purchase_cost, 0), pv.standard_cost) \
         FROM product_variants pv JOIN products p ON p.id = pv.product_id \
         LEFT JOIN stocks s ON s.product_variant_id = pv.id AND s.branch_id = $3 \
         WHERE p.company_id = $2 AND p.is_active = true AND pv.is_active = true \
           AND ($4::bigint IS NULL OR p.product_category_id = $4)",
    )
    .bind(opname_id)
    .bind(company_id)
    .bind(branch_id)
    .bind(category_id)
    .execute(db)
    .await?;
    Ok(result.rows_affected())
}

pub async fn items_of(
    db: impl PgExecutor<'_>,
    opname_id: i64,
) -> sqlx::Result<Vec<StockOpnameItem>> {
    sqlx::query_as(&format!(
        "SELECT {ITEM_COLUMNS} FROM stock_opname_items WHERE stock_opname_id = $1 \
         ORDER BY product_name, variant_name, id"
    ))
    .bind(opname_id)
    .fetch_all(db)
    .await
}

pub async fn item(
    db: impl PgExecutor<'_>,
    opname_id: i64,
    item_id: i64,
) -> sqlx::Result<Option<StockOpnameItem>> {
    sqlx::query_as(&format!(
        "SELECT {ITEM_COLUMNS} FROM stock_opname_items WHERE id = $1 AND stock_opname_id = $2"
    ))
    .bind(item_id)
    .bind(opname_id)
    .fetch_optional(db)
    .await
}

pub async fn update_item(
    db: impl PgExecutor<'_>,
    i: &StockOpnameItem,
    counted_by: i64,
) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE stock_opname_items SET counted_stock = $1, variance_qty = $2, variance_value = $3, \
             notes = $4, counted_by = $5, counted_at = $6, updated_at = NOW() \
         WHERE id = $7",
    )
    .bind(i.counted_stock)
    .bind(i.variance_qty)
    .bind(i.variance_value)
    .bind(&i.notes)
    .bind(counted_by)
    .bind(i.counted_at)
    .bind(i.id)
    .execute(db)
    .await?;
    Ok(())
}

/// `(total, counted)` per opname.
pub async fn item_stats(
    db: impl PgExecutor<'_>,
    opname_ids: &[i64],
) -> sqlx::Result<HashMap<i64, (i64, i64)>> {
    let rows: Vec<(i64, i64, i64)> = sqlx::query_as(
        "SELECT stock_opname_id, COUNT(*), COUNT(*) FILTER (WHERE counted_stock IS NOT NULL) \
         FROM stock_opname_items WHERE stock_opname_id = ANY($1) GROUP BY stock_opname_id",
    )
    .bind(opname_ids)
    .fetch_all(db)
    .await?;
    Ok(rows
        .into_iter()
        .map(|(id, total, counted)| (id, (total, counted)))
        .collect())
}
