CREATE TABLE IF NOT EXISTS product_categories (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES product_categories(id) ON DELETE SET NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_product_categories_company_code ON product_categories(company_id, code);
CREATE INDEX idx_product_categories_company_id ON product_categories(company_id);
CREATE INDEX idx_product_categories_parent_id ON product_categories(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_product_categories_is_active ON product_categories(company_id, is_active);
