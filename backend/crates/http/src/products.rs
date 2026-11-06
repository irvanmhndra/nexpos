//! `/api/v1/products`: list, get, scanned-code lookup, create, update, and delete,
//! like the Go `ProductHandler`/`ProductService`. Writes run in one transaction
//! (Go applies them step by step), and R2 images are removed only after commit.

use std::collections::{HashMap, HashSet};

use axum::{Router, extract::State, http::StatusCode, routing::get};
use chrono::{DateTime, Utc};
use nexpos_db::{
    categories,
    products::{self as db, ListFilter, Product, ProductInput, Variant, VariantInput},
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use serde_json::{Map, Number, Value};

use crate::{
    AppError, AppResult, AppState,
    envelope::{Pagination, Reply, success, success_message, success_with_pagination},
    extract::{AuthUser, JsonBody, Rule, Validator},
    params::{QueryParams, RawId, parse_id},
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/api/v1/products", get(list).post(create))
        .route("/api/v1/products/lookup", get(lookup))
        .route(
            "/api/v1/products/{id}",
            get(show).put(update).delete(destroy),
        )
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
pub(crate) fn go_numbers(value: Value) -> Value {
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

// ============== Writes ==============

#[derive(Default, Deserialize)]
#[serde(default)]
struct VariantRequest {
    id: Option<i64>,
    sku: Option<String>,
    barcode: Option<String>,
    name: Option<String>,
    attributes: Option<Map<String, Value>>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    price: Option<Decimal>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    standard_cost: Option<Decimal>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    last_purchase_cost: Option<Decimal>,
    is_default: Option<bool>,
    is_active: Option<bool>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    sale_price: Option<Decimal>,
    sale_start: Option<String>,
    sale_end: Option<String>,
}

impl VariantRequest {
    fn sku(&self) -> &str {
        self.sku.as_deref().unwrap_or_default()
    }

    /// A trimmed barcode; a blank one is no barcode.
    fn barcode(&self) -> Option<&str> {
        self.barcode
            .as_deref()
            .map(str::trim)
            .filter(|b| !b.is_empty())
    }

    /// `position` is the variant's place in the request: the first one is always a default.
    fn input(&self, position: usize) -> VariantInput<'_> {
        VariantInput {
            sku: self.sku(),
            barcode: self.barcode(),
            name: self.name.as_deref().unwrap_or_default(),
            attributes: self.attributes.clone().map(Value::Object),
            price: self.price.unwrap_or(Decimal::ZERO),
            standard_cost: self.standard_cost.unwrap_or(Decimal::ZERO),
            last_purchase_cost: self.last_purchase_cost.unwrap_or(Decimal::ZERO),
            is_default: self.is_default.unwrap_or(false) || position == 0,
            is_active: self.is_active.unwrap_or(true),
            sale_price: self.sale_price,
            sale_start: optional_time(self.sale_start.as_deref()),
            sale_end: optional_time(self.sale_end.as_deref()),
        }
    }
}

/// An empty or unparseable time is silently no time, as in Go.
fn optional_time(value: Option<&str>) -> Option<DateTime<Utc>> {
    DateTime::parse_from_rfc3339(value?)
        .ok()
        .map(|t| t.with_timezone(&Utc))
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct ProductRequest {
    name: Option<String>,
    category_id: Option<i64>,
    description: Option<String>,
    image_data: Option<String>,
    image_url: Option<String>,
    is_active: Option<bool>,
    variants: Option<Vec<VariantRequest>>,
}

impl ProductRequest {
    fn variants(&self) -> &[VariantRequest] {
        self.variants.as_deref().unwrap_or_default()
    }

    /// `CreateProductRequest` / `UpdateProductRequest` tags, diving into each variant.
    fn validate(&self) -> AppResult<()> {
        let non_negative = |v: Option<Decimal>| v.is_none_or(|d| d >= Decimal::ZERO);
        let mut v = Validator::new();
        v.field(
            "name",
            self.name.as_deref().unwrap_or_default(),
            &[Rule::Required, Rule::Min(1), Rule::Max(255)],
        );
        match &self.variants {
            None => {
                v.push("variants", "variants is required");
            }
            Some(variants) if variants.is_empty() => {
                v.push("variants", "variants must be at least 1 characters");
            }
            Some(variants) => {
                for variant in variants {
                    v.field(
                        "sku",
                        variant.sku(),
                        &[Rule::Required, Rule::Min(1), Rule::Max(100)],
                    )
                    .optional("barcode", variant.barcode.as_deref(), &[Rule::Max(100)])
                    .field(
                        "name",
                        variant.name.as_deref().unwrap_or_default(),
                        &[Rule::Required, Rule::Min(1), Rule::Max(255)],
                    )
                    .invalid_unless("price", non_negative(variant.price))
                    .invalid_unless("standard_cost", non_negative(variant.standard_cost))
                    .invalid_unless(
                        "last_purchase_cost",
                        non_negative(variant.last_purchase_cost),
                    );
                }
            }
        }
        v.finish()
    }

    fn input(&self, is_active: bool) -> ProductInput<'_> {
        ProductInput {
            category_id: self.category_id,
            name: self.name.as_deref().unwrap_or_default(),
            description: self.description.as_deref(),
            image_data: self.image_data.as_deref(),
            image_url: self.image_url.as_deref(),
            is_active,
        }
    }
}

/// Each barcode must be unique in the request and unused by the company's other
/// products. Pass product id 0 on create.
async fn check_barcodes(
    state: &AppState,
    company_id: i64,
    product_id: i64,
    variants: &[VariantRequest],
) -> AppResult<()> {
    let mut seen = HashSet::new();
    for barcode in variants.iter().filter_map(VariantRequest::barcode) {
        if !seen.insert(barcode) {
            return Err(AppError::bad_request(format!(
                "Duplicate barcode in request: {barcode}"
            )));
        }
        if db::barcode_in_other_product(&state.db, company_id, barcode, product_id).await? {
            return Err(AppError::bad_request(format!(
                "Barcode already registered: {barcode}"
            )));
        }
    }
    Ok(())
}

async fn check_category(
    state: &AppState,
    company_id: i64,
    category_id: Option<i64>,
) -> AppResult<()> {
    if let Some(id) = category_id
        && categories::get(&state.db, company_id, id).await?.is_none()
    {
        return Err(AppError::bad_request("Category not found"));
    }
    Ok(())
}

fn sku_taken(sku: &str) -> AppError {
    AppError::bad_request(format!("SKU already exists: {sku}"))
}

/// Best effort: a leftover object is far less bad than a failed request.
async fn delete_image(state: &AppState, url: &str) {
    let Some(storage) = &state.storage else {
        return;
    };
    let Some(key) = storage.key_from_url(url) else {
        return;
    };
    if let Err(err) = storage.delete(&key).await {
        tracing::warn!(%key, error = %err, "failed to delete an orphaned product image");
    }
}

async fn create(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<ProductRequest>,
) -> AppResult<Reply<ProductResponse>> {
    req.validate()?;
    let company_id = user.company_id;
    check_category(&state, company_id, req.category_id).await?;
    for variant in req.variants() {
        if db::sku_exists(&state.db, variant.sku()).await? {
            return Err(sku_taken(variant.sku()));
        }
    }
    check_barcodes(&state, company_id, 0, req.variants()).await?;

    let mut tx = state.db.begin().await?;
    let id = db::create_product(
        &mut *tx,
        company_id,
        &req.input(req.is_active.unwrap_or(true)),
    )
    .await?;
    for (position, variant) in req.variants().iter().enumerate() {
        db::create_variant(&mut *tx, id, &variant.input(position)).await?;
    }
    tx.commit().await?;

    let product = load(&state, company_id, id).await?;
    Ok(success(
        StatusCode::CREATED,
        "Product created successfully",
        product,
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
    JsonBody(req): JsonBody<ProductRequest>,
) -> AppResult<Reply<ProductResponse>> {
    let id = parse_id(&id, "product")?;
    req.validate()?;
    let company_id = user.company_id;
    let current = db::get(&state.db, company_id, id)
        .await?
        .ok_or_else(|| AppError::not_found("Product not found"))?;
    check_category(&state, company_id, req.category_id).await?;

    let existing = db::variants_of(&state.db, &[id])
        .await?
        .remove(&id)
        .unwrap_or_default();
    for variant in req.variants() {
        if db::sku_in_other_product(&state.db, variant.sku(), id).await? {
            return Err(sku_taken(variant.sku()));
        }
    }
    check_barcodes(&state, company_id, id, req.variants()).await?;

    // Variants naming one of this product's rows update it; the others reuse rows
    // that were left out, then insert. Rows still unused are deleted.
    let existing_ids: HashSet<i64> = existing.iter().map(|v| v.id).collect();
    let (explicit, new): (Vec<&VariantRequest>, Vec<&VariantRequest>) = req
        .variants()
        .iter()
        .partition(|v| v.id.is_some_and(|id| existing_ids.contains(&id)));
    let kept: HashSet<i64> = explicit.iter().filter_map(|v| v.id).collect();
    let mut orphans = existing.iter().filter(|v| !kept.contains(&v.id));

    let mut tx = state.db.begin().await?;
    let is_active = req.is_active.unwrap_or(current.is_active);
    db::update_product(&mut *tx, company_id, id, &req.input(is_active)).await?;

    for (position, variant) in explicit.iter().enumerate() {
        if let Some(variant_id) = variant.id {
            db::update_variant(&mut *tx, variant_id, &variant.input(position)).await?;
        }
    }
    for (offset, variant) in new.iter().enumerate() {
        let input = variant.input(explicit.len() + offset);
        match orphans.next() {
            Some(orphan) => db::update_variant(&mut *tx, orphan.id, &input).await?,
            None => db::create_variant(&mut *tx, id, &input).await?,
        }
    }
    for orphan in orphans {
        if let Err(err) = db::delete_variant(&mut *tx, orphan.id).await {
            return Err(variant_delete_error(err, orphan));
        }
    }
    tx.commit().await?;

    // The image was replaced or cleared: remove the old object.
    let new_image = req.image_url.as_deref().unwrap_or_default();
    if let Some(old_image) = current
        .image_url
        .as_deref()
        .filter(|old| !old.is_empty() && *old != new_image)
    {
        delete_image(&state, old_image).await;
    }

    let product = load(&state, company_id, id).await?;
    Ok(success(
        StatusCode::OK,
        "Product updated successfully",
        product,
    ))
}

/// A variant referenced by stock or order history can't be removed. Postgres reports
/// a foreign key as 23503, or 23001 for a RESTRICT one.
fn variant_delete_error(err: sqlx::Error, variant: &Variant) -> AppError {
    let code = err
        .as_database_error()
        .and_then(|e| e.code().map(|c| c.into_owned()));
    if matches!(code.as_deref(), Some("23503" | "23001")) {
        AppError::bad_request(format!(
            "Variant '{}' (SKU: {}) tidak dapat dihapus karena sudah memiliki riwayat transaksi",
            variant.name, variant.sku
        ))
    } else {
        err.into()
    }
}

async fn destroy(
    State(state): State<AppState>,
    user: AuthUser,
    RawId(id): RawId,
) -> AppResult<Reply<()>> {
    let id = parse_id(&id, "product")?;
    let company_id = user.company_id;
    let current = db::get(&state.db, company_id, id)
        .await?
        .ok_or_else(|| AppError::not_found("Product not found"))?;

    let mut tx = state.db.begin().await?;
    db::delete_variants_of(&mut *tx, id).await?;
    db::delete_product(&mut *tx, company_id, id).await?;
    tx.commit().await?;

    if let Some(image) = &current.image_url {
        delete_image(&state, image).await;
    }
    Ok(success_message(
        StatusCode::OK,
        "Product deleted successfully",
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
