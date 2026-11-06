//! `expense_categories` and `expenses`, like the Go repositories. Dates are bound
//! as text and cast, so an unparseable one is an error from Postgres, as in Go.

use chrono::{DateTime, NaiveDate, Utc};
use rust_decimal::Decimal;
use sqlx::{PgExecutor, PgPool, Postgres, QueryBuilder};

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct ExpenseCategory {
    pub id: i64,
    pub name: String,
    pub description: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct Expense {
    pub id: i64,
    pub branch_id: Option<i64>,
    pub category_id: Option<i64>,
    pub amount: Decimal,
    pub description: String,
    pub reference_no: Option<String>,
    pub expense_date: NaiveDate,
    pub recorded_by: Option<i64>,
    pub notes: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub struct ExpenseInput<'a> {
    pub branch_id: Option<i64>,
    pub category_id: Option<i64>,
    pub amount: Decimal,
    pub description: &'a str,
    pub reference_no: Option<&'a str>,
    pub expense_date: &'a str,
    pub notes: Option<&'a str>,
}

#[derive(Clone, Copy)]
pub struct ExpenseFilter<'a> {
    pub branch_id: Option<i64>,
    pub category_id: Option<i64>,
    pub date_from: &'a str,
    pub date_to: &'a str,
}

#[derive(Debug, sqlx::FromRow)]
pub struct CategoryTotal {
    pub id: i64,
    pub name: String,
    pub total_amount: Decimal,
    pub count: i64,
}

// ============== Categories ==============

const CATEGORY_COLUMNS: &str = "id, name, description, created_at, updated_at";

