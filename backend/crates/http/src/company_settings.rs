//! `/api/v1/company-settings`: same responses as the Go
//! `CompanySettingsHandler`/`CompanySettingsService`.

use axum::{Router, extract::State, http::StatusCode, routing::get};
use chrono::{DateTime, Utc};
use nexpos_db::company_settings::{self as db, CompanySettings};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

use crate::{
    AppResult, AppState,
    envelope::{Reply, success},
    extract::{AuthUser, JsonBody, Validator},
};

pub fn router() -> Router<AppState> {
    Router::new().route("/api/v1/company-settings", get(show).put(update))
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct UpdateRequest {
    tax_enabled: Option<bool>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    tax_rate: Option<Decimal>,
    tax_inclusive: Option<bool>,
    rounding_enabled: Option<bool>,
    #[serde(deserialize_with = "crate::money::deserialize_opt")]
    rounding_amount: Option<Decimal>,
    auto_complete_counter_orders: Option<bool>,
    require_customer_for_delivery: Option<bool>,
    receipt_header: Option<String>,
    receipt_footer: Option<String>,
    show_tax_on_receipt: Option<bool>,
    offline_mode_enabled: Option<bool>,
    max_offline_days: Option<i32>,
}

#[derive(Serialize)]
struct SettingsResponse {
    id: i64,
    company_id: i64,
    tax_enabled: bool,
    #[serde(serialize_with = "crate::money::serialize")]
    tax_rate: Decimal,
    tax_inclusive: bool,
    rounding_enabled: bool,
    #[serde(serialize_with = "crate::money::serialize")]
    rounding_amount: Decimal,
    auto_complete_counter_orders: bool,
    require_customer_for_delivery: bool,
    receipt_header: Option<String>,
    receipt_footer: Option<String>,
    show_tax_on_receipt: bool,
    offline_mode_enabled: bool,
    max_offline_days: i32,
    #[serde(serialize_with = "crate::time::serialize")]
    created_at: DateTime<Utc>,
    #[serde(serialize_with = "crate::time::serialize")]
    updated_at: DateTime<Utc>,
}

impl From<CompanySettings> for SettingsResponse {
    fn from(s: CompanySettings) -> Self {
        Self {
            id: s.id,
            company_id: s.company_id,
            tax_enabled: s.tax_enabled,
            tax_rate: s.tax_rate,
            tax_inclusive: s.tax_inclusive,
            rounding_enabled: s.rounding_enabled,
            rounding_amount: s.rounding_amount,
            auto_complete_counter_orders: s.auto_complete_counter_orders,
            require_customer_for_delivery: s.require_customer_for_delivery,
            receipt_header: s.receipt_header,
            receipt_footer: s.receipt_footer,
            show_tax_on_receipt: s.show_tax_on_receipt,
            offline_mode_enabled: s.offline_mode_enabled,
            max_offline_days: s.max_offline_days,
            created_at: s.created_at,
            updated_at: s.updated_at,
        }
    }
}

/// What a company has before its settings are first saved: id 0 and Go's zero time.
pub(crate) fn defaults(company_id: i64) -> CompanySettings {
    CompanySettings {
        id: 0,
        company_id,
        tax_enabled: false,
        tax_rate: Decimal::ZERO,
        tax_inclusive: true,
        rounding_enabled: false,
        rounding_amount: Decimal::ZERO,
        auto_complete_counter_orders: true,
        require_customer_for_delivery: true,
        receipt_header: None,
        receipt_footer: None,
        show_tax_on_receipt: true,
        offline_mode_enabled: false,
        max_offline_days: 7,
        created_at: crate::time::zero(),
        updated_at: crate::time::zero(),
    }
}

async fn show(State(state): State<AppState>, user: AuthUser) -> AppResult<Reply<SettingsResponse>> {
    let settings = db::get(&state.db, user.company_id)
        .await?
        .unwrap_or_else(|| defaults(user.company_id));
    Ok(success(
        StatusCode::OK,
        "Company settings retrieved successfully",
        settings.into(),
    ))
}

async fn update(
    State(state): State<AppState>,
    user: AuthUser,
    JsonBody(req): JsonBody<UpdateRequest>,
) -> AppResult<Reply<SettingsResponse>> {
    Validator::new()
        .invalid_unless(
            "tax_rate",
            req.tax_rate
                .is_none_or(|v| v >= Decimal::ZERO && v <= Decimal::ONE_HUNDRED),
        )
        .invalid_unless(
            "rounding_amount",
            req.rounding_amount.is_none_or(|v| v >= Decimal::ZERO),
        )
        .invalid_unless(
            "max_offline_days",
            req.max_offline_days.is_none_or(|v| (1..=30).contains(&v)),
        )
        .finish()?;

    let mut s = db::get(&state.db, user.company_id)
        .await?
        .unwrap_or_else(|| defaults(user.company_id));

    // Only the fields that were sent change; `null` means "not sent".
    s.tax_enabled = req.tax_enabled.unwrap_or(s.tax_enabled);
    s.tax_rate = req.tax_rate.unwrap_or(s.tax_rate);
    s.tax_inclusive = req.tax_inclusive.unwrap_or(s.tax_inclusive);
    s.rounding_enabled = req.rounding_enabled.unwrap_or(s.rounding_enabled);
    s.rounding_amount = req.rounding_amount.unwrap_or(s.rounding_amount);
    s.auto_complete_counter_orders = req
        .auto_complete_counter_orders
        .unwrap_or(s.auto_complete_counter_orders);
    s.require_customer_for_delivery = req
        .require_customer_for_delivery
        .unwrap_or(s.require_customer_for_delivery);
    s.receipt_header = req.receipt_header.or(s.receipt_header);
    s.receipt_footer = req.receipt_footer.or(s.receipt_footer);
    s.show_tax_on_receipt = req.show_tax_on_receipt.unwrap_or(s.show_tax_on_receipt);
    s.offline_mode_enabled = req.offline_mode_enabled.unwrap_or(s.offline_mode_enabled);
    s.max_offline_days = req.max_offline_days.unwrap_or(s.max_offline_days);

    // Like Go, the response keeps the decimals as sent, before the column rounds them.
    (s.id, s.created_at, s.updated_at) = db::upsert(&state.db, &s).await?;
    Ok(success(
        StatusCode::OK,
        "Company settings updated successfully",
        s.into(),
    ))
}
