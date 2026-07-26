-- ============================================
-- Sample Data Insertion Script: Setup Tenants and Configurations
-- File: 1_setup_tenants_and_configs.sql
-- Description: This script creates sample tenants and their configurations
-- ============================================

-- Enable row-level security temporarily for tenant-specific tables
SET LOCAL row_security = off;

-- ============================================
-- Create sample tenants
-- ============================================
INSERT INTO tenants (tenant_id, name, domain, status) VALUES
    ('tenant_1', 'Burger Haven', 'burgerhaven.example.com', 'active'),
    ('tenant_2', 'Sushi Palace', 'sushipalace.example.com', 'active'),
    ('tenant_3', 'Veggie Delight', 'veggiedelight.example.com', 'active'),
    ('tenant_4', 'Italian Eats', 'italianeats.example.com', 'active'),
    ('tenant_5', 'Spicy Wings', 'spicywings.example.com', 'active'),
    ('tenant_6', 'Healthy Bites', 'healthybites.example.com', 'active'),
    ('tenant_7', 'Dessert Den', 'dessertden.example.com', 'active'),
    ('tenant_8', 'Breakfast Club', 'breakfastclub.example.com', 'active');



-- Verify tenant insertion
SELECT COUNT(*) AS total_tenants FROM tenants WHERE status = 'active';

-- ============================================
-- Create sample tenant configurations
-- ============================================
INSERT INTO tenant_configurations (tenant_id, default_currency, default_language, tax_rate, delivery_fee) VALUES
    -- Burger Haven
    ('tenant_1', 'USD', 'en', 0.08, 2.99),
    -- Sushi Palace
    ('tenant_2', 'JPY', 'ja', 0.05, 0.00),
    -- Veggie Delight
    ('tenant_3', 'USD', 'en', 0.07, 1.99),
    -- Italian Eats
    ('tenant_4', 'EUR', 'it', 0.22, 3.99),
    -- Spicy Wings
    ('tenant_5', 'USD', 'en', 0.09, 2.49),
    -- Healthy Bites
    ('tenant_6', 'USD', 'en', 0.06, 1.49),
    -- Dessert Den
    ('tenant_7', 'USD', 'en', 0.08, 1.99),
    -- Breakfast Club
    ('tenant_8', 'USD', 'en', 0.07, 1.99);

-- Verify configuration insertion
SELECT t.tenant_id, t.name, tc.default_currency, tc.default_language, tc.tax_rate, tc.delivery_fee
FROM tenants t
JOIN tenant_configurations tc ON t.tenant_id = tc.tenant_id
ORDER BY t.tenant_id;

-- ============================================
-- Create sample order statuses (shared across all tenants)
-- ============================================
INSERT INTO order_statuses (status_id, name, description, is_terminal) VALUES
    ('status_draft', 'draft', 'Order created but not yet confirmed', FALSE),
    ('status_pending', 'pending', 'Order received, waiting for payment', FALSE),
    ('status_processing', 'processing', 'Order being prepared', FALSE),
    ('status_out_for_delivery', 'out_for_delivery', 'Order on its way', FALSE),
    ('status_delivered', 'delivered', 'Order delivered to customer', TRUE),
    ('status_cancelled', 'cancelled', 'Order cancelled by customer', TRUE),
    ('status_failed', 'failed', 'Order processing failed', TRUE),
    ('status_completed', 'completed', 'Order fully processed', TRUE);

-- Verify order status insertion
SELECT COUNT(*) AS total_statuses FROM order_statuses;

-- ============================================
-- Create sample payment methods (shared across all tenants)
-- ============================================
INSERT INTO payment_methods (method_id, name, description) VALUES
    ('pm_credit_card', 'credit_card', 'Credit card payment'),
    ('pm_debit_card', 'debit_card', 'Debit card payment'),
    ('pm_paypal', 'paypal', 'PayPal payment'),
    ('pm_bank_transfer', 'bank_transfer', 'Bank transfer'),
    ('pm_cash_on_delivery', 'cash_on_delivery', 'Cash on delivery'),
    ('pm_gift_card', 'gift_card', 'Gift card payment');

-- Verify payment method insertion
SELECT COUNT(*) AS total_payment_methods FROM payment_methods;

-- ============================================
-- Create sample roles
-- ============================================
INSERT INTO user_roles (role_id, tenant_id, name, description) VALUES
    -- Admin roles (tenant-specific)
    ('role_burger_haven_admin','tenant_1', 'burger_haven_admin', 'Full access for Burger Haven'),
    ('role_sushi_palace_admin','tenant_2', 'sushi_palace_admin', 'Full access for Sushi Palace'),
    ('role_veggie_delight_admin','tenant_3', 'veggie_delight_admin', 'Full access for Veggie Delight'),
    ('role_italian_eats_admin','tenant_4', 'italian_eats_admin', 'Full access for Italian Eats'),
    ('role_spicy_wings_admin','tenant_5', 'spicy_wings_admin', 'Full access for Spicy Wings'),
    ('role_healthy_bites_admin','tenant_6', 'healthy_bites_admin', 'Full access for Healthy Bites'),
    ('role_dessert_den_admin','tenant_7', 'dessert_den_admin', 'Full access for Dessert Den'),
    ('role_breakfast_club_admin','tenant_8', 'breakfast_club_admin', 'Full access for Breakfast Club'),

    -- Shared roles for tenant_1 (examples)
    ('role_restaurant_manager_t1','tenant_1', 'restaurant_manager', 'Restaurant operations manager'),
    ('role_menu_editor_t1','tenant_1', 'menu_editor', 'Menu content editor'),
    ('role_inventory_manager_t1','tenant_1', 'inventory_manager', 'Inventory control'),
    ('role_customer_support_t1','tenant_1', 'customer_support', 'Customer service representative'),
    ('role_finance_officer_t1','tenant_1', 'finance_officer', 'Financial transactions');

-- Verify role insertion
SELECT COUNT(*) AS total_roles FROM user_roles;

-- ============================================
-- Re-enable row-level security
-- ============================================
SET LOCAL row_security = on;