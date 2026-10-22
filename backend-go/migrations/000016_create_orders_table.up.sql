CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    order_no VARCHAR(50) NOT NULL,
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    cashier_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total_discount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total_tax DECIMAL(15, 2) NOT NULL DEFAULT 0,
    grand_total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    applied_promotions JSONB,
    notes TEXT,
    is_offline BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_orders_company_order_no ON orders(company_id, order_no);
CREATE INDEX idx_orders_company_id ON orders(company_id);
CREATE INDEX idx_orders_branch_id ON orders(branch_id);
CREATE INDEX idx_orders_customer_id ON orders(customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX idx_orders_cashier_id ON orders(cashier_id);
CREATE INDEX idx_orders_status ON orders(company_id, status);
CREATE INDEX idx_orders_created_at ON orders(company_id, created_at DESC);
