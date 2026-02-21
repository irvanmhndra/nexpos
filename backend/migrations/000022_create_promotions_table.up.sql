CREATE TABLE promotions (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- discount | bundle | conditional
    discount_type VARCHAR(50), -- percentage | fixed
    discount_value DECIMAL(10,2),
    min_purchase DECIMAL(10,2),
    max_discount DECIMAL(10,2),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ,
    priority INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_promotions_company_code ON promotions(company_id, code);
CREATE INDEX idx_promotions_company_id ON promotions(company_id);
CREATE INDEX idx_promotions_active ON promotions(is_active, start_at, end_at);
CREATE INDEX idx_promotions_type ON promotions(type);

COMMENT ON TABLE promotions IS 'Promotional campaigns and discounts';
COMMENT ON COLUMN promotions.type IS 'Type: discount, bundle, conditional';
COMMENT ON COLUMN promotions.discount_type IS 'For type=discount: percentage or fixed amount';
COMMENT ON COLUMN promotions.discount_value IS 'Discount amount (% or fixed)';
COMMENT ON COLUMN promotions.min_purchase IS 'Minimum purchase required to apply promo';
COMMENT ON COLUMN promotions.max_discount IS 'Maximum discount amount (for percentage type)';
COMMENT ON COLUMN promotions.priority IS 'Higher priority promos apply first';
