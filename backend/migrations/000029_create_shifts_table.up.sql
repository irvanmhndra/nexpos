CREATE TABLE IF NOT EXISTS shifts (
    id              BIGSERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    cashier_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status          VARCHAR(20) NOT NULL DEFAULT 'open',
    opening_float   DECIMAL(15,2) NOT NULL DEFAULT 0,
    closing_float   DECIMAL(15,2) NOT NULL DEFAULT 0,
    expected_cash   DECIMAL(15,2) NOT NULL DEFAULT 0,
    actual_cash     DECIMAL(15,2),
    cash_difference DECIMAL(15,2),
    notes           TEXT,
    opened_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_shift_status CHECK (status IN ('open','closed'))
);
CREATE INDEX idx_shifts_company_id ON shifts(company_id);
CREATE INDEX idx_shifts_branch_id ON shifts(branch_id);
CREATE INDEX idx_shifts_cashier_id ON shifts(cashier_id);
CREATE INDEX idx_shifts_status ON shifts(status);
