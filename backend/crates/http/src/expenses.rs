//! `/api/v1/expense-categories` and `/api/v1/expenses`: same requests, responses,
//! and errors as the Go handlers and `ExpenseService`.
//!
//! Go quirk kept for the web app: creating an expense echoes the date and amount
//! as sent, while every read shows the columns (a date as a midnight UTC time, an
//! amount rounded to 2 places).

use axum::{
    Router,
    extract::State,
    http::StatusCode,
    routing::{get, put},
};
use chrono::{DateTime, NaiveDate, Utc};
use nexpos_db::expenses::{self as db, Expense, ExpenseCategory, ExpenseFilter, ExpenseInput};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route(
            "/api/v1/expense-categories",
            get(list_categories).post(create_category),
        )
        .route(
            "/api/v1/expense-categories/{id}",
            put(update_category).delete(delete_category),
        )
        .route("/api/v1/expenses", get(list).post(create))
        .route("/api/v1/expenses/summary", get(summary))
        .route(
            "/api/v1/expenses/{id}",
            get(show).put(update).delete(destroy),
        )
}

// ============== Categories ==============

#[derive(Default, Deserialize)]
#[serde(default)]
struct CategoryRequest {
    name: Option<String>,
    description: Option<String>,
}

impl CategoryRequest {
    fn validated_name(&self) -> AppResult<&str> {
        let name = self.name.as_deref().unwrap_or_default();
        Validator::new()
            .field(
                "name",
                name,
                &[Rule::Required, Rule::Min(1), Rule::Max(100)],
            )
            .finish()?;
        Ok(name)
    }
}

#[derive(Serialize)]
struct CategoryResponse {
    id: i64,
    name: String,
    description: Option<String>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<ExpenseCategory> for CategoryResponse {
    fn from(c: ExpenseCategory) -> Self {
        Self {
            id: c.id,
            name: c.name,
            description: c.description,
            created_at: c.created_at,
            updated_at: c.updated_at,
        }
    }
}

/// Go's message is "Expense category not found not found" (see `branches::not_found`).
fn category_not_found() -> AppError {
    AppError::not_found("Expense category not found")
}

async fn list_categories(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<CategoryResponse>>> {
    let (page, per_page) = query.page();
    let (rows, total) = db::list_categories(
        &state.db,
        user.company_id,
        query.get("search"),
        per_page,
        (page - 1) * per_page,
    )
    .await?;
    let pagination = Pagination::new(total, page, per_page, 0);
    let data = rows.into_iter().map(CategoryResponse::from).collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Expense categories retrieved",
        data,
        pagination,
        None,
    ))
}

async fn create_category(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<CategoryRequest>,
) -> AppResult<Reply<CategoryResponse>> {
    let name = req.validated_name()?;
    let row =
        db::create_category(&state.db, user.company_id, name, req.description.as_deref()).await?;
    Ok(success(
        StatusCode::CREATED,
        "Expense category created successfully",
        row.into(),
    ))
}

fn parse_category_id(raw: &str) -> AppResult<i64> {
    parse_id(raw, "expense category")
}

async fn update_category(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<CategoryRequest>,
) -> AppResult<Reply<CategoryResponse>> {
    let id = parse_category_id(&id)?;
    let name = req.validated_name()?;
    let row = db::update_category(
        &state.db,
        user.company_id,
        id,
        name,
        req.description.as_deref(),
    )
    .await?
    .ok_or_else(category_not_found)?;
    Ok(success(
        StatusCode::OK,
        "Expense category updated successfully",
        row.into(),
    ))
}

async fn delete_category(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_category_id(&id)?;
    if !db::delete_category(&state.db, user.company_id, id).await? {
        return Err(category_not_found());
    }
    Ok(success_message(
        StatusCode::OK,
        "Expense category deleted successfully",
    ))
}

// ============== Expenses ==============

#[derive(Default, Deserialize)]
#[serde(default)]
struct ExpenseRequest {
    branch_id: Option<i64>,
    category_id: Option<i64>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    amount: Option<Decimal>,
    description: Option<String>,
    reference_no: Option<String>,
    expense_date: Option<String>,
    notes: Option<String>,
}

const MIN_AMOUNT: Decimal = Decimal::from_parts(1, 0, 0, false, 2);

impl ExpenseRequest {
    /// `CreateExpenseRequest` / `UpdateExpenseRequest` tags.
    fn validate(&self) -> AppResult<ExpenseInput<'_>> {
        let amount = self.amount.unwrap_or(Decimal::ZERO);
        let description = self.description.as_deref().unwrap_or_default();
        let expense_date = self.expense_date.as_deref().unwrap_or_default();
        let mut v = Validator::new();
        if amount.is_zero() {
            v.push("amount", "amount is required");
        } else if amount < MIN_AMOUNT {
            // go-playground's `min` message says "characters" even for numbers.
            v.push("amount", "amount must be at least 0.01 characters");
        }
        v.field(
            "description",
            description,
            &[Rule::Required, Rule::Min(1), Rule::Max(500)],
        )
        .optional(
            "reference_no",
            self.reference_no.as_deref(),
            &[Rule::Max(100)],
        )
        .field("expense_date", expense_date, &[Rule::Required])
        .finish()?;
        Ok(ExpenseInput {
            branch_id: self.branch_id,
            category_id: self.category_id,
            amount,
            description,
            reference_no: self.reference_no.as_deref(),
            expense_date,
            notes: self.notes.as_deref(),
        })
    }
}

