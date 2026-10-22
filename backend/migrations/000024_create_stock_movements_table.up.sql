CREATE TABLE IF NOT EXISTS stock_movements (
    id BIGSERIAL PRIMARY KEY,
    product_variant_id BIGINT NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('IN', 'OUT', 'ADJUST', 'TRANSFER')),
    quantity INT NOT NULL,
    stock_before INT NOT NULL DEFAULT 0,
    stock_after INT NOT NULL DEFAULT 0,
    unit_cost DECIMAL(15, 2),
    reference_type VARCHAR(50),
    reference_id BIGINT,
    note TEXT,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stock_movements_variant_id ON stock_movements(product_variant_id);
CREATE INDEX idx_stock_movements_branch_id ON stock_movements(branch_id);
CREATE INDEX idx_stock_movements_type ON stock_movements(type);
CREATE INDEX idx_stock_movements_created_at ON stock_movements(created_at DESC);
