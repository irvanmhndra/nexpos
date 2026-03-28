CREATE TABLE IF NOT EXISTS suppliers (
    id           BIGSERIAL PRIMARY KEY,
    company_id   BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name         VARCHAR(200) NOT NULL,
    code         VARCHAR(50) NOT NULL,
    contact_name VARCHAR(100),
    phone        VARCHAR(50),
    email        VARCHAR(150),
    address      TEXT,
    notes        TEXT,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);
CREATE INDEX idx_suppliers_company_id ON suppliers(company_id);
