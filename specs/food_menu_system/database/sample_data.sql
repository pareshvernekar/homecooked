-- Sample data seed for development
-- Run with: psql -d your_db -f sample_data.sql

-- Use transaction for idempotent seeding
BEGIN;

-- Set session tenant/user (used by get_current_* functions)
SELECT set_config('current_tenant_id', 'tenant_1', true);
SELECT set_config('current_user_id', 'user_1', true);

-- Tenants and configurations
INSERT INTO tenants (tenant_id, name, domain, status) VALUES
('tenant_1', 'Restaurant Chain 1', 'restaurant1.example.com', 'active')
ON CONFLICT (tenant_id) DO NOTHING;

INSERT INTO tenant_configurations (tenant_id, default_currency, default_language, tax_rate, delivery_fee) VALUES
('tenant_1', 'USD', 'en', 0.08, 2.99)
ON CONFLICT (tenant_id) DO NOTHING;

-- Roles
INSERT INTO user_roles (role_id, tenant_id, name, description) VALUES
('role_admin_t1', 'tenant_1', 'admin', 'Tenant administrator'),
('role_staff_t1', 'tenant_1', 'staff', 'Restaurant staff')
ON CONFLICT (role_id) DO NOTHING;

-- Users
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, is_admin) VALUES
('user_1', 'tenant_1', 'owner1@restaurant1.example.com', 'changeme_hash', 'Owner', 'One', TRUE),
('user_2', 'tenant_1', 'staff1@restaurant1.example.com', 'changeme_hash', 'Staff', 'One', FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Assign roles
INSERT INTO user_role_assignments (assignment_id, user_id, role_id, tenant_id) VALUES
('ura_1', 'user_1', 'role_admin_t1', 'tenant_1'),
('ura_2', 'user_2', 'role_staff_t1', 'tenant_1')
ON CONFLICT (assignment_id) DO NOTHING;

-- Restaurants
INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, email, phone, address_line1, city, state, postal_code, country, is_active, created_by) VALUES
('restaurant_123', 'tenant_1', 'Resto 123', 'contact@resto123.example.com', '+1234567890', '123 Main St', 'Townsville', 'State', '12345', 'Country', TRUE, 'user_1')
ON CONFLICT (restaurant_id) DO NOTHING;

-- Restaurant categories
INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, created_by) VALUES
('cat_1', 'tenant_1', 'restaurant_123', 'Mains', 'Main dishes', 1, 'user_1'),
('cat_2', 'tenant_1', 'restaurant_123', 'Sides', 'Side dishes', 2, 'user_1')
ON CONFLICT (category_id) DO NOTHING;

-- Tenant menu items
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, created_by) VALUES
('menu_item_1', 'tenant_1', 'restaurant_123', 'cat_1', 'Grilled Chicken', 'Juicy grilled chicken', 10.00, 12.00, 'USD', TRUE, 'user_1'),
('menu_item_2', 'tenant_1', 'restaurant_123', 'cat_2', 'Fries', 'Crispy fries', 2.00, 3.00, 'USD', TRUE, 'user_1')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Spec-aligned shared categories/items
INSERT INTO food_category (id, tenant_id, name, description) VALUES
('fc_1', 'tenant_1', 'Proteins', 'Protein based items')
ON CONFLICT (id) DO NOTHING;

INSERT INTO food_item (id, tenant_id, name, description, price, category_id) VALUES
('fi_1', 'tenant_1', 'Grilled Chicken', 'Delicious grilled chicken', 12.00, 'fc_1')
ON CONFLICT (id) DO NOTHING;

INSERT INTO menu_item (id, tenant_id, food_item_id, menu_id, menu_type, description, size, price, sequence) VALUES
('mi_1', 'tenant_1', 'fi_1', 'weekly_1', 'weekly', 'Grilled chicken - regular', 'regular', 12.00, 1)
ON CONFLICT (id) DO NOTHING;

-- Order statuses
INSERT INTO order_statuses (status_id, name, description, is_terminal) VALUES
('status_new', 'new', 'New order', FALSE),
('status_preparing', 'preparing', 'Preparing', FALSE),
('status_completed', 'completed', 'Completed', TRUE)
ON CONFLICT (status_id) DO NOTHING;

-- Payment methods
INSERT INTO payment_methods (method_id, name, description) VALUES
('pm_card', 'card', 'Credit/Debit Card'),
('pm_cash', 'cash', 'Cash on delivery')
ON CONFLICT (method_id) DO NOTHING;

-- Create a sample order and items
INSERT INTO tenant_orders (order_id, tenant_id, user_id, restaurant_id, status_id, order_number, total_amount, subtotal, payment_method, payment_status, created_by) VALUES
('order_1', 'tenant_1', 'user_2', 'restaurant_123', 'status_new', 'ORD-1001', 15.00, 15.00, 'card', 'pending', 'user_2')
ON CONFLICT (order_id) DO NOTHING;

INSERT INTO tenant_order_items (order_item_id, order_id, menu_item_id, quantity, unit_price, total_price) VALUES
('order_item_1', 'order_1', 'menu_item_1', 1, 12.00, 12.00),
('order_item_2', 'order_1', 'menu_item_2', 1, 3.00, 3.00)
ON CONFLICT (order_item_id) DO NOTHING;

-- Payment record for order
INSERT INTO tenant_payments (payment_id, tenant_id, order_id, payment_method_id, amount, currency, transaction_id, status) VALUES
('payment_1', 'tenant_1', 'order_1', 'pm_card', 15.00, 'USD', 'txn_1001', 'completed')
ON CONFLICT (payment_id) DO NOTHING;

-- Reviews
INSERT INTO tenant_reviews (review_id, tenant_id, user_id, restaurant_id, menu_item_id, rating, title, review_text, is_approved) VALUES
('review_1', 'tenant_1', 'user_2', 'restaurant_123', 'menu_item_1', 5, 'Great', 'Really enjoyed the chicken', TRUE)
ON CONFLICT (review_id) DO NOTHING;

-- Inventory sample
INSERT INTO tenant_inventory_categories (category_id, tenant_id, name, description, created_by) VALUES
('inv_cat_1', 'tenant_1', 'Produce', 'Fruits and vegetables', 'user_1')
ON CONFLICT (category_id) DO NOTHING;

INSERT INTO tenant_inventory_items (item_id, tenant_id, category_id, name, sku, unit, cost_price, selling_price, created_by) VALUES
('inv_item_1', 'tenant_1', 'inv_cat_1', 'Lettuce', 'SKU-LETTUCE', 'head', 0.50, 1.00, 'user_1')
ON CONFLICT (item_id) DO NOTHING;

INSERT INTO tenant_inventory_stock (stock_id, tenant_id, item_id, quantity) VALUES
('stock_1', 'tenant_1', 'inv_item_1', 100.00)
ON CONFLICT (stock_id) DO NOTHING;

COMMIT;

-- End of sample data
