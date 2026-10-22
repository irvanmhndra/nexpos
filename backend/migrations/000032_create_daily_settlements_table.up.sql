CREATE TABLE IF NOT EXISTS daily_settlements (
    id              BIGSERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    settlement_date DATE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    total_sales     DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_refunds   DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_expenses  DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    recorded_by     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    finalized_by    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    finalized_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, branch_id, settlement_date),
    CONSTRAINT chk_settlement_status CHECK (status IN ('draft','finalized'))
);
CREATE INDEX idx_daily_settlements_company_id ON daily_settlements(company_id);
CREATE INDEX idx_daily_settlements_branch_id ON daily_settlements(branch_id);
CREATE INDEX idx_daily_settlements_date ON daily_settlements(settlement_date);

CREATE TABLE IF NOT EXISTS daily_settlement_items (
    id                    BIGSERIAL PRIMARY KEY,
    daily_settlement_id   BIGINT NOT NULL REFERENCES daily_settlements(id) ON DELETE CASCADE,
    payment_method        VARCHAR(50) NOT NULL,
    expected_amount       DECIMAL(15,2) NOT NULL DEFAULT 0,
    actual_amount         DECIMAL(15,2),
    variance_amount       DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes                 TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(daily_settlement_id, payment_method)
);
CREATE INDEX idx_daily_settlement_items_settlement_id ON daily_settlement_items(daily_settlement_id);
