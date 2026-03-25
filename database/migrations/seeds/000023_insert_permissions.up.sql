-- Permissions seed (granular CRUD: {operation}_{resource})
INSERT INTO permissions (name, description) VALUES
    -- Products
    ('create_product', 'Create products'),
    ('update_product', 'Update products'),
    ('delete_product', 'Delete products'),
    -- Categories
    ('create_category', 'Create categories'),
    ('update_category', 'Update categories'),
    ('delete_category', 'Delete categories'),
    -- Orders
    ('view_orders', 'View orders'),
    ('manage_orders', 'Manage order status'),
    -- Staff
    ('view_staff', 'View staff members'),
    ('create_staff', 'Add staff members'),
    ('update_staff', 'Update staff members'),
    ('delete_staff', 'Delete staff members'),
    -- Coupons
    ('view_coupons', 'View coupons'),
    ('create_coupon', 'Create coupons'),
    ('update_coupon', 'Update coupons'),
    ('delete_coupon', 'Delete coupons'),
    -- Shop
    ('update_shop', 'Update shop settings'),
    ('delete_shop', 'Delete the shop'),
    -- Metrics
    ('view_metrics', 'View dashboard metrics'),
    -- Billing
    ('manage_billing', 'Manage billing and subscription');

-- Role permissions
-- Owner: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'owner';

-- Admin: all except delete_shop and manage_billing
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
  AND p.name NOT IN ('delete_shop', 'manage_billing');

-- Encargado: products, categories, orders, coupons, metrics (no staff, no shop settings)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'encargado'
  AND p.name IN (
    'create_product', 'update_product', 'delete_product',
    'create_category', 'update_category', 'delete_category',
    'view_orders', 'manage_orders',
    'view_staff',
    'view_coupons', 'create_coupon', 'update_coupon', 'delete_coupon',
    'view_metrics'
  );

-- Operador: view and manage orders only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'operador'
  AND p.name IN ('view_orders', 'manage_orders');

-- Delivery: view orders only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'delivery'
  AND p.name IN ('view_orders');
