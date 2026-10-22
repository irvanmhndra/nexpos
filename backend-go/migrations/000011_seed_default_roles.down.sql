-- Revert to original 5-role system
DELETE FROM role_permissions;

-- Remove staff role
DELETE FROM roles WHERE code = 'staff';

-- Re-add removed roles
INSERT INTO roles (company_id, code, name, description, is_system) VALUES
(NULL, 'manager', 'Manager', 'Can manage daily operations and view reports', true),
(NULL, 'cashier', 'Cashier', 'Can process orders and basic operations', true),
(NULL, 'inventory', 'Inventory Staff', 'Can manage inventory and stock', true);

-- Re-seed all role permissions (same as migration 000008)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'owner';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'admin' AND p.code NOT IN ('settings.update');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'manager' AND p.code IN (
    'users.view', 'products.view', 'products.create', 'products.update',
    'inventory.view', 'inventory.adjust',
    'orders.view', 'orders.create', 'orders.void',
    'reports.view', 'reports.export', 'branches.view'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'cashier' AND p.code IN (
    'products.view', 'inventory.view', 'orders.view', 'orders.create'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'inventory' AND p.code IN (
    'products.view', 'products.create', 'products.update',
    'inventory.view', 'inventory.adjust'
);