pub async fn get_category(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<ExpenseCategory>> {
    sqlx::query_as(&format!(
        "SELECT {CATEGORY_COLUMNS} FROM expense_categories WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn list_categories(
    db: &PgPool,
    company_id: i64,
    search: &str,
    limit: i64,
    offset: i64,
) -> sqlx::Result<(Vec<ExpenseCategory>, i64)> {
    let filter = |q: &mut QueryBuilder<'_, Postgres>| {
        q.push(" WHERE company_id = ").push_bind(company_id);
        if !search.is_empty() {
            q.push(" AND name ILIKE ").push_bind(format!("%{search}%"));
        }
    };
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM expense_categories");
    filter(&mut count);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {CATEGORY_COLUMNS} FROM expense_categories"));
    filter(&mut rows);
    rows.push(" ORDER BY name ASC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

pub async fn create_category(
    db: impl PgExecutor<'_>,
    company_id: i64,
    name: &str,
    description: Option<&str>,
) -> sqlx::Result<ExpenseCategory> {
    sqlx::query_as(&format!(
        "INSERT INTO expense_categories (company_id, name, description) VALUES ($1, $2, $3) \
         RETURNING {CATEGORY_COLUMNS}"
    ))
    .bind(company_id)
    .bind(name)
    .bind(description)
    .fetch_one(db)
    .await
}

pub async fn update_category(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    name: &str,
    description: Option<&str>,
) -> sqlx::Result<Option<ExpenseCategory>> {
    sqlx::query_as(&format!(
        "UPDATE expense_categories SET name = $1, description = $2, updated_at = NOW() \
         WHERE id = $3 AND company_id = $4 RETURNING {CATEGORY_COLUMNS}"
    ))
    .bind(name)
    .bind(description)
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn delete_category(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<bool> {
    let res = sqlx::query("DELETE FROM expense_categories WHERE id = $1 AND company_id = $2")
        .bind(id)
        .bind(company_id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}

// ============== Expenses ==============

const EXPENSE_COLUMNS: &str = "id, branch_id, category_id, amount, description, reference_no, expense_date, \
     recorded_by, notes, created_at, updated_at";

fn push_filter(q: &mut QueryBuilder<'_, Postgres>, company_id: i64, f: ExpenseFilter<'_>) {
    q.push(" WHERE company_id = ").push_bind(company_id);
    if let Some(branch_id) = f.branch_id {
        q.push(" AND branch_id = ").push_bind(branch_id);
    }
    if let Some(category_id) = f.category_id {
        q.push(" AND category_id = ").push_bind(category_id);
    }
    if !f.date_from.is_empty() {
        q.push(" AND expense_date >= ")
            .push_bind(f.date_from.to_owned())
            .push("::date");
    }
    if !f.date_to.is_empty() {
        q.push(" AND expense_date <= ")
            .push_bind(f.date_to.to_owned())
            .push("::date");
    }
}

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
) -> sqlx::Result<Option<Expense>> {
    sqlx::query_as(&format!(
        "SELECT {EXPENSE_COLUMNS} FROM expenses WHERE id = $1 AND company_id = $2"
    ))
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn list(
    db: &PgPool,
    company_id: i64,
    f: ExpenseFilter<'_>,
    limit: i64,
    offset: i64,
) -> sqlx::Result<(Vec<Expense>, i64)> {
    let mut count = QueryBuilder::new("SELECT COUNT(*) FROM expenses");
    push_filter(&mut count, company_id, f);
    let total: i64 = count.build_query_scalar().fetch_one(db).await?;

    let mut rows = QueryBuilder::new(format!("SELECT {EXPENSE_COLUMNS} FROM expenses"));
    push_filter(&mut rows, company_id, f);
    rows.push(" ORDER BY expense_date DESC, created_at DESC LIMIT ")
        .push_bind(limit)
        .push(" OFFSET ")
        .push_bind(offset);
    Ok((rows.build_query_as().fetch_all(db).await?, total))
}

/// Returns the new row's `(id, created_at, updated_at)`.
pub async fn create(
    db: impl PgExecutor<'_>,
    company_id: i64,
    recorded_by: i64,
    e: &ExpenseInput<'_>,
) -> sqlx::Result<(i64, DateTime<Utc>, DateTime<Utc>)> {
    sqlx::query_as(
        "INSERT INTO expenses (company_id, branch_id, category_id, amount, description, reference_no, \
             expense_date, recorded_by, notes) \
         VALUES ($1, $2, $3, $4, $5, $6, $7::date, $8, $9) RETURNING id, created_at, updated_at",
    )
    .bind(company_id)
    .bind(e.branch_id)
    .bind(e.category_id)
    .bind(e.amount)
    .bind(e.description)
    .bind(e.reference_no)
    .bind(e.expense_date)
    .bind(recorded_by)
    .bind(e.notes)
    .fetch_one(db)
    .await
}

pub async fn update(
    db: impl PgExecutor<'_>,
    company_id: i64,
    id: i64,
    e: &ExpenseInput<'_>,
) -> sqlx::Result<Option<Expense>> {
    sqlx::query_as(&format!(
        "UPDATE expenses SET branch_id = $1, category_id = $2, amount = $3, description = $4, reference_no = $5, \
             expense_date = $6::date, notes = $7, updated_at = NOW() \
         WHERE id = $8 AND company_id = $9 RETURNING {EXPENSE_COLUMNS}"
    ))
    .bind(e.branch_id)
    .bind(e.category_id)
    .bind(e.amount)
    .bind(e.description)
    .bind(e.reference_no)
    .bind(e.expense_date)
    .bind(e.notes)
    .bind(id)
    .bind(company_id)
    .fetch_optional(db)
    .await
}

pub async fn delete(db: impl PgExecutor<'_>, company_id: i64, id: i64) -> sqlx::Result<bool> {
    let res = sqlx::query("DELETE FROM expenses WHERE id = $1 AND company_id = $2")
        .bind(id)
        .bind(company_id)
        .execute(db)
        .await?;
    Ok(res.rows_affected() > 0)
}

pub async fn total(db: &PgPool, company_id: i64, f: ExpenseFilter<'_>) -> sqlx::Result<Decimal> {
    let mut q = QueryBuilder::new("SELECT COALESCE(SUM(amount), 0) FROM expenses");
    push_filter(&mut q, company_id, f);
    q.build_query_scalar().fetch_one(db).await
}

/// Totals per category that has expenses in the filter, by category name.
pub async fn totals_by_category(
    db: &PgPool,
    company_id: i64,
    f: ExpenseFilter<'_>,
) -> sqlx::Result<Vec<CategoryTotal>> {
    let mut q = QueryBuilder::new(
        "SELECT c.id, c.name, t.total_amount, t.count FROM expense_categories c JOIN \
         (SELECT category_id, COALESCE(SUM(amount), 0) AS total_amount, COUNT(*) AS count FROM expenses",
    );
    push_filter(&mut q, company_id, f);
    q.push(" AND category_id IS NOT NULL GROUP BY category_id) t ON t.category_id = c.id WHERE c.company_id = ")
        .push_bind(company_id)
        .push(" ORDER BY c.name ASC");
    q.build_query_as().fetch_all(db).await
}

/// The total and count of expenses with no category.
pub async fn uncategorized(
    db: &PgPool,
    company_id: i64,
    f: ExpenseFilter<'_>,
) -> sqlx::Result<(Decimal, i64)> {
    let mut q = QueryBuilder::new("SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM expenses");
    push_filter(&mut q, company_id, f);
    q.push(" AND category_id IS NULL");
    q.build_query_as().fetch_one(db).await
}

/// Whether a branch belongs to the company.
pub async fn own_branch(
    db: impl PgExecutor<'_>,
    company_id: i64,
    branch_id: i64,
) -> sqlx::Result<bool> {
    sqlx::query_scalar("SELECT EXISTS (SELECT 1 FROM branches WHERE company_id = $1 AND id = $2)")
        .bind(company_id)
        .bind(branch_id)
        .fetch_one(db)
        .await
}
