//! `product_categories`, like the Go `ProductCategoryRepository`. The parent's
//! name comes from a join rather than one lookup per row.

use chrono::{DateTime, Utc};
use sqlx::{PgExecutor, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Category {
    pub id: i64,
    pub code: String,
    pub name: String,
    pub parent_id: Option<i64>,
    pub parent_name: Option<String>,
    pub sort_order: i32,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub struct CategoryInput<'a> {
    pub code: &'a str,
    pub name: &'a str,
    pub parent_id: Option<i64>,
    pub sort_order: i32,
    pub is_active: bool,
}

pub struct ListFilter<'a> {
    pub search: &'a str,
    pub is_active: Option<bool>,
    pub limit: i64,
    pub offset: i64,
}

const SELECT: &str = "SELECT c.id, c.code, c.name, c.parent_id, p.name AS parent_name, \
     c.sort_order, c.is_active, c.created_at, c.updated_at \
     FROM product_categories c \
     LEFT JOIN product_categories p ON p.id = c.parent_id AND p.company_id = c.company_id";

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<Category>> {
    sqlx::query_as(&format!("{SELECT} WHERE c.id = $1 AND c.company_id = $2"))
        .bind(id)
        .bind(company_id)
        .fetch_optional(db)
        .await
}

pub async fn list(
    db: &sqlx::PgPool,
    company_id: i64,
    f: &ListFilter<'_>,
) -> sqlx::Result<(Vec<Category>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE c.company_id = ").push_bind(company_id);
        if !f.search.is_empty() {
            let like = format!("%{}%", f.search);
            q.push(" AND (c.name ILIKE ")
                .push_bind(like.clone())
                .push(" OR c.code ILIKE ")
                .push_bind(like)
                .push(")");
        }
        if let Some(active) = f.is_active {
            q.push(" AND c.is_active = ").push_bind(active);
        }
    };

    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM product_categories c");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(SELECT);
    filter(&mut rows);
    rows.push(" ORDER BY c.sort_order ASC, c.name ASC LIMIT ")
        .push_bind(f.limit)
        .push(" OFFSET ")
        .push_bind(f.offset);
    let categories = rows.build_query_as().fetch_all(db).await?;
    Ok((categories, total))
}

/// Every active category, for dropdowns.
pub async fn all_active(db: impl PgExecutor<'_>, company_id: i64) -> sqlx::Result<Vec<Category>> {
    sqlx::query_as(&format!(
        "{SELECT} WHERE c.company_id = $1 AND c.is_active = true ORDER BY c.sort_order ASC, c.name ASC"
    ))
    .bind(company_id)
    .fetch_all(db)
    .await
}

pub async fn code_exists(
    db: impl PgExecutor<'_>,
    company_id: i64,
    code: &str,
    exclude_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM product_categories WHERE company_id = $1 AND code = $2 AND id != $3)",
    )
    .bind(company_id)
    .bind(code)
    .bind(exclude_id)
    .fetch_one(db)
    .await
}

pub async fn has_children(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM product_categories WHERE company_id = $1 AND parent_id = $2)",
    )
    .bind(company_id)
    .bind(id)
    .fetch_one(db)
    .await
}

pub async fn has_products(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    sqlx::query_scalar(
        "SELECT EXISTS(SELECT 1 FROM products WHERE company_id = $1 AND product_category_id = $2)",
    )
    .bind(company_id)
    .bind(id)
    .fetch_one(db)
    .await
}

pub async fn create(
    db: impl PgExecutor<'_>,
    company_id: i64,
    c: &CategoryInput<'_>,
) -> sqlx::Result<i64> {
    sqlx::query_scalar(
        "INSERT INTO product_categories (company_id, parent_id, code, name, sort_order, is_active) \
         VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
    )
    .bind(company_id)
    .bind(c.parent_id)
    .bind(c.code)
    .bind(c.name)
    .bind(c.sort_order)
    .bind(c.is_active)
    .fetch_one(db)
    .await
}

/// `true` if a row was updated.
pub async fn update(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    c: &CategoryInput<'_>,
) -> sqlx::Result<bool> {
    let res = sqlx::query(
        "UPDATE product_categories \
         SET code = $1, name = $2, parent_id = $3, sort_order = $4, is_active = $5, updated_at = NOW() \
         WHERE id = $6 AND company_id = $7",
    )
    .bind(c.code)
    .bind(c.name)
    .bind(c.parent_id)
    .bind(c.sort_order)
    .bind(c.is_active)
    .bind(id)
    .bind(company_id)
    .execute(db)
    .await?;
    Ok(res.rows_affected() > 0)
}

/// `true` if a row was deleted.
pub async fn delete(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    let res = sqlx::query("DELETE FROM product_categories WHERE id = $1 AND company_id = $2")
        .bind(id)
        .bind(company_id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}
