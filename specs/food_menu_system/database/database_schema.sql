-- Food Menu System schema SQL
-- Generated from specs/food_menu_system/database/database_schema.md
-- PostgreSQL 14+

-- Schemas / Settings
SET client_min_messages = WARNING;

-- First create the database (required for Docker entrypoint)
CREATE DATABASE homecooked;

\c homecooked

-- Set search_path to use custom schema first
  SET search_path = tenant_management, public;


-- 2. Core Tables
CREATE TABLE IF NOT EXISTS tenants (
    tenant_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active','suspended','deleted')),
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS tenant_configurations (
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE PRIMARY KEY,
    default_currency VARCHAR(3) DEFAULT 'USD',
    default_language VARCHAR(10) DEFAULT 'en',
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    delivery_fee DECIMAL(10,2) DEFAULT 0.00,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);

-- 3. User Management

CREATE TABLE IF NOT EXISTS user_roles (
    role_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);


CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    created_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL,
    updated_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Spec-aligned (singular) tables

CREATE TABLE IF NOT EXISTS tenant_restaurants (
    restaurant_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address_line1 VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    created_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tenant_restaurant_categories (
    category_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    restaurant_id VARCHAR(50) REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_category_id VARCHAR(50) REFERENCES tenant_restaurant_categories(category_id) ON DELETE SET NULL,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    created_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tenant_menu_items (
    menu_item_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    restaurant_id VARCHAR(50) REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    category_id VARCHAR(50) REFERENCES tenant_restaurant_categories(category_id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    base_price DECIMAL(10, 2) NOT NULL,  
    price DECIMAL(10, 2) NOT NULL, 
    currency VARCHAR(3) DEFAULT 'USD',
    is_active BOOLEAN DEFAULT TRUE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    created_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tenant_menu_item_images (
    image_id VARCHAR(50) PRIMARY KEY,
    menu_item_id VARCHAR(50) REFERENCES tenant_menu_items(menu_item_id) ON DELETE CASCADE,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    url VARCHAR(512) NOT NULL,
    alt_text VARCHAR(255),
    is_primary BOOLEAN DEFAULT FALSE,
    display_order INTEGER DEFAULT 0,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);


CREATE TABLE IF NOT EXISTS food_category (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);

CREATE TABLE IF NOT EXISTS food_item (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    category_id VARCHAR(50) REFERENCES food_category(id) ON DELETE RESTRICT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);

CREATE TABLE IF NOT EXISTS menu_item (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    food_item_id VARCHAR(50) REFERENCES food_item(id) ON DELETE CASCADE,
    menu_id VARCHAR(50) NOT NULL,
    menu_type VARCHAR(20) NOT NULL,
    description VARCHAR(100),
    size VARCHAR(50),
    price DECIMAL(10,2) NOT NULL,
    sequence INTEGER NOT NULL,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);

-- 6. Order Statuses
CREATE TABLE IF NOT EXISTS order_statuses (
    status_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);

CREATE INDEX IF NOT EXISTS idx_order_statuses_name ON order_statuses(name);

-- 7. Orders
CREATE TABLE IF NOT EXISTS tenant_orders (
    order_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    user_id VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL,
    restaurant_id VARCHAR(50) REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    status_id VARCHAR(50) REFERENCES order_statuses(status_id) ON DELETE SET NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    tax_amount DECIMAL(10,2) DEFAULT 0.00,
    delivery_fee DECIMAL(10,2) DEFAULT 0.00,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    subtotal DECIMAL(12,2) NOT NULL,
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'pending',
    tracking_number VARCHAR(100),
    delivery_address_line1 VARCHAR(255) NOT NULL,
    delivery_address_line2 VARCHAR(255),
    delivery_city VARCHAR(100) NOT NULL,
    delivery_state VARCHAR(100) NOT NULL,
    delivery_postal_code VARCHAR(20) NOT NULL,
    delivery_country VARCHAR(100) NOT NULL,
    delivery_instructions TEXT,
    pickup_location VARCHAR(255),
    estimated_delivery_time BIGINT,
    actual_delivery_time BIGINT,
    notes TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    created_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL,
    updated_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tenant_order_items (
    order_item_id VARCHAR(50) PRIMARY KEY,
    order_id VARCHAR(50) REFERENCES tenant_orders(order_id) ON DELETE CASCADE,
    menu_item_id VARCHAR(50) REFERENCES tenant_menu_items(menu_item_id) ON DELETE SET NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    total_price DECIMAL(10,2) NOT NULL,
    notes TEXT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);

-- 8. Payments
CREATE TABLE IF NOT EXISTS payment_methods (
    method_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS tenant_payments (
    payment_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    order_id VARCHAR(50) REFERENCES tenant_orders(order_id) ON DELETE SET NULL,
    payment_method_id VARCHAR(50) REFERENCES payment_methods(method_id) ON DELETE SET NULL,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    transaction_id VARCHAR(100),
    payment_gateway VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',
    payment_date BIGINT,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);

-- 9. Reviews
CREATE TABLE IF NOT EXISTS tenant_reviews (
    review_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    user_id VARCHAR(50) REFERENCES users(user_id) ON DELETE CASCADE,
    restaurant_id VARCHAR(50) REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    menu_item_id VARCHAR(50) REFERENCES tenant_menu_items(menu_item_id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title VARCHAR(255),
    review_text TEXT,
    is_approved BOOLEAN DEFAULT FALSE,
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT,
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW()))::BIGINT
);

-- 10. Indexes
CREATE INDEX IF NOT EXISTS idx_tenants_domain ON tenants(domain);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_tenant_restaurants_tenant_id ON tenant_restaurants(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_restaurants_name ON tenant_restaurants(name);
CREATE INDEX IF NOT EXISTS idx_tenant_menu_items_tenant_id ON tenant_menu_items(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_menu_items_restaurant_id ON tenant_menu_items(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_menu_items_tenant_id ON tenant_menu_items(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_menu_items_restaurant_id ON tenant_menu_items(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_orders_tenant_id ON tenant_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_orders_user_id ON tenant_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_orders_restaurant_id ON tenant_orders(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_orders_status_id ON tenant_orders(status_id);
CREATE INDEX IF NOT EXISTS idx_tenant_payments_tenant_id ON tenant_payments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_payments_order_id ON tenant_payments(order_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_id_name ON user_roles(tenant_id, name);

-- 11. Triggers: update_timestamp
CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = (EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000)::BIGINT;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach triggers (safe: only create triggers if tables exist)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'tenants') THEN
        EXECUTE 'CREATE TRIGGER update_tenants_timestamp
        BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_timestamp();';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'users') THEN
        EXECUTE 'CREATE TRIGGER update_users_timestamp
        BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamp();';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'tenant_restaurants') THEN
        EXECUTE 'CREATE TRIGGER update_restaurants_timestamp
        BEFORE UPDATE ON tenant_restaurants FOR EACH ROW EXECUTE FUNCTION update_timestamp();';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'tenant_menu_items') THEN
        EXECUTE 'CREATE TRIGGER update_menu_items_timestamp
        BEFORE UPDATE ON tenant_menu_items FOR EACH ROW EXECUTE FUNCTION update_timestamp();';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'tenant_orders') THEN
        EXECUTE 'CREATE TRIGGER update_orders_timestamp
        BEFORE UPDATE ON tenant_orders FOR EACH ROW EXECUTE FUNCTION update_timestamp();';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'tenant_payments') THEN
        EXECUTE 'CREATE TRIGGER update_payments_timestamp
        BEFORE UPDATE ON tenant_payments FOR EACH ROW EXECUTE FUNCTION update_timestamp();';
    END IF;
END;
$$;

-- 12. Functions: tenant/user context (return VARCHAR(50))
CREATE OR REPLACE FUNCTION get_current_tenant_id() RETURNS VARCHAR(50) AS $$
DECLARE
    tenant_id VARCHAR(50);
BEGIN
    tenant_id := current_setting('current_tenant_id', true);
    RETURN tenant_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE FUNCTION get_current_user_id() RETURNS VARCHAR(50) AS $$
DECLARE
    user_id VARCHAR(50);
BEGIN
    user_id := current_setting('current_user_id', true);
    RETURN user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- View for tenant context
CREATE OR REPLACE VIEW current_tenant_context AS
SELECT
    current_setting('current_tenant_id', true) AS tenant_id,
    current_setting('current_user_id', true) AS user_id,
    (SELECT name FROM tenants WHERE tenant_id = current_setting('current_tenant_id', true)) AS tenant_name;

-- 13. Row-Level Security: enable and policies
-- Enable RLS on tenant-specific tables
ALTER TABLE IF EXISTS tenant_restaurants ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_restaurant_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_menu_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_menu_item_images ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_inventory_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_inventory_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_inventory_stock ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_order_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_reviews ENABLE ROW LEVEL SECURITY;

CREATE OR REPLACE FUNCTION get_current_tenant_id()
RETURNS VARCHAR(50) AS $$
BEGIN
      RETURN current_setting('app.current_tenant_id');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Policies: use explicit casting to VARCHAR(50)

DO $$
BEGIN
   IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'tenant_restaurants') THEN
      EXECUTE 'CREATE POLICY tenant_restaurants_policy ON tenant_restaurants FOR ALL USING (tenant_id = get_current_tenant_id()::VARCHAR(50));';
   END IF;
   IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'tenant_menu_items') THEN
      EXECUTE 'CREATE POLICY tenant_menu_items_policy ON tenant_menu_items FOR ALL USING (tenant_id = get_current_tenant_id()::VARCHAR(50));';
   END IF;
   IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'tenant_orders') THEN
     EXECUTE 'CREATE POLICY tenant_orders_policy ON tenant_orders FOR ALL USING (tenant_id = get_current_tenant_id()::VARCHAR(50));';
   END IF;
END;
$$;

INSERT INTO tenants (tenant_id, name, domain, status) VALUES
('tenant_1', 'Restaurant Chain 1', 'restaurant1.example.com', 'active')
ON CONFLICT (tenant_id) DO NOTHING;

INSERT INTO tenants (tenant_id, name, domain, status) VALUES
('tenant_2', 'Restaurant Chain 2', 'restaurant2.example.com', 'active')
ON CONFLICT (tenant_id) DO NOTHING;

INSERT INTO tenants (tenant_id, name, domain, status) VALUES
('tenant_3', 'Restaurant Chain 3', 'restaurant3.example.com', 'active')
ON CONFLICT (tenant_id) DO NOTHING;

INSERT INTO tenant_configurations (tenant_id, default_currency, default_language, tax_rate, delivery_fee)
VALUES
('tenant_1', 'USD', 'en', 0.08, 2.99)
ON CONFLICT (tenant_id) DO NOTHING;

INSERT INTO tenant_configurations (tenant_id, default_currency, default_language, tax_rate, delivery_fee)
VALUES
('tenant_2', 'EUR', 'fr', 0.20, 3.99)
ON CONFLICT (tenant_id) DO NOTHING;

INSERT INTO tenant_configurations (tenant_id, default_currency, default_language, tax_rate, delivery_fee)
VALUES
('tenant_3', 'USD', 'es', 0.07, 1.99)
ON CONFLICT (tenant_id) DO NOTHING;

-- Detailed sample data for core domain tables
-- Roles
INSERT INTO user_roles (role_id, tenant_id, name, description) 
VALUES
('role_admin_t1', 'tenant_1', 'admin', 'Tenant administrator'),
('role_staff_t1', 'tenant_1', 'staff', 'Restaurant staff')
ON CONFLICT (role_id) DO NOTHING;

-- Users
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, is_admin) VALUES
('user_1', 'tenant_1', 'owner1@restaurant1.example.com', 'changeme_hash', 'Owner', 'One', TRUE),
('user_2', 'tenant_1', 'staff1@restaurant1.example.com', 'changeme_hash', 'Staff', 'One', FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Restaurants
INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, email, phone, address_line1, city, state, postal_code, country, is_active, created_by) VALUES
('restaurant_123', 'tenant_1', 'Resto 123', 'contact@resto123.example.com', '+1234567890', '123 Main St', 'Townsville', 'State', '12345', 'Country', TRUE, 'user_1')
ON CONFLICT (restaurant_id) DO NOTHING;

-- Restaurant categories
INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, created_by) VALUES
('cat_1', 'tenant_1', 'restaurant_123', 'Mains', 'Main dishes', 1, 'user_1'),
('cat_2', 'tenant_1', 'restaurant_123', 'Sides', 'Side dishes', 2, 'user_1')
ON CONFLICT (category_id) DO NOTHING;

-- Menu items (tenant-scoped)
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, created_by) VALUES
('menu_item_1', 'tenant_1', 'restaurant_123', 'cat_1', 'Grilled Chicken', 'Juicy grilled chicken', 10.00, 12.00, 'USD', TRUE, 'user_1'),
('menu_item_2', 'tenant_1', 'restaurant_123', 'cat_2', 'Fries', 'Crispy fries', 2.00, 3.00, 'USD', TRUE, 'user_1')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Spec-aligned categories and items (shared/singular)
INSERT INTO food_category (id, tenant_id, name, description,is_active) VALUES
('fc_1', 'tenant_1', 'vegetarian', 'Vegetarian based items', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO food_category (id, tenant_id, name, description,is_active) VALUES
('fc_2', 'tenant_1', 'non-vegetarian', 'Non-vegetarian based items', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO food_category (id, tenant_id, name, description,is_active) VALUES
('fc_3', 'tenant_1', 'vegan', 'Vegan based items', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO food_category (id, tenant_id, name, description,is_active) VALUES
('fc_4', 'tenant_1', 'gluten-free', 'Gluten-free based items', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO food_category (id, tenant_id, name, description,is_active) VALUES
('fc_5', 'tenant_1', 'dairy-free', 'Dairy-free based items', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO food_category (id, tenant_id, name, description,is_active) VALUES
('fc_6', 'tenant_1', 'jain', 'Jain diet items', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO food_item (id, tenant_id, name, description, price, category_id) VALUES
('fi_1', 'tenant_1', 'Grilled Chicken', 'Delicious grilled chicken', 12.00, 'fc_1')
ON CONFLICT (id) DO NOTHING;

INSERT INTO menu_item (id, tenant_id, food_item_id, menu_id, menu_type, description, size, price, sequence) VALUES
('mi_1', 'tenant_1', 'fi_1', 'weekly_1', 'weekly', 'Grilled chicken - regular', 'regular', 12.00, 1)
ON CONFLICT (id) DO NOTHING;

-- Order statuses and payment methods
INSERT INTO order_statuses (status_id, name, description, is_active) VALUES
('status_new', 'new', 'New order', FALSE),
('status_preparing', 'preparing', 'Preparing', FALSE),
('status_completed', 'completed', 'Completed', TRUE)
ON CONFLICT (status_id) DO NOTHING;

INSERT INTO payment_methods (method_id, name, description) VALUES
('pm_card', 'card', 'Credit/Debit Card'),
('pm_cash', 'cash', 'Cash on delivery')
ON CONFLICT (method_id) DO NOTHING;

-- Example payment record
INSERT INTO tenant_payments (payment_id, tenant_id, order_id, payment_method_id, amount, currency, transaction_id, status) VALUES
('payment_1', 'tenant_1', NULL, 'pm_card', 12.00, 'USD', 'txn_123', 'completed')
ON CONFLICT (payment_id) DO NOTHING;

-- 15. Sample queries (for reference)
-- SELECT * FROM tenant_restaurants WHERE tenant_id = 'tenant_1' AND is_active = TRUE;

-- End of schema
