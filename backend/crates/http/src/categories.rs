//! `/api/v1/product-categories`: same requests, responses, and errors as the Go
//! `ProductCategoryHandler`/`ProductCategoryService`.

use axum::{Router, extract::State, http::StatusCode, routing::get};
use chrono::{DateTime, Utc};
use nexpos_db::categories::{self as db, Category, CategoryInput, ListFilter};
use serde::{Deserialize, Serialize};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/product-categories", get(list).post(create))
        .route("/api/v1/product-categories/all", get(all))
        .route(
            "/api/v1/product-categories/{id}",
            get(show).put(update).delete(destroy),
        )
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct CategoryRequest {
    code: Option<String>,
    name: Option<String>,
    parent_id: Option<i64>,
    sort_order: Option<i32>,
    is_active: Option<bool>,
}

impl CategoryRequest {
    fn code(&self) -> &str {
        self.code.as_deref().unwrap_or_default()
    }

    fn name(&self) -> &str {
        self.name.as_deref().unwrap_or_default()
    }

    fn sort_order(&self) -> i32 {
        self.sort_order.unwrap_or(0)
    }

    /// `CreateProductCategoryRequest` / `UpdateProductCategoryRequest` tags.
    fn validate(&self) -> AppResult<()> {
        Validator::new()
            .field(
                "code",
                self.code(),
                &[Rule::Required, Rule::Min(1), Rule::Max(50)],
            )
            .field(
                "name",
                self.name(),
                &[Rule::Required, Rule::Min(1), Rule::Max(255)],
            )
            .invalid_unless("sort_order", self.sort_order() >= 0)
            .finish()
    }
}

#[derive(Serialize)]
struct CategoryResponse {
    id: i64,
    code: String,
    name: String,
    parent_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    parent_name: Option<String>,
    sort_order: i32,
    is_active: bool,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<Category> for CategoryResponse {
    fn from(c: Category) -> Self {
        Self {
            id: c.id,
            code: c.code,
            name: c.name,
            parent_id: c.parent_id,
            parent_name: c.parent_name,
            sort_order: c.sort_order,
            is_active: c.is_active,
            created_at: c.created_at,
            updated_at: c.updated_at,
        }
    }
}

/// Go's message is "Category not found not found" (see `branches::not_found`).
fn not_found() -> AppError {
    AppError::not_found("Category not found")
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<CategoryResponse>>> {
    let (page, per_page) = query.page();
    let filter = ListFilter {
        search: query.get("search"),
        is_active: query.bool("is_active"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (rows, total) = db::list(&state.db, user.company_id, &filter).await?;
    // The Go handler never sets `count`, so it is always 0.
    let pagination = Pagination::new(total, page, per_page, 0);
    let data = rows.into_iter().map(CategoryResponse::from).collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Category list retrieved",
        data,
        pagination,
        None,
    ))
}

async fn all(
    State(state): State<AppState>,
    user: AuthUser,
) -> AppResult<Reply<Vec<CategoryResponse>>> {
    let rows = db::all_active(&state.db, user.company_id).await?;
    let data = rows.into_iter().map(CategoryResponse::from).collect();
    Ok(success(StatusCode::OK, "All categories retrieved", data))
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<CategoryRequest>,
) -> AppResult<Reply<CategoryResponse>> {
    req.validate()?;
    let company_id = user.company_id;
    if db::code_exists(&state.db, company_id, req.code(), 0).await? {
        return Err(AppError::bad_request("Category code already exists"));
    }
    if let Some(parent_id) = req.parent_id
        && db::get(&state.db, company_id, parent_id).await?.is_none()
    {
        return Err(AppError::bad_request("Parent category not found"));
    }

    let input = CategoryInput {
        code: req.code(),
        name: req.name(),
        parent_id: req.parent_id,
        sort_order: req.sort_order(),
        is_active: req.is_active.unwrap_or(true),
    };
    let id = db::create(&state.db, company_id, &input).await?;
    let category = db::get(&state.db, company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::CREATED,
        "Category created successfully",
        category.into(),
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<CategoryResponse>> {
    let id = parse_id(&id, "category")?;
    let category = db::get(&state.db, user.company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Category retrieved successfully",
        category.into(),
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<CategoryRequest>,
) -> AppResult<Reply<CategoryResponse>> {
    let id = parse_id(&id, "category")?;
    req.validate()?;
    let company_id = user.company_id;
    let current = db::get(&state.db, company_id, id)
        .await?
        .ok_or_else(not_found)?;
    if db::code_exists(&state.db, company_id, req.code(), id).await? {
        return Err(AppError::bad_request("Category code already exists"));
    }
    if let Some(parent_id) = req.parent_id {
        if parent_id == id {
            return Err(AppError::bad_request("Category cannot be its own parent"));
        }
        let parent = db::get(&state.db, company_id, parent_id)
            .await?
            .ok_or_else(|| AppError::bad_request("Parent category not found"))?;
        if parent.parent_id == Some(id) {
            return Err(AppError::bad_request("Circular reference detected"));
        }
    }

    let input = CategoryInput {
        code: req.code(),
        name: req.name(),
        parent_id: req.parent_id,
        sort_order: req.sort_order(),
        is_active: req.is_active.unwrap_or(current.is_active),
    };
    if !db::update(&state.db, company_id, id, &input).await? {
        return Err(not_found());
    }
    let category = db::get(&state.db, company_id, id)
        .await?
        .ok_or_else(not_found)?;
    Ok(success(
        StatusCode::OK,
        "Category updated successfully",
        category.into(),
    ))
}

async fn destroy(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_id(&id, "category")?;
    let company_id = user.company_id;
    if db::get(&state.db, company_id, id).await?.is_none() {
        return Err(not_found());
    }
    if db::has_children(&state.db, company_id, id).await? {
        return Err(AppError::bad_request(
            "Cannot delete category with subcategories",
        ));
    }
    if db::has_products(&state.db, company_id, id).await? {
        return Err(AppError::bad_request(
            "Cannot delete category with products",
        ));
    }
    if !db::delete(&state.db, company_id, id).await? {
        return Err(not_found());
    }
    Ok(success_message(
        StatusCode::OK,
        "Category deleted successfully",
    ))
}
