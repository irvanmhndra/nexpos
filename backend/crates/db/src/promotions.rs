//! `promotions`, like the Go `promotionRepository`.

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Promotion {
    pub id: i64,
    pub code: String,
    pub name: String,
    pub description: Option<String>,
    #[sqlx(rename = "type")]
    pub kind: String,
    pub discount_type: Option<String>,
    pub discount_value: Option<Decimal>,
    pub min_purchase: Option<Decimal>,
    pub max_discount: Option<Decimal>,
    pub start_at: DateTime<Utc>,
    pub end_at: Option<DateTime<Utc>>,
    pub priority: i32,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub struct PromotionInput<'a> {
    pub code: &'a str,
    pub name: &'a str,
    pub description: Option<&'a str>,
    pub kind: &'a str,
    pub discount_type: Option<&'a str>,
    pub discount_value: Option<Decimal>,
    pub min_purchase: Option<Decimal>,
    pub max_discount: Option<Decimal>,
    pub start_at: DateTime<Utc>,
    pub end_at: Option<DateTime<Utc>>,
    pub priority: i32,
    pub is_active: bool,
}

pub struct ListFilter<'a> {
    pub search: &'a str,
    pub is_active: Option<bool>,
    pub kind: &'a str,
    pub status: &'a str,
    pub limit: i64,
    pub offset: i64,
}

const COLUMNS: &str = "id, code, name, description, type, discount_type, discount_value, min_purchase, \
     max_discount, start_at, end_at, priority, is_active, created_at, updated_at";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<Promotion>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM promotions WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn list(
    db: &PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<Promotion>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (name ILIKE ")
                .push_bind(like.clone())
                .push(" OR code ILIKE ")
                .push_bind(like.clone())
                .push(" OR description ILIKE ")
                .push_bind(like)
                .push(")");
        }
        if let Some(active) = f.is_active {
            q.push(" AND is_active = ").push_bind(active);
        }
        if !f.kind.is_empty() {
            q.push(" AND type = ").push_bind(f.kind.to_owned());
        }
        // Derived status, evaluated against the database clock; any other value is no filter.
        match f.status {
            "inactive" => {
                q.push(" AND is_active = false");
            }
            "scheduled" => {
                q.push(" AND is_active = true AND start_at > NOW()");
            }
            "active" => {
                q.push(" AND is_active = true AND start_at <= NOW() AND (end_at IS NULL OR end_at >= NOW())");
            }
            "expired" => {
                q.push(" AND is_active = true AND end_at IS NOT NULL AND end_at < NOW()");
            }
            _ => {}
        }
    };

    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM promotions");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {COLUMNS} FROM promotions"));
    filter(&mut rows);
    rows.push(" ORDER BY priority DESC, created_at DESC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

pub async fn code_exists(
    db: impl PgExecutor<'_>,
    company_id: i64,
    code: &str,
    exclude_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM promotions WHERE company_id = $1 AND code = $2 AND id != $3)",
    )
    .bind(company_id)
    .bind(code)
    .bind(exclude_id)
    .fetch_one(db)
    .await
}

/// Returns the new row's `(id, created_at, updated_at)`.
pub async fn create(
    db: impl PgExecutor<'_>,
    company_id: i64,
    p: &PromotionInput<'_>,
) -> sqlx::Result<(i64, DateTime<Utc>, DateTime<Utc>)> {
    sqlx::query_as(
        "INSERT INTO promotions (company_id, code, name, description, type, discount_type, discount_value, \
             min_purchase, max_discount, start_at, end_at, priority, is_active, created_at, updated_at) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW()) \
         RETURNING id, created_at, updated_at",
    )
    .bind(company_id)
    .bind(p.code)
    .bind(p.name)
    .bind(p.description)
    .bind(p.kind)
    .bind(p.discount_type)
    .bind(p.discount_value)
    .bind(p.min_purchase)
    .bind(p.max_discount)
    .bind(p.start_at)
    .bind(p.end_at)
    .bind(p.priority)
    .bind(p.is_active)
    .fetch_one(db)
    .await
}

/// Returns the new `updated_at`, or `None` if there is no such promotion.
pub async fn update(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    p: &PromotionInput<'_>,
) -> sqlx::Result<Option<DateTime<Utc>>> {
    sqlx::query_scalar(
        "UPDATE promotions SET code = $1, name = $2, description = $3, type = $4, discount_type = $5, \
             discount_value = $6, min_purchase = $7, max_discount = $8, start_at = $9, end_at = $10, \
             priority = $11, is_active = $12, updated_at = NOW() \
         WHERE id = $13 AND company_id = $14 RETURNING updated_at",
    )
    .bind(p.code)
    .bind(p.name)
    .bind(p.description)
    .bind(p.kind)
    .bind(p.discount_type)
    .bind(p.discount_value)
    .bind(p.min_purchase)
    .bind(p.max_discount)
    .bind(p.start_at)
    .bind(p.end_at)
    .bind(p.priority)
    .bind(p.is_active)
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn delete(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    let res = sqlx::query("DELETE FROM promotions WHERE id = $1 AND company_id = $2")
        .bind(id)
        .bind(company_id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}

pub async fn by_code(
    db: impl PgExecutor<'_>,
    company_id: i64,
    code: &str,
) -> sqlx::Result<Option<Promotion>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM promotions WHERE company_id = $1 AND code = $2"
    ))
    .bind(company_id)
    .bind(code)
    .fetch_optional(db)
    .await
}

/// Running now, highest priority first.
pub async fn active(
    db: impl PgExecutor<'_>,
    company_id: i64,
    now: DateTime<Utc>,
) -> sqlx::Result<Vec<Promotion>> {
    sqlx::query_as(&format!(
        "SELECT {COLUMNS} FROM promotions \
         WHERE company_id = $1 AND is_active = true AND start_at <= $2 AND (end_at IS NULL OR end_at >= $2) \
         ORDER BY priority DESC, created_at DESC"
    ))
    .bind(company_id)
    .bind(now)
    .fetch_all(db)
    .await
}
