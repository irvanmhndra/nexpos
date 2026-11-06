//! `purchase_orders` and `purchase_order_items`, like the Go repository.

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct PurchaseOrder {
    pub id: i64,
    pub company_id: i64,
    pub branch_id: i64,
    pub supplier_id: i64,
    pub po_number: String,
    pub status: String,
    pub notes: Option<String>,
    pub total_amount: Decimal,
    pub ordered_at: Option<DateTime<Utc>>,
    pub received_at: Option<DateTime<Utc>>,
    pub cancelled_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct PurchaseOrderItem {
    pub id: i64,
    pub product_variant_id: Option<i64>,
    pub sku: String,
    pub variant_name: String,
    pub quantity: i32,
    pub received_quantity: i32,
    pub unit_cost: Decimal,
    pub subtotal: Decimal,
}

pub struct NewItem<'a> {
    pub purchase_order_id: i64,
    pub product_variant_id: Option<i64>,
    pub sku: &'a str,
    pub variant_name: &'a str,
    pub quantity: i64,
    pub unit_cost: Decimal,
    pub subtotal: Decimal,
}

pub struct NewPurchaseOrder<'a> {
    pub company_id: i64,
    pub branch_id: i64,
    pub supplier_id: i64,
    pub po_number: &'a str,
    pub notes: Option<&'a str>,
    pub total_amount: Decimal,
    pub ordered_at: DateTime<Utc>,
}

const COLUMNS: &str = "id, company_id, branch_id, supplier_id, po_number, status, notes, total_amount, \
     ordered_at, received_at, cancelled_at, created_at, updated_at";

const ITEM_COLUMNS: &str =
    "id, product_variant_id, sku, variant_name, quantity, received_quantity, unit_cost, subtotal";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<PurchaseOrder>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM purchase_orders WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// Locks the row: a purchase order received twice at once must not double-count.
pub async fn lock(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<PurchaseOrder>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM purchase_orders WHERE id = $1 AND company_id = $2 FOR UPDATE"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// `PO-<prefix>-NNNN`: the company's count for the day, plus one.
pub async fn next_number(
    db: impl PgExecutor<'_>,
    company_id: i64,
    date_prefix: &str,
) -> sqlx::Result<String> {
    let prefix = format!("PO-{date_prefix}-");
    let count: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM purchase_orders WHERE company_id = $1 AND po_number LIKE $2",
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
    supplier_id: Option<i64>,
    limit: i64,
    offset: i64,
) -> sqlx::Result<(Vec<PurchaseOrder>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !status.is_empty() {
            q.push(" AND status = ").push_bind(status.to_owned());
        }
        if let Some(supplier) = supplier_id {
            q.push(" AND supplier_id = ").push_bind(supplier);
        }
    };
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM purchase_orders");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM purchase_orders"));
    filter(&mut rows);
    rows.push(" ORDER BY created_at DESC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

pub async fn create(
    db: impl PgExecutor<'_>,
    po: &NewPurchaseOrder<'_>,
) -> sqlx::Result<PurchaseOrder> {
    sqlx::query_as(&format!(
        "INSERT INTO purchase_orders (company_id, branch_id, supplier_id, po_number, status, notes, total_amount, ordered_at) \
         VALUES ($1, $2, $3, $4, 'draft', $5, $6, $7) RETURNING {COLUMNS}"
    ))
    .bind(po.company_id)
    .bind(po.branch_id)
    .bind(po.supplier_id)
    .bind(po.po_number)
    .bind(po.notes)
    .bind(po.total_amount)
    .bind(po.ordered_at)
    .fetch_one(db)
    .await
}

pub async fn create_item(db: impl PgExecutor<'_>, i: &NewItem<'_>) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "INSERT INTO purchase_order_items (purchase_order_id, product_variant_id, sku, variant_name, \
             quantity, received_quantity, unit_cost, subtotal) \
         VALUES ($1, $2, $3, $4, $5, 0, $6, $7) RETURNING id",
    )
    .bind(i.purchase_order_id)
    .bind(i.product_variant_id)
    .bind(i.sku)
    .bind(i.variant_name)
    .bind(i.quantity)
    .bind(i.unit_cost)
    .bind(i.subtotal)
    .fetch_one(db)
    .await
}

pub async fn items_of(db: impl PgExecutor<'_>, id: i64) -> sqlx::Result<Vec<PurchaseOrderItem>> {
    sqlx::query_as(&format!(
        "SELECT {ITEM_COLUMNS} FROM purchase_order_items WHERE purchase_order_id = $1 ORDER BY id"
    ))
    .bind(id)
    .fetch_all(db)
    .await
}

pub async fn set_received(
    db: impl PgExecutor<'_>,
    item_id: i64,
    received_quantity: i64,
) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE purchase_order_items SET received_quantity = $1, updated_at = NOW() WHERE id = $2",
    )
    .bind(received_quantity)
    .bind(item_id)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn set_status(db: impl PgExecutor<'_>, po: &PurchaseOrder) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE purchase_orders SET status = $1, received_at = $2, updated_at = NOW() \
         WHERE id = $3 AND company_id = $4",
    )
    .bind(&po.status)
    .bind(po.received_at)
    .bind(po.id)
    .bind(po.company_id)
    .execute(db)
    .await?;
    Ok(())
}
