//! Reading `products` and `product_variants`, like the Go repositories. Variants
//! load in one query per page rather than one per product.

use std::collections::HashMap;

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use serde_json::Value;
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder, types::Json};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Product {
    pub id: i64,
    pub name: String,
    pub category_id: Option<i64>,
    pub category_name: Option<String>,
    pub description: Option<String>,
    pub image_data: Option<String>,
    pub image_url: Option<String>,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Variant {
    pub id: i64,
    pub product_id: i64,
    pub sku: String,
    pub barcode: Option<String>,
    pub name: String,
    pub attributes: Option<Json<Value>>,
    pub price: Decimal,
    pub standard_cost: Decimal,
    pub last_purchase_cost: Decimal,
    pub is_default: bool,
    pub is_active: bool,
    pub sale_price: Option<Decimal>,
    pub sale_start: Option<DateTime<Utc>>,
    pub sale_end: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// A variant found by a scanned code.
#[derive(Debug, sqlx::FromRow)]
pub struct CodeMatch {
    pub id: i64,
    pub product_id: i64,
    pub barcode: Option<String>,
}

pub struct ListFilter<'a> {
    pub search: &'a str,
    pub category_id: Option<i64>,
    pub is_active: Option<bool>,
    pub limit: i64,
    pub offset: i64,
}

const SELECT: &str = "SELECT p.id, p.name, p.product_category_id AS category_id, c.name AS category_name, \
     p.description, p.image_data, p.image_url, p.is_active, p.created_at, p.updated_at \
     FROM products p \
     LEFT JOIN product_categories c ON c.id = p.product_category_id AND c.company_id = p.company_id";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<Product>> {
    sqlx::query_as(&format!("{SELECT} WHERE p.id = $1 AND p.company_id = $2"))
        .bind(id)
        .bind(company_id)
        .fetch_optional(db)
        .await
}

pub async fn list(
    db: &PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<Product>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE p.company_id = ").push_bind(company_id);
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (p.name ILIKE ")
                .push_bind(like.clone())
                .push(" OR EXISTS (SELECT 1 FROM product_variants pv WHERE pv.product_id = p.id AND (pv.sku ILIKE ")
                .push_bind(like.clone())
                .push(" OR pv.barcode ILIKE ")
                .push_bind(like)
                .push(")))");
        }
        if let Some(category_id) = f.category_id {
            q.push(" AND p.product_category_id = ")
                .push_bind(category_id);
        }
        if let Some(active) = f.is_active {
            q.push(" AND p.is_active = ").push_bind(active);
        }
    };

    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM products p");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(SELECT);
    filter(&mut rows);
    rows.push(" ORDER BY p.created_at DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    let products = rows.build_query_as().fetch_all(db).await?;
    Ok((products, total))
}

/// The variants of these products, keyed by product and ordered default first, then by name.
pub async fn variants_of(
    db: impl PgExecutor<'_>,
    product_ids: &[i64],
) -> sqlx::Result<HashMap<i64, Vec<Variant>>> {
    let rows: Vec<Variant> = sqlx::query_as(
        "SELECT id, product_id, sku, barcode, name, attributes, price, standard_cost, last_purchase_cost, \
             is_default, is_active, sale_price, sale_start, sale_end, created_at, updated_at \
         FROM product_variants WHERE product_id = ANY($1) ORDER BY is_default DESC, name ASC",
    )
    .bind(product_ids)
    .fetch_all(db)
    .await?;

    let mut by_product: HashMap<i64, Vec<Variant>> = HashMap::new();
    for row in rows {
        by_product.entry(row.product_id).or_default().push(row);
    }
    Ok(by_product)
}

/// A scanned code within a company: barcode matches first, then active variants.
pub async fn find_by_code(
    db: impl PgExecutor<'_>,
    company_id: i64,
    code: &str,
) -> sqlx::Result<Option<CodeMatch>> {
    sqlx::query_as(
        "SELECT pv.id, pv.product_id, pv.barcode FROM product_variants pv \
         JOIN products p ON p.id = pv.product_id \
         WHERE p.company_id = $1 AND (pv.barcode = $2 OR pv.sku = $2) \
         ORDER BY (pv.barcode = $2) DESC, pv.is_active DESC LIMIT 1",
    )
    .bind(company_id)
    .bind(code)
    .fetch_optional(db)
    .await
}

// ============== Writes ==============

pub struct ProductInput<'a> {
    pub category_id: Option<i64>,
    pub name: &'a str,
    pub description: Option<&'a str>,
    pub image_data: Option<&'a str>,
    pub image_url: Option<&'a str>,
    pub is_active: bool,
}

pub struct VariantInput<'a> {
    pub sku: &'a str,
    pub barcode: Option<&'a str>,
    pub name: &'a str,
    pub attributes: Option<Value>,
    pub price: Decimal,
    pub standard_cost: Decimal,
    pub last_purchase_cost: Decimal,
    pub is_default: bool,
    pub is_active: bool,
    pub sale_price: Option<Decimal>,
    pub sale_start: Option<DateTime<Utc>>,
    pub sale_end: Option<DateTime<Utc>>,
}

