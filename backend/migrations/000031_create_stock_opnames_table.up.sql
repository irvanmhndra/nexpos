CREATE TABLE IF NOT EXISTS stock_opnames (
    id                   BIGSERIAL PRIMARY KEY,
    company_id           BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id            BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    opname_number        VARCHAR(50) NOT NULL,
    status               VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    notes                TEXT,
    total_variance_qty   INT NOT NULL DEFAULT 0,
    total_variance_value DECIMAL(15,2) NOT NULL DEFAULT 0,
    started_by           BIGINT REFERENCES users(id) ON DELETE SET NULL,
    completed_by         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    started_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at         TIMESTAMPTZ,
    cancelled_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, opname_number),
    CONSTRAINT chk_opname_status CHECK (status IN ('in_progress','completed','cancelled'))
);
CREATE INDEX idx_stock_opnames_company_id ON stock_opnames(company_id);
CREATE INDEX idx_stock_opnames_branch_id ON stock_opnames(branch_id);
CREATE INDEX idx_stock_opnames_status ON stock_opnames(status);

CREATE TABLE IF NOT EXISTS stock_opname_items (
    id                  BIGSERIAL PRIMARY KEY,
    stock_opname_id     BIGINT NOT NULL REFERENCES stock_opnames(id) ON DELETE CASCADE,
    product_variant_id  BIGINT NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    sku                 VARCHAR(100) NOT NULL,
    product_name        VARCHAR(255) NOT NULL,
    variant_name        VARCHAR(255) NOT NULL,
    system_stock        INT NOT NULL DEFAULT 0,
    counted_stock       INT,
    variance_qty        INT NOT NULL DEFAULT 0,
    unit_cost           DECIMAL(15,2) NOT NULL DEFAULT 0,
    variance_value      DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes               TEXT,
    counted_by          BIGINT REFERENCES users(id) ON DELETE SET NULL,
    counted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(stock_opname_id, product_variant_id)
);
CREATE INDEX idx_stock_opname_items_opname_id ON stock_opname_items(stock_opname_id);
CREATE INDEX idx_stock_opname_items_variant_id ON stock_opname_items(product_variant_id);
