-- Company settings table for configurable business settings

CREATE TABLE IF NOT EXISTS company_settings (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,

    -- Tax configuration
    tax_enabled BOOLEAN NOT NULL DEFAULT false,
    tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 0,
    tax_inclusive BOOLEAN NOT NULL DEFAULT true,

    -- Rounding configuration
    rounding_enabled BOOLEAN NOT NULL DEFAULT false,
    rounding_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,

    -- Order settings
    auto_complete_counter_orders BOOLEAN NOT NULL DEFAULT true,
    require_customer_for_delivery BOOLEAN NOT NULL DEFAULT true,

    -- Receipt settings
    receipt_header TEXT NULL,
    receipt_footer TEXT NULL,
    show_tax_on_receipt BOOLEAN NOT NULL DEFAULT true,

    -- Offline settings
    offline_mode_enabled BOOLEAN NOT NULL DEFAULT false,
    max_offline_days INT NOT NULL DEFAULT 7,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_company_settings_company_id UNIQUE (company_id)
);

CREATE INDEX idx_company_settings_company_id ON company_settings(company_id);

-- Insert default settings for existing companies
INSERT INTO company_settings (company_id)
SELECT id FROM companies
ON CONFLICT (company_id) DO NOTHING;

COMMENT ON TABLE company_settings IS 'Per-company configuration settings';
COMMENT ON COLUMN company_settings.tax_rate IS 'Tax percentage (e.g., 11 for 11% PPN)';
COMMENT ON COLUMN company_settings.tax_inclusive IS 'If true, prices already include tax. If false, tax added on top.';
COMMENT ON COLUMN company_settings.rounding_amount IS 'Round totals to nearest value (e.g., 100 for nearest Rp100)';
COMMENT ON COLUMN company_settings.auto_complete_counter_orders IS 'Auto-complete counter orders when fully paid';
