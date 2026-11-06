//! Reading `/api/v1/products`: list, get, and scanned-code lookup, like the Go
//! `ProductHandler`/`ProductService`. Writes on these paths aren't ported yet
//! and go to the Go API.

use std::collections::HashMap;

use axum::{Router, extract::State, http::StatusCode, routing::get};
use chrono::{DateTime, Utc};
use nexpos_db::products::{self as db, ListFilter, Product, Variant};
use rust_decimal::Decimal;
use serde::Serialize;
use serde_json::{Number, Value};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_with_pagination},
    extract::AuthUser,
    params::{QueryParams, RawId, parse_id},
    proxy,
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/products", get(list).fallback(proxy::fallback))
        .route(
            "/api/v1/products/lookup",
            get(lookup).fallback(proxy::fallback),
        )
        .route("/api/v1/products/{id}", get(show).fallback(proxy::fallback))
}

#[derive(Serialize)]
struct VariantResponse {
    id: i64,
    sku: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    barcode: Option<String>,
    name: String,
    attributes: Option<Value>,
    #[serde(serialize_with = "crate::money::serialize")]
    price: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    standard_cost: Decimal,
    #[serde(serialize_with = "crate::money::serialize")]
    last_purchase_cost: Decimal,
    is_default: bool,
    is_active: bool,
    #[serde(
        skip_serializing_if = "Option::is_none",
        serialize_with = "crate::money::serialize_some"
    )]
    sale_price: Option<Decimal>,
    #[serde(
        skip_serializing_if = "Option::is_none",
        serialize_with = "crate::time::serialize_some"
    )]
    sale_start: Option<DateTime<Utc>>,
    #[serde(
        skip_serializing_if = "Option::is_none",
        serialize_with = "crate::time::serialize_some"
    )]
    sale_end: Option<DateTime<Utc>>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<Variant> for VariantResponse {
    fn from(v: Variant) -> Self {
        Self {
            id: v.id,
            sku: v.sku,
            barcode: v.barcode,
            name: v.name,
            attributes: v.attributes.map(|a| go_numbers(a.0)),
            price: v.price,
            standard_cost: v.standard_cost,
            last_purchase_cost: v.last_purchase_cost,
            is_default: v.is_default,
            is_active: v.is_active,
            sale_price: v.sale_price,
            sale_start: v.sale_start,
            sale_end: v.sale_end,
            created_at: v.created_at,
            updated_at: v.updated_at,
        }
    }
}

/// Go decodes attributes into float64 and prints `2.0` as `2`.
fn go_numbers(value: Value) -> Value {
    match value {
        Value::Number(n) => match n.as_f64() {
            Some(f) if n.is_f64() && f.fract() == 0.0 && f.abs() < 1e15 => {
                Value::Number(Number::from(f as i64))
            }
            _ => Value::Number(n),
        },
        Value::Array(items) => Value::Array(items.into_iter().map(go_numbers).collect()),
        Value::Object(map) => {
            Value::Object(map.into_iter().map(|(k, v)| (k, go_numbers(v))).collect())
        }
        other => other,
    }
}

#[derive(Serialize)]
struct ProductResponse {
    id: i64,
    name: String,
    category_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    category_name: Option<String>,
    description: Option<String>,
    image_data: Option<String>,
    image_url: Option<String>,
    is_active: bool,
    variants: Vec<VariantResponse>,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

fn to_response(p: Product, variants: &mut HashMap<i64, Vec<Variant>>) -> ProductResponse {
    let own = variants.remove(&p.id).unwrap_or_default();
    ProductResponse {
        id: p.id,
        name: p.name,
        category_id: p.category_id,
        category_name: p.category_name,
        description: p.description,
        image_data: p.image_data,
        image_url: p.image_url,
        is_active: p.is_active,
        variants: own.into_iter().map(VariantResponse::from).collect(),
        created_at: p.created_at,
        updated_at: p.updated_at,
    }
}

async fn load(state: &AppState, company_id: i64, id: i64) -> AppResult<ProductResponse> {
    let product = db::get(&state.db, company_id, id)
        .await?
        // Go's message is "Product not found not found" (see `branches::not_found`).
        .ok_or_else(|| AppError::not_found("Product not found"))?;
    let mut variants = db::variants_of(&state.db, &[product.id]).await?;
    Ok(to_response(product, &mut variants))
}

async fn list(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<Vec<ProductResponse>>> {
    let (page, per_page) = query.page();
    let filter = ListFilter {
        search: query.get("search").trim(),
        category_id: query.int("category_id"),
        is_active: query.bool("is_active"),
        limit: per_page,
        offset: (page - 1) * per_page,
    };
    let (products, total) = db::list(&state.db, user.company_id, &filter).await?;
    let ids: Vec<i64> = products.iter().map(|p| p.id).collect();
    let mut variants = db::variants_of(&state.db, &ids).await?;
    // The Go handler never sets `count`, so it is always 0.
    let pagination = Pagination::new(total, page, per_page, 0);
    let data = products
        .into_iter()
        .map(|p| to_response(p, &mut variants))
        .collect();
    Ok(success_with_pagination(
        StatusCode::OK,
        "Product list retrieved",
        data,
        pagination,
        None,
    ))
}

async fn show(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<ProductResponse>> {
    let id = parse_id(&id, "product")?;
    let product = load(&state, user.company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Product retrieved successfully",
        product,
    ))
}

#[derive(Serialize)]
struct LookupResponse {
    product: ProductResponse,
    matched_variant_id: i64,
    matched_by: &'static str,
}

async fn lookup(
    State(state): State<AppState>,
    user: AuthUser,
    query: QueryParams,
) -> AppResult<Reply<LookupResponse>> {
    let code = query.get("code");
    if code.is_empty() {
        return Err(AppError::bad_request("code query parameter is required"));
    }
    let code = code.trim();
    if code.is_empty() {
        return Err(AppError::bad_request("code is required"));
    }

    let matched = db::find_by_code(&state.db, user.company_id, code)
        .await?
        .ok_or_else(|| AppError::not_found(format!("No product found for code: {code}")))?;
    let product = load(&state, user.company_id, matched.product_id).await?;
    let matched_by = if matched.barcode.as_deref() == Some(code) {
        "barcode"
    } else {
        "sku"
    };
    Ok(success(
        StatusCode::OK,
        "Product found",
        LookupResponse {
            product,
            matched_variant_id: matched.id,
            matched_by,
        },
    ))
}

#[cfg(test)]
mod tests {
    use serde_json::json;

    use super::*;

    #[test]
    fn whole_floats_print_like_go() {
        let parsed: Value =
            serde_json::from_str(r#"{"a":2.0,"b":[1.5,3.0],"c":"2.0","d":7}"#).unwrap();
        assert_eq!(
            serde_json::to_string(&go_numbers(parsed)).unwrap(),
            json!({"a":2,"b":[1.5,3],"c":"2.0","d":7}).to_string()
        );
    }
}
