//! `stocks` and `stock_movements`, like the Go repositories. Every change to a
//! quantity locks its row first, then records a movement.

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::{PgConnection, PgExecutor, PgPool, Postgres, QueryBuilder};

#[derive(Debug, Clone, Copy, sqlx::FromRow)]
pub struct StockLevel {
    pub quantity: i32,
    pub min_quantity: i32,
}

pub struct NewMovement<'a> {
    pub variant_id: i64,
    pub branch_id: i64,
    pub kind: &'a str,
    pub quantity: i64,
    pub stock_before: i64,
    pub stock_after: i64,
    pub unit_cost: Option<Decimal>,
    pub reference_type: Option<&'a str>,
    pub reference_id: Option<i64>,
    pub note: Option<&'a str>,
    pub created_by: Option<i64>,
}

/// What's on hand; a variant never stocked at the branch has none.
pub async fn on_hand(
    db: impl PgExecutor<'_>,
    variant_id: i64,
    branch_id: i64,
) -> sqlx::Result<i32> {
    let quantity: Option<i32> = sqlx::query_scalar(
        "SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
    )
    .bind(variant_id)
    .bind(branch_id)
    .fetch_optional(db)
    .await?;
    Ok(quantity.unwrap_or(0))
}

/// Locks the stock row until the transaction ends, creating it at zero if needed.
pub async fn lock(
    conn: &mut PgConnection,
    variant_id: i64,
    branch_id: i64,
) -> sqlx::Result<StockLevel> {
    sqlx::query(
        "INSERT INTO stocks (product_variant_id, branch_id, quantity, min_quantity, updated_at) \
         VALUES ($1, $2, 0, 0, NOW()) ON CONFLICT (product_variant_id, branch_id) DO NOTHING",
    )
    .bind(variant_id)
    .bind(branch_id)
    .execute(&mut *conn)
    .await?;
    sqlx::query_as(
        "SELECT quantity, min_quantity FROM stocks \
         WHERE product_variant_id = $1 AND branch_id = $2 FOR UPDATE",
    )
    .bind(variant_id)
    .bind(branch_id)
    .fetch_one(&mut *conn)
    .await
}

pub async fn set_quantity(
    db: impl PgExecutor<'_>,
    variant_id: i64,
    branch_id: i64,
    quantity: i64,
) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE stocks SET quantity = $3, updated_at = NOW() \
         WHERE product_variant_id = $1 AND branch_id = $2",
    )
    .bind(variant_id)
    .bind(branch_id)
    .bind(quantity)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn set_min_quantity(
    db: impl PgExecutor<'_>,
    variant_id: i64,
    branch_id: i64,
    min_quantity: i64,
) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO stocks (product_variant_id, branch_id, quantity, min_quantity, updated_at) \
         VALUES ($1, $2, 0, $3, NOW()) \
         ON CONFLICT (product_variant_id, branch_id) DO UPDATE SET min_quantity = $3, updated_at = NOW()",
    )
    .bind(variant_id)
    .bind(branch_id)
    .bind(min_quantity)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn record_movement(db: impl PgExecutor<'_>, m: &NewMovement<'_>) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO stock_movements (product_variant_id, branch_id, type, quantity, stock_before, \
             stock_after, unit_cost, reference_type, reference_id, note, created_by) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
    )
    .bind(m.variant_id)
    .bind(m.branch_id)
    .bind(m.kind)
    .bind(m.quantity)
    .bind(m.stock_before)
    .bind(m.stock_after)
    .bind(m.unit_cost)
    .bind(m.reference_type)
    .bind(m.reference_id)
    .bind(m.note)
    .bind(m.created_by)
    .execute(db)
    .await?;
    Ok(())
}

// ============== Inventory ==============

#[derive(Debug, sqlx::FromRow)]
pub struct InventoryRow {
    pub product_variant_id: i64,
    pub product_id: i64,
    pub product_name: String,
    pub variant_name: String,
    pub sku: String,
    pub category_name: String,
    pub image_data: Option<String>,
    pub branch_id: i64,
    pub branch_name: String,
    pub current_stock: i32,
    pub min_stock: i32,
    pub stock_value: Decimal,
    pub has_variants: bool,
}

