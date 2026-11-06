//! `company_settings`, like the Go `CompanySettingsRepository`. A company has no
//! row until its settings are first saved.

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::PgExecutor;

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct CompanySettings {
    pub id: i64,
    pub company_id: i64,
    pub tax_enabled: bool,
    pub tax_rate: Decimal,
    pub tax_inclusive: bool,
    pub rounding_enabled: bool,
    pub rounding_amount: Decimal,
    pub auto_complete_counter_orders: bool,
    pub require_customer_for_delivery: bool,
    pub receipt_header: Option<String>,
    pub receipt_footer: Option<String>,
    pub show_tax_on_receipt: bool,
    pub offline_mode_enabled: bool,
    pub max_offline_days: i32,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

pub async fn get(
    db: impl PgExecutor<'_>,
    company_id: i64,
) -> sqlx::Result<Option<CompanySettings>> {
    sqlx::query_as(
        "SELECT id, company_id, tax_enabled, tax_rate, tax_inclusive, rounding_enabled, rounding_amount, \
             auto_complete_counter_orders, require_customer_for_delivery, receipt_header, receipt_footer, \
             show_tax_on_receipt, offline_mode_enabled, max_offline_days, created_at, updated_at \
         FROM company_settings WHERE company_id = $1",
    )
    .bind(company_id)
    .fetch_optional(db)
    .await
}

/// Saves the settings and returns the row's `(id, created_at, updated_at)`.
pub async fn upsert(
    db: impl PgExecutor<'_>,
    s: &CompanySettings,
) -> sqlx::Result<(i64, DateTime<Utc>, DateTime<Utc>)> {
    sqlx::query_as(
        "INSERT INTO company_settings (company_id, tax_enabled, tax_rate, tax_inclusive, rounding_enabled, \
             rounding_amount, auto_complete_counter_orders, require_customer_for_delivery, receipt_header, \
             receipt_footer, show_tax_on_receipt, offline_mode_enabled, max_offline_days) \
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) \
         ON CONFLICT (company_id) DO UPDATE SET \
             tax_enabled = EXCLUDED.tax_enabled, tax_rate = EXCLUDED.tax_rate, \
             tax_inclusive = EXCLUDED.tax_inclusive, rounding_enabled = EXCLUDED.rounding_enabled, \
             rounding_amount = EXCLUDED.rounding_amount, \
             auto_complete_counter_orders = EXCLUDED.auto_complete_counter_orders, \
             require_customer_for_delivery = EXCLUDED.require_customer_for_delivery, \
             receipt_header = EXCLUDED.receipt_header, receipt_footer = EXCLUDED.receipt_footer, \
             show_tax_on_receipt = EXCLUDED.show_tax_on_receipt, \
             offline_mode_enabled = EXCLUDED.offline_mode_enabled, \
             max_offline_days = EXCLUDED.max_offline_days, updated_at = NOW() \
         RETURNING id, created_at, updated_at",
    )
    .bind(s.company_id)
    .bind(s.tax_enabled)
    .bind(s.tax_rate)
    .bind(s.tax_inclusive)
    .bind(s.rounding_enabled)
    .bind(s.rounding_amount)
    .bind(s.auto_complete_counter_orders)
    .bind(s.require_customer_for_delivery)
    .bind(&s.receipt_header)
    .bind(&s.receipt_footer)
    .bind(s.show_tax_on_receipt)
    .bind(s.offline_mode_enabled)
    .bind(s.max_offline_days)
    .fetch_one(db)
    .await
}
