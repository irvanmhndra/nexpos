-- Junction table for role-permission mapping
CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_role_permissions_unique ON role_permissions(role_id, permission_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Seed default role permissions
-- Owner gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'owner';

-- Admin gets all except settings.update
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'admin' AND p.code NOT IN ('settings.update');

-- Manager gets view/create/update permissions + reports
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'manager' AND p.code IN (
    'users.view', 'products.view', 'products.create', 'products.update',
    'inventory.view', 'inventory.adjust',
    'orders.view', 'orders.create', 'orders.void',
    'reports.view', 'reports.export',
    'branches.view'
);

-- Cashier gets basic POS operations
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'cashier' AND p.code IN (
    'products.view', 'inventory.view',
    'orders.view', 'orders.create'
);

-- Inventory staff gets inventory-focused permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'inventory' AND p.code IN (
    'products.view', 'products.create', 'products.update',
    'inventory.view', 'inventory.adjust'
);