#[derive(Debug, sqlx::FromRow)]
pub struct InventoryStats {
    pub total_sku: i64,
    pub out_of_stock: i64,
    pub low_stock: i64,
    pub total_stock_value: Decimal,
}

pub struct InventoryFilter<'a> {
    pub branch_id: i64,
    pub search: &'a str,
    pub category: &'a str,
    pub status: &'a str,
    pub limit: i64,
    pub offset: i64,
}

/// Every active variant. The branch goes in the joins, so a variant without a
/// stock row still shows, at zero; without a branch nothing joins.
pub async fn list_inventory(
    db: &PgPool,
    company_id: i64,
    f: &InventoryFilter<'_>,
) -> sqlx::Result<(Vec<InventoryRow>, i64)> {
    let from = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(
            " FROM product_variants pv JOIN products p ON p.id = pv.product_id \
              LEFT JOIN product_categories pc ON pc.id = p.product_category_id",
        );
        if f.branch_id > 0 {
            q.push(" LEFT JOIN stocks s ON s.product_variant_id = pv.id AND s.branch_id = ")
                .push_bind(f.branch_id)
                .push(" LEFT JOIN branches b ON b.company_id = ")
                .push_bind(company_id)
                .push(" AND b.id = ")
                .push_bind(f.branch_id);
        } else {
            q.push(" LEFT JOIN stocks s ON false LEFT JOIN branches b ON false");
        }
        q.push(" WHERE p.company_id = ")
            .push_bind(company_id)
            .push(" AND pv.is_active = true AND p.is_active = true");
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (p.name ILIKE ")
                .push_bind(like.clone())
                .push(" OR pv.sku ILIKE ")
                .push_bind(like)
                .push(")");
        }
        if !f.category.is_empty() {
            q.push(" AND pc.name = ").push_bind(f.category.to_owned());
        }
        match f.status {
            "out_of_stock" => {
                q.push(" AND COALESCE(s.quantity, 0) <= 0");
            }
            "low" => {
                q.push(
                    " AND COALESCE(s.quantity, 0) > 0 AND COALESCE(s.quantity, 0) <= COALESCE(s.min_quantity, 0)",
                );
            }
            "normal" => {
                q.push(" AND COALESCE(s.quantity, 0) > COALESCE(s.min_quantity, 0)");
            }
            _ => {}
        }
    };

    let mut count = QueryBuilder::new("SELECT COUNT(*)");
    from(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(
        "SELECT pv.id AS product_variant_id, p.id AS product_id, p.name AS product_name, \
             pv.name AS variant_name, pv.sku, COALESCE(pc.name, '') AS category_name, p.image_data, \
             COALESCE(b.id, 0) AS branch_id, COALESCE(b.name, '') AS branch_name, \
             COALESCE(s.quantity, 0) AS current_stock, COALESCE(s.min_quantity, 0) AS min_stock, \
             GREATEST(COALESCE(s.quantity, 0), 0) * pv.standard_cost AS stock_value, \
             (SELECT COUNT(*) FROM product_variants pv2 WHERE pv2.product_id = p.id AND pv2.is_active = true) > 1 AS has_variants",
    );
    from(&mut rows);
    rows.push(" ORDER BY p.name, pv.name LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

pub async fn inventory_stats(
    db: &PgPool,
    company_id: i64,
    branch_id: i64,
) -> sqlx::Result<InventoryStats> {
    let mut q = QueryBuilder::<Postgres>::new(
        "SELECT COUNT(*) AS total_sku, \
             COUNT(*) FILTER (WHERE COALESCE(s.quantity,0) <= 0) AS out_of_stock, \
             COUNT(*) FILTER (WHERE COALESCE(s.quantity,0) > 0 AND COALESCE(s.quantity,0) <= COALESCE(s.min_quantity,0)) AS low_stock, \
             COALESCE(SUM(GREATEST(COALESCE(s.quantity,0), 0) * pv.standard_cost), 0) AS total_stock_value \
         FROM product_variants pv JOIN products p ON p.id = pv.product_id",
    );
    if branch_id > 0 {
        q.push(" LEFT JOIN stocks s ON s.product_variant_id = pv.id AND s.branch_id = ")
            .push_bind(branch_id);
    } else {
        q.push(" LEFT JOIN stocks s ON false");
    }
    q.push(" WHERE p.company_id = ")
        .push_bind(company_id)
        .push(" AND pv.is_active = true AND p.is_active = true");
    q.build_query_as().fetch_one(db).await
}

// ============== Movements ==============

#[derive(Debug, sqlx::FromRow)]
pub struct MovementRow {
    pub id: i64,
    pub product_variant_id: i64,
    pub product_name: String,
    pub variant_name: String,
    pub sku: String,
    pub branch_id: i64,
    pub branch_name: String,
    #[sqlx(rename = "type")]
    pub kind: String,
    pub quantity: i32,
    pub stock_before: i32,
    pub stock_after: i32,
    pub unit_cost: Option<Decimal>,
    pub reference_type: Option<String>,
    pub reference_id: Option<i64>,
    pub note: Option<String>,
    pub created_by: Option<i64>,
    pub created_by_name: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, sqlx::FromRow)]
pub struct MovementStats {
    pub total_movements: i64,
    pub stock_in: i64,
    pub stock_out: i64,
    pub adjustments: i64,
}

pub struct MovementFilter<'a> {
    pub branch_id: i64,
    pub kind: &'a str,
    pub search: &'a str,
    pub start_date: &'a str,
    pub end_date: &'a str,
    pub limit: i64,
    pub offset: i64,
}

