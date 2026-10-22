-- Reset roles to simplified 3-role system: owner, admin, staff

-- First, delete existing role_permissions for roles we're removing
DELETE FROM role_permissions WHERE role_id IN (
    SELECT id FROM roles WHERE code IN ('manager', 'cashier', 'inventory')
);

-- Delete old roles (keep owner, admin; remove manager, cashier, inventory)
DELETE FROM roles WHERE code IN ('manager', 'cashier', 'inventory');

-- Update admin role description
UPDATE roles SET
    name = 'Administrator',
    description = 'Can manage users, products, inventory, and daily operations'
WHERE code = 'admin';

-- Insert staff role if not exists
INSERT INTO roles (company_id, code, name, description, is_system)
SELECT NULL, 'staff', 'Staff', 'Basic access for daily operations (cashier, inventory)', true
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'staff' AND company_id IS NULL);

-- Clear and re-seed role_permissions for clean state
DELETE FROM role_permissions;

-- Owner gets ALL permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'owner';

-- Admin gets all except settings.update and users.delete
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'admin'
  AND p.code NOT IN ('settings.update', 'users.delete');

-- Staff gets basic operational permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'staff'
  AND p.code IN (
    'products.view',
    'inventory.view',
    'orders.view',
    'orders.create'
  );
