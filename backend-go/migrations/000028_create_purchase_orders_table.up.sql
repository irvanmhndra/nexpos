CREATE TABLE IF NOT EXISTS purchase_orders (
    id           BIGSERIAL PRIMARY KEY,
    company_id   BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id    BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    supplier_id  BIGINT NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
    po_number    VARCHAR(50) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'draft',
    notes        TEXT,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    ordered_at   TIMESTAMPTZ,
    received_at  TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, po_number),
    CONSTRAINT chk_po_status CHECK (status IN ('draft','ordered','partial','received','cancelled'))
);
CREATE INDEX idx_purchase_orders_company_id ON purchase_orders(company_id);
CREATE INDEX idx_purchase_orders_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id                  BIGSERIAL PRIMARY KEY,
    purchase_order_id   BIGINT NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_variant_id  BIGINT REFERENCES product_variants(id) ON DELETE SET NULL,
    sku                 VARCHAR(100) NOT NULL,
    variant_name        VARCHAR(255) NOT NULL,
    quantity            INT NOT NULL DEFAULT 0,
    received_quantity   INT NOT NULL DEFAULT 0,
    unit_cost           DECIMAL(15,2) NOT NULL DEFAULT 0,
    subtotal            DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_po_items_purchase_order_id ON purchase_order_items(purchase_order_id);
