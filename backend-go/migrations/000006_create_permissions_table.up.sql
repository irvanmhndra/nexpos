-- Permissions are system-defined (global, not per company)
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    module VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_permissions_code ON permissions(code);
CREATE INDEX idx_permissions_module ON permissions(module);

-- Seed default permissions
INSERT INTO permissions (code, name, module, description) VALUES
-- User Management
('users.view', 'View Users', 'users', 'Can view user list and details'),
('users.create', 'Create Users', 'users', 'Can create new users'),
('users.update', 'Update Users', 'users', 'Can update user information'),
('users.delete', 'Delete Users', 'users', 'Can delete users'),

-- Product Management
('products.view', 'View Products', 'products', 'Can view product list and details'),
('products.create', 'Create Products', 'products', 'Can create new products'),
('products.update', 'Update Products', 'products', 'Can update product information'),
('products.delete', 'Delete Products', 'products', 'Can delete products'),

-- Inventory Management
('inventory.view', 'View Inventory', 'inventory', 'Can view stock levels'),
('inventory.adjust', 'Adjust Inventory', 'inventory', 'Can adjust stock quantities'),

-- Order Management
('orders.view', 'View Orders', 'orders', 'Can view order list and details'),
('orders.create', 'Create Orders', 'orders', 'Can create new orders (cashier)'),
('orders.void', 'Void Orders', 'orders', 'Can void/cancel orders'),
('orders.refund', 'Refund Orders', 'orders', 'Can process refunds'),

-- Reports
('reports.view', 'View Reports', 'reports', 'Can view reports and analytics'),
('reports.export', 'Export Reports', 'reports', 'Can export reports'),

-- Settings
('settings.view', 'View Settings', 'settings', 'Can view company settings'),
('settings.update', 'Update Settings', 'settings', 'Can update company settings'),

-- Branch Management
('branches.view', 'View Branches', 'branches', 'Can view branch list'),
('branches.manage', 'Manage Branches', 'branches', 'Can create/update/delete branches');
