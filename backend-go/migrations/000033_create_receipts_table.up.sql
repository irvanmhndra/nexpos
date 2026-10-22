CREATE TABLE IF NOT EXISTS receipts (
    id              BIGSERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    order_id        BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_no        VARCHAR(50) NOT NULL,
    cashier_id      BIGINT NOT NULL,
    cashier_name    VARCHAR(255) NOT NULL,
    customer_id     BIGINT,
    customer_name   VARCHAR(255),
    items           JSONB NOT NULL DEFAULT '[]'::jsonb,
    payments        JSONB NOT NULL DEFAULT '[]'::jsonb,
    total_amount    DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_discount  DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_tax       DECIMAL(15,2) NOT NULL DEFAULT 0,
    grand_total     DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    completed_at    TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, order_id)
);
CREATE INDEX idx_receipts_company_id ON receipts(company_id);
CREATE INDEX idx_receipts_order_id ON receipts(order_id);
