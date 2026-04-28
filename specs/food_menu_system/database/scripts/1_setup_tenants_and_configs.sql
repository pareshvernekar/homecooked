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
INSERT INTO tenants (name, domain, status) VALUES
    ('Burger Haven', 'burgerhaven.example.com', 'active'),
    ('Sushi Palace', 'sushipalace.example.com', 'active'),
    ('Veggie Delight', 'veggiedelight.example.com', 'active'),
    ('Italian Eats', 'italianeats.example.com', 'active'),
    ('Spicy Wings', 'spicywings.example.com', 'active'),
    ('Healthy Bites', 'healthybites.example.com', 'active'),
    ('Dessert Den', 'dessertden.example.com', 'active'),
    ('Breakfast Club', 'breakfastclub.example.com', 'active');

-- Verify tenant insertion
SELECT COUNT(*) AS total_tenants FROM tenants WHERE status = 'active';

-- ============================================
-- Create sample tenant configurations
-- ============================================
INSERT INTO tenant_configurations (tenant_id, default_currency, default_language, tax_rate, delivery_fee) VALUES
    -- Burger Haven
    (1, 'USD', 'en', 0.08, 2.99),
    -- Sushi Palace
    (2, 'JPY', 'ja', 0.05, 0.00),
    -- Veggie Delight
    (3, 'USD', 'en', 0.07, 1.99),
    -- Italian Eats
    (4, 'EUR', 'it', 0.22, 3.99),
    -- Spicy Wings
    (5, 'USD', 'en', 0.09, 2.49),
    -- Healthy Bites
    (6, 'USD', 'en', 0.06, 1.49),
    -- Dessert Den
    (7, 'USD', 'en', 0.08, 1.99),
    -- Breakfast Club
    (8, 'USD', 'en', 0.07, 1.99);

-- Verify configuration insertion
SELECT t.tenant_id, t.name, tc.default_currency, tc.default_language, tc.tax_rate, tc.delivery_fee
FROM tenants t
JOIN tenant_configurations tc ON t.tenant_id = tc.tenant_id
ORDER BY t.tenant_id;

-- ============================================
-- Create sample order statuses (shared across all tenants)
-- ============================================
INSERT INTO order_statuses (name, description, is_terminal) VALUES
    ('draft', 'Order created but not yet confirmed', FALSE),
    ('pending', 'Order received, waiting for payment', FALSE),
    ('processing', 'Order being prepared', FALSE),
    ('out_for_delivery', 'Order on its way', FALSE),
    ('delivered', 'Order delivered to customer', TRUE),
    ('cancelled', 'Order cancelled by customer', TRUE),
    ('failed', 'Order processing failed', TRUE),
    ('completed', 'Order fully processed', TRUE);

-- Verify order status insertion
SELECT COUNT(*) AS total_statuses FROM order_statuses;

-- ============================================
-- Create sample payment methods (shared across all tenants)
-- ============================================
INSERT INTO payment_methods (name, description) VALUES
    ('credit_card', 'Credit card payment'),
    ('debit_card', 'Debit card payment'),
    ('paypal', 'PayPal payment'),
    ('bank_transfer', 'Bank transfer'),
    ('cash_on_delivery', 'Cash on delivery'),
    ('gift_card', 'Gift card payment');

-- Verify payment method insertion
SELECT COUNT(*) AS total_payment_methods FROM payment_methods;

-- ============================================
-- Create sample roles
-- ============================================
INSERT INTO user_roles (tenant_id, name, description) VALUES
    -- Admin roles (tenant-specific)
    (1, 'burger_haven_admin', 'Full access for Burger Haven'),
    (2, 'sushi_palace_admin', 'Full access for Sushi Palace'),
    (3, 'veggie_delight_admin', 'Full access for Veggie Delight'),
    (4, 'italian_eats_admin', 'Full access for Italian Eats'),
    (5, 'spicy_wings_admin', 'Full access for Spicy Wings'),
    (6, 'healthy_bites_admin', 'Full access for Healthy Bites'),
    (7, 'dessert_den_admin', 'Full access for Dessert Den'),
    (8, 'breakfast_club_admin', 'Full access for Breakfast Club'),
    
    -- Shared roles
    (1, 'restaurant_manager', 'Restaurant operations manager'),
    (1, 'menu_editor', 'Menu content editor'),
    (1, 'inventory_manager', 'Inventory control'),
    (1, 'customer_support', 'Customer service representative'),
    (1, 'finance_officer', 'Financial transactions');

-- Verify role insertion
SELECT COUNT(*) AS total_roles FROM user_roles;

-- ============================================
-- Re-enable row-level security
-- ============================================
SET LOCAL row_security = on;