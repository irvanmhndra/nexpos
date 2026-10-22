CREATE TABLE IF NOT EXISTS expense_categories (
    id          BIGSERIAL PRIMARY KEY,
    company_id  BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, name)
);
CREATE INDEX idx_expense_categories_company_id ON expense_categories(company_id);

CREATE TABLE IF NOT EXISTS expenses (
    id           BIGSERIAL PRIMARY KEY,
    company_id   BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id    BIGINT REFERENCES branches(id) ON DELETE SET NULL,
    category_id  BIGINT REFERENCES expense_categories(id) ON DELETE SET NULL,
    amount       DECIMAL(15,2) NOT NULL,
    description  VARCHAR(500) NOT NULL,
    reference_no VARCHAR(100),
    expense_date DATE NOT NULL,
    recorded_by  BIGINT REFERENCES users(id) ON DELETE SET NULL,
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_expenses_company_id ON expenses(company_id);
CREATE INDEX idx_expenses_branch_id ON expenses(branch_id);
CREATE INDEX idx_expenses_expense_date ON expenses(expense_date);
CREATE INDEX idx_expenses_category_id ON expenses(category_id);