#[derive(Serialize)]
struct ExpenseResponse {
    id: i64,
    branch_id: Option<i64>,
    category_id: Option<i64>,
    #[serde(serialize_with = "crate::money::serialize")]
    amount: Decimal,
    description: String,
    reference_no: Option<String>,
    expense_date: String,
    recorded_by: Option<i64>,
    notes: Option<String>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

/// Go scans the date into a string, which formats it as a midnight UTC time.
fn go_date(date: NaiveDate) -> String {
    date.format("%Y-%m-%dT00:00:00Z").to_string()
}

impl From<Expense> for ExpenseResponse {
    fn from(e: Expense) -> Self {
        Self {
            id: e.id,
            branch_id: e.branch_id,
            category_id: e.category_id,
            amount: e.amount,
            description: e.description,
            reference_no: e.reference_no,
            expense_date: go_date(e.expense_date),
            recorded_by: e.recorded_by,
            notes: e.notes,
            created_at: e.created_at,
            updated_at: e.updated_at,
        }
    }
}

/// Go's message is "Expense not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Expense not found")
}

/// The category must be the company's (404), then the branch (400), in Go's order.
async fn check_references(
    state: &AppState,
    company_id: i64,
    input: &ExpenseInput<'_>,
) -> AppResult<()> {
    if let Some(category_id) = input.category_id
        && db::get_category(&state.db, company_id, category_id)
            .await?
            .is_none()
    {
        return Err(category_not_found());
    }
    if let Some(branch_id) = input.branch_id
        && !db::own_branch(&state.db, company_id, branch_id).await?
    {
        return Err(AppError::bad_request("Branch not found"));
    }
    Ok(())
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<ExpenseResponse>>> {
    let (page, per_page) = query.page();
    let filter = ExpenseFilter {
        branch_id: query.int("branch_id"),
        category_id: query.int("category_id"),
        date_from: query.get("date_from"),
        date_to: query.get("date_to"),
    };
    let (rows, total) = db::list(
        &state.db,
        user.company_id,
        filter,
        per_page,
        (page - 1) * per_page,
    )
    .await?;
    let pagination = Pagination::new(total, page, per_page, 0);
    let data = rows.into_iter().map(ExpenseResponse::from).collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Expenses retrieved",
        data,
        pagination,
        None,
    ))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<ExpenseRequest>,
) -> AppResult<Reply<ExpenseResponse>> {
    let input = req.validate()?;
    check_references(&state, user.company_id, &input).await?;
    let (id, created_at, updated_at) =
        db::create(&state.db, user.company_id, user.user_id, &input).await?;
    let body = ExpenseResponse {
        id,
        branch_id: input.branch_id,
        category_id: input.category_id,
        amount: input.amount,
        description: input.description.to_owned(),
        reference_no: input.reference_no.map(str::to_owned),
        expense_date: input.expense_date.to_owned(),
        recorded_by: Some(user.user_id),
        notes: input.notes.map(str::to_owned),
        created_at,
        updated_at,
    };
    Ok(success(
        StatusCode::CREATED,
        "Expense created successfully",
        body,
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<ExpenseResponse>> {
    let id = parse_id(&id, "expense")?;
    let row = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Expense retrieved successfully",
        row.into(),
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<ExpenseRequest>,
) -> AppResult<Reply<ExpenseResponse>> {
    let id = parse_id(&id, "expense")?;
    let input = req.validate()?;
    if db::get(&state.db, user.company_id, id).await?.is_none() {
        return Err(not_found());
    }
    check_references(&state, user.company_id, &input).await?;
    let row = db::update(&state.db, user.company_id, id, &input)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Expense updated successfully",
        row.into(),
    ))
}

async fn destroy(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_id(&id, "expense")?;
    if !db::delete(&state.db, user.company_id, id).await? {
        return Err(not_found());
    }
    Ok(success_message(
        StatusCode::OK,
        "Expense deleted successfully",
    ))
}

#[derive(Serialize)]
struct CategoryTotalResponse {
    category_id: Option<i64>,
    category_name: String,
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
    count: i64,
}

#[derive(Serialize)]
struct SummaryResponse {
    #[serde(serialize_with = "crate::money::serialize")]
    total_amount: Decimal,
    date_from: String,
    date_to: String,
    by_category: Vec<CategoryTotalResponse>,
}

async fn summary(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<SummaryResponse>> {
    let filter = ExpenseFilter {
        branch_id: query.int("branch_id"),
        category_id: None,
        date_from: query.get("date_from"),
        date_to: query.get("date_to"),
    };
    let total_amount = db::total(&state.db, user.company_id, filter).await?;
    let mut by_category: Vec<CategoryTotalResponse> =
        db::totals_by_category(&state.db, user.company_id, filter)
            .await?
            .into_iter()
            .map(|c| CategoryTotalResponse {
                category_id: Some(c.id),
                category_name: c.name,
                total_amount: c.total_amount,
                count: c.count,
            })
            .collect();

    let (uncategorized, count) = db::uncategorized(&state.db, user.company_id, filter).await?;
    if count > 0 {
        by_category.push(CategoryTotalResponse {
            category_id: None,
            category_name: format!("Uncategorized ({count})"),
            total_amount: uncategorized,
            count,
        });
    }

    let body = SummaryResponse {
        total_amount,
        date_from: filter.date_from.to_owned(),
        date_to: filter.date_to.to_owned(),
        by_category,
    };
    Ok(success(StatusCode::OK, "Expense summary retrieved", body))
}