pub async fn list_movements(
    db: &PgPool,
    company_id: i64,
    f: &MovementFilter<'_>,
) -> sqlx::Result<(Vec<MovementRow>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE p.company_id = ").push_bind(company_id);
        if f.branch_id > 0 {
            q.push(" AND sm.branch_id = ").push_bind(f.branch_id);
        }
        if !f.kind.is_empty() {
            q.push(" AND sm.type = ").push_bind(f.kind.to_owned());
        }
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (p.name ILIKE ")
                .push_bind(like.clone())
                .push(" OR pv.sku ILIKE ")
                .push_bind(like)
                .push(")");
        }
        if !f.start_date.is_empty() {
            q.push(" AND sm.created_at >= ")
                .push_bind(f.start_date.to_owned())
                .push("::timestamptz");
        }
        if !f.end_date.is_empty() {
            q.push(" AND sm.created_at < ")
                .push_bind(f.end_date.to_owned())
                .push("::timestamptz");
        }
    };

    let mut count = QueryBuilder::new(
        "SELECT COUNT(*) FROM stock_movements sm \
         JOIN product_variants pv ON pv.id = sm.product_variant_id \
         JOIN products p ON p.id = pv.product_id",
    );
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(
        "SELECT sm.id, sm.product_variant_id, p.name AS product_name, pv.name AS variant_name, pv.sku, \
             sm.branch_id, b.name AS branch_name, sm.type, sm.quantity, sm.stock_before, sm.stock_after, \
             sm.unit_cost, sm.reference_type, sm.reference_id, sm.note, sm.created_by, \
             COALESCE(u.name, '') AS created_by_name, sm.created_at \
         FROM stock_movements sm \
         JOIN product_variants pv ON pv.id = sm.product_variant_id \
         JOIN products p ON p.id = pv.product_id \
         JOIN branches b ON b.id = sm.branch_id \
         LEFT JOIN users u ON u.id = sm.created_by",
    );
    filter(&mut rows);
    rows.push(" ORDER BY sm.created_at DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

/// Movements since `since` (the start of the month).
pub async fn movement_stats(
    db: &PgPool,
    company_id: i64,
    branch_id: i64,
    since: DateTime<Utc>,
) -> sqlx::Result<MovementStats> {
    let mut q = QueryBuilder::<Postgres>::new(
        "SELECT COUNT(*) AS total_movements, \
             COUNT(*) FILTER (WHERE sm.type = 'IN') AS stock_in, \
             COUNT(*) FILTER (WHERE sm.type = 'OUT') AS stock_out, \
             COUNT(*) FILTER (WHERE sm.type = 'ADJUST') AS adjustments \
         FROM stock_movements sm \
         JOIN product_variants pv ON pv.id = sm.product_variant_id \
         JOIN products p ON p.id = pv.product_id WHERE p.company_id = ",
    );
    q.push_bind(company_id)
        .push(" AND sm.created_at >= ")
        .push_bind(since);
    if branch_id > 0 {
        q.push(" AND sm.branch_id = ").push_bind(branch_id);
    }
    q.build_query_as().fetch_one(db).await
}
