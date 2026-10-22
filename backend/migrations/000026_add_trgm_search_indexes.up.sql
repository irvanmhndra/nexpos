CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_products_name_trgm
    ON products USING GIN (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_product_variants_sku_trgm
    ON product_variants USING GIN (sku gin_trgm_ops);
