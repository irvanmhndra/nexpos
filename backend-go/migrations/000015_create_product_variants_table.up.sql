CREATE TABLE IF NOT EXISTS product_variants (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    attributes JSONB,
    price DECIMAL(15, 2) NOT NULL DEFAULT 0,
    standard_cost DECIMAL(15, 2) NOT NULL DEFAULT 0,
    last_purchase_cost DECIMAL(15, 2) NOT NULL DEFAULT 0,
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_product_variants_sku ON product_variants(sku);
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_is_active ON product_variants(product_id, is_active);
CREATE INDEX idx_product_variants_is_default ON product_variants(product_id, is_default);