/// SKUs are unique across all companies.
pub async fn sku_exists(db: impl PgExecutor<'_>, sku: &str) -> sqlx::Result<bool> {
    sqlx::query_scalar("SELECT EXISTS(SELECT 1 FROM product_variants WHERE sku = $1)")
        .bind(sku)
        .fetch_one(db)
        .await
}

pub async fn sku_in_other_product(
    db: impl PgExecutor<'_>,
    sku: &str,
    product_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM product_variants WHERE sku = $1 AND product_id != $2)",
    )
    .bind(sku)
    .bind(product_id)
    .fetch_one(db)
    .await
}

/// Barcodes are unique per company. Pass product id 0 on create to check every product.
pub async fn barcode_in_other_product(
    db: impl PgExecutor<'_>,
    company_id: i64,
    barcode: &str,
    product_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM product_variants pv JOIN products p ON p.id = pv.product_id \
         WHERE p.company_id = $1 AND pv.barcode = $2 AND pv.product_id != $3)",
    )
    .bind(company_id)
    .bind(barcode)
    .bind(product_id)
    .fetch_one(db)
    .await
}

pub async fn create_product(
    db: impl PgExecutor<'_>,
    company_id: i64,
    p: &ProductInput<'_>,
) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "INSERT INTO products (company_id, product_category_id, name, description, image_data, image_url, is_active) \
         VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
    )
    .bind(company_id)
    .bind(p.category_id)
    .bind(p.name)
    .bind(p.description)
    .bind(p.image_data)
    .bind(p.image_url)
    .bind(p.is_active)
    .fetch_one(db)
    .await
}

pub async fn update_product(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    p: &ProductInput<'_>,
) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE products SET product_category_id = $1, name = $2, description = $3, image_data = $4, \
             image_url = $5, is_active = $6, updated_at = NOW() WHERE id = $7 AND company_id = $8",
    )
    .bind(p.category_id)
    .bind(p.name)
    .bind(p.description)
    .bind(p.image_data)
    .bind(p.image_url)
    .bind(p.is_active)
    .bind(id)
    .bind(company_id)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn delete_product(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<()> {
    sqlx::query("DELETE FROM products WHERE id = $1 AND company_id = $2")
        .bind(id)
        .bind(company_id)
        .execute(db)
        .await?;
    Ok(())
}

pub async fn create_variant(
    db: impl PgExecutor<'_>,
    product_id: i64,
    v: &VariantInput<'_>,
) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO product_variants (product_id, sku, name, attributes, price, standard_cost, last_purchase_cost, \
             is_default, is_active, sale_price, sale_start, sale_end, barcode) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
    )
    .bind(product_id)
    .bind(v.sku)
    .bind(v.name)
    .bind(v.attributes.as_ref().map(Json))
    .bind(v.price)
    .bind(v.standard_cost)
    .bind(v.last_purchase_cost)
    .bind(v.is_default)
    .bind(v.is_active)
    .bind(v.sale_price)
    .bind(v.sale_start)
    .bind(v.sale_end)
    .bind(v.barcode)
    .execute(db)
    .await?;
    Ok(())
}

/// The id must come from this product's own variants.
pub async fn update_variant(
    db: impl PgExecutor<'_>,
    id: i64,
    v: &VariantInput<'_>,
) -> sqlx::Result<()> {
    sqlx::query(
        "UPDATE product_variants SET sku = $1, name = $2, attributes = $3, price = $4, standard_cost = $5, \
             last_purchase_cost = $6, is_default = $7, is_active = $8, sale_price = $9, sale_start = $10, \
             sale_end = $11, barcode = $12, updated_at = NOW() WHERE id = $13",
    )
    .bind(v.sku)
    .bind(v.name)
    .bind(v.attributes.as_ref().map(Json))
    .bind(v.price)
    .bind(v.standard_cost)
    .bind(v.last_purchase_cost)
    .bind(v.is_default)
    .bind(v.is_active)
    .bind(v.sale_price)
    .bind(v.sale_start)
    .bind(v.sale_end)
    .bind(v.barcode)
    .bind(id)
    .execute(db)
    .await?;
    Ok(())
}

pub async fn delete_variant(db: impl PgExecutor<'_>, id: i64) -> sqlx::Result<()> {
    sqlx::query("DELETE FROM product_variants WHERE id = $1")
        .bind(id)
        .execute(db)
        .await?;
    Ok(())
}

pub async fn delete_variants_of(db: impl PgExecutor<'_>, product_id: i64) -> sqlx::Result<()> {
    sqlx::query("DELETE FROM product_variants WHERE product_id = $1")
        .bind(product_id)
        .execute(db)
        .await?;
    Ok(())
}
