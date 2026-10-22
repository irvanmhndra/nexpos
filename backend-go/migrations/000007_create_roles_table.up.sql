-- Roles are company-scoped
CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT REFERENCES companies(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- company_id NULL = system role (available to all companies)
CREATE UNIQUE INDEX idx_roles_company_code ON roles(company_id, code) WHERE company_id IS NOT NULL;
CREATE UNIQUE INDEX idx_roles_system_code ON roles(code) WHERE company_id IS NULL;
CREATE INDEX idx_roles_company_id ON roles(company_id);

-- Seed system roles (company_id = NULL means available to all)
INSERT INTO roles (company_id, code, name, description, is_system) VALUES
(NULL, 'owner', 'Owner', 'Full access to all features', true),
(NULL, 'admin', 'Administrator', 'Can manage users, products, and settings', true),
(NULL, 'manager', 'Manager', 'Can manage daily operations and view reports', true),
(NULL, 'cashier', 'Cashier', 'Can process orders and basic operations', true),
(NULL, 'inventory', 'Inventory Staff', 'Can manage inventory and stock', true);
