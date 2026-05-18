-- Food Menu System Database Schema
-- Version: 1.0
-- Last Updated: 2026-05-02

-- Create Tenants Table
CREATE TABLE tenants (
    tenant_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Create Tenant Configurations Table
CREATE TABLE tenant_configurations (
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE PRIMARY KEY,
    default_currency VARCHAR(3) DEFAULT 'USD',
    default_language VARCHAR(10) DEFAULT 'en',
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    delivery_fee DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create Users Table
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    profile_picture_url VARCHAR(512),
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE
);

-- Create User Roles Table
CREATE TABLE user_roles (
    role_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create User Role Assignments Table
CREATE TABLE user_role_assignments (
    assignment_id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES user_roles(role_id) ON DELETE CASCADE,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create Restaurants Table
CREATE TABLE tenant_restaurants (
    restaurant_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255) NOT NULL,
    website_url VARCHAR(255),
    logo_url VARCHAR(512),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create Restaurant Categories Table
CREATE TABLE tenant_restaurant_categories (
    category_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    restaurant_id INTEGER REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_category_id INTEGER REFERENCES tenant_restaurant_categories(category_id) ON DELETE SET NULL,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create Menu Items Table
CREATE TABLE tenant_menu_items (
    menu_item_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    restaurant_id INTEGER REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES tenant_restaurant_categories(category_id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_price DECIMAL(10,2) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    is_vegetarian BOOLEAN DEFAULT FALSE,
    is_vegan BOOLEAN DEFAULT FALSE,
    is_gluten_free BOOLEAN DEFAULT FALSE,
    is_dairy_free BOOLEAN DEFAULT FALSE,
    preparation_time_minutes INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create Menu Item Images Table
CREATE TABLE tenant_menu_item_images (
    image_id SERIAL PRIMARY KEY,
    menu_item_id INTEGER REFERENCES tenant_menu_items(menu_item_id) ON DELETE CASCADE,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    url VARCHAR(512) NOT NULL,
    alt_text VARCHAR(255),
    is_primary BOOLEAN DEFAULT FALSE,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create Inventory Categories Table
CREATE TABLE tenant_inventory_categories (
    category_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_category_id INTEGER REFERENCES tenant_inventory_categories(category_id) ON DELETE SET NULL,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create Inventory Items Table
CREATE TABLE tenant_inventory_items (
    item_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES tenant_inventory_categories(category_id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sku VARCHAR(50) UNIQUE,
    unit VARCHAR(50) NOT NULL,
    cost_price DECIMAL(10,2) NOT NULL,
    selling_price DECIMAL(10,2),
    supplier_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create Inventory Stock Table
CREATE TABLE tenant_inventory_stock (
    stock_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    item_id INTEGER REFERENCES tenant_inventory_items(item_id) ON DELETE CASCADE,
    location_id INTEGER,
    quantity DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    reorder_threshold DECIMAL(15,2) DEFAULT 0.00,
    last_stock_take_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create Order Statuses Table
CREATE TABLE order_statuses (
    status_id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_terminal BOOLEAN DEFAULT FALSE
);

-- Create Orders Table
CREATE TABLE tenant_orders (
    order_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(user_id) ON DELETE SET NULL,
    restaurant_id INTEGER REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    status_id INTEGER REFERENCES order_statuses(status_id) ON DELETE SET NULL,
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
    estimated_delivery_time TIMESTAMP WITH TIME ZONE,
    actual_delivery_time TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create Order Items Table
CREATE TABLE tenant_order_items (
    order_item_id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES tenant_orders(order_id) ON DELETE CASCADE,
    menu_item_id INTEGER REFERENCES tenant_menu_items(menu_item_id) ON DELETE SET NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    total_price DECIMAL(10,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create Payment Methods Table
CREATE TABLE payment_methods (
    method_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

-- Create Payments Table
CREATE TABLE tenant_payments (
    payment_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    order_id INTEGER REFERENCES tenant_orders(order_id) ON DELETE SET NULL,
    payment_method_id INTEGER REFERENCES payment_methods(method_id) ON DELETE SET NULL,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    transaction_id VARCHAR(100),
    payment_gateway VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',
    payment_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create Reviews Table
CREATE TABLE tenant_reviews (
    review_id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
    restaurant_id INTEGER REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    menu_item_id INTEGER REFERENCES tenant_menu_items(menu_item_id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title VARCHAR(255),
    review_text TEXT,
    is_approved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create Indexes for Core Tables

-- Tenants
CREATE INDEX idx_tenants_domain ON tenants(domain);
CREATE INDEX idx_tenants_status ON tenants(status);

-- Users
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(is_active);

-- Restaurants
CREATE INDEX idx_tenant_restaurants_tenant_id ON tenant_restaurants(tenant_id);
CREATE INDEX idx_tenant_restaurants_name ON tenant_restaurants(name);
CREATE INDEX idx_tenant_restaurants_status ON tenant_restaurants(is_active);

-- Menu Items
CREATE INDEX idx_tenant_menu_items_tenant_id ON tenant_menu_items(tenant_id);
CREATE INDEX idx_tenant_menu_items_restaurant_id ON tenant_menu_items(restaurant_id);
CREATE INDEX idx_tenant_menu_items_category_id ON tenant_menu_items(category_id);
CREATE INDEX idx_tenant_menu_items_name ON tenant_menu_items(name);
CREATE INDEX idx_tenant_menu_items_status ON tenant_menu_items(is_active);

-- Orders
CREATE INDEX idx_tenant_orders_tenant_id ON tenant_orders(tenant_id);
CREATE INDEX idx_tenant_orders_user_id ON tenant_orders(user_id);
CREATE INDEX idx_tenant_orders_restaurant_id ON tenant_orders(restaurant_id);
CREATE INDEX idx_tenant_orders_status ON tenant_orders(status_id);
CREATE INDEX idx_tenant_orders_order_number ON tenant_orders(order_number);
CREATE INDEX idx_tenant_orders_created_at ON tenant_orders(created_at);
CREATE INDEX idx_tenant_orders_status_active ON tenant_orders(tenant_id, status_id, is_active) WHERE is_active = TRUE;

-- Order Items
CREATE INDEX idx_tenant_order_items_order_id ON tenant_order_items(order_id);
CREATE INDEX idx_tenant_order_items_menu_item_id ON tenant_order_items(menu_item_id);

-- Payments
CREATE INDEX idx_tenant_payments_tenant_id ON tenant_payments(tenant_id);
CREATE INDEX idx_tenant_payments_order_id ON tenant_payments(order_id);
CREATE INDEX idx_tenant_payments_status ON tenant_payments(status);



-- Enable Row Level Security
ALTER TABLE IF EXISTS public.tenant_restaurants ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_restaurant_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_menu_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_menu_item_images ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_inventory_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_inventory_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_inventory_stock ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_order_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.tenant_reviews ENABLE ROW LEVEL SECURITY;

-- Create Function to Update Timestamps
CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;$$ LANGUAGE plpgsql;

-- Apply Triggers to Update Timestamps
CREATE TRIGGER IF NOT EXISTS update_tenants_timestamp BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER IF NOT EXISTS update_users_timestamp BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER IF NOT EXISTS update_restaurants_timestamp BEFORE UPDATE ON tenant_restaurants FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER IF NOT EXISTS update_menu_items_timestamp BEFORE UPDATE ON tenant_menu_items FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER IF NOT EXISTS update_orders_timestamp BEFORE UPDATE ON tenant_orders FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER IF NOT EXISTS update_payments_timestamp BEFORE UPDATE ON tenant_payments FOR EACH ROW EXECUTE FUNCTION update_timestamp();


-- Create Views for Tenant-Specific Data

-- Restaurant Menu View
CREATE VIEW IF NOT EXISTS tenant_restaurant_menu AS
SELECT
    r.restaurant_id,
    r.name AS restaurant_name,
    r.logo_url,
    c.category_id,
    c.name AS category_name,
    mi.menu_item_id,
    mi.name AS menu_item_name,
    mi.description,
    mi.base_price,
    mi.price,
    mi.is_active,
    mi.is_featured,
    mi.is_vegetarian,
    mi.is_vegan,
    mi.is_gluten_free,
    mi.is_dairy_free,
    mi.preparation_time_minutes,
    mi.created_at,
    mi.updated_at
FROM
    tenant_restaurants r
JOIN
    tenant_restaurant_categories c ON r.restaurant_id = c.restaurant_id
JOIN
    tenant_menu_items mi ON c.category_id = mi.category_id
WHERE
    r.tenant_id = current_setting('app.current_tenant_id')::INTEGER
    AND mi.is_active = TRUE
ORDER BY
    r.restaurant_id, c.display_order, mi.display_order;

-- User Orders View
CREATE VIEW IF NOT EXISTS user_orders AS
SELECT
    o.order_id,
    o.order_number,
    o.total_amount,
    o.tax_amount,
    o.delivery_fee,
    o.discount_amount,
    o.subtotal,
    o.payment_method,
    o.payment_status,
    o.status_id,
    s.name AS status_name,
    o.created_at,
    o.updated_at,
    o.created_by,
    u.email AS created_by_email,
    u.first_name || ' ' || u.last_name AS created_by_name,
    oi.menu_item_id,
    mi.name AS menu_item_name,
    oi.quantity,
    oi.unit_price,
    oi.discount_amount,
    oi.total_price
FROM
    tenant_orders o
JOIN
    order_statuses s ON o.status_id = s.status_id
LEFT JOIN
    tenant_order_items oi ON o.order_id = oi.order_id
LEFT JOIN
    tenant_menu_items mi ON oi.menu_item_id = mi.menu_item_id
WHERE
    o.tenant_id = current_setting('app.current_tenant_id')::INTEGER
    AND (o.user_id = current_setting('app.current_user_id')::INTEGER OR o.created_by IS NULL)
ORDER BY
    o.created_at DESC;

-- Create Functions for Tenant-Specific Operations

-- Function to Get Current Tenant ID
CREATE OR REPLACE FUNCTION get_current_tenant_id() RETURNS INTEGER AS $$
DECLARE
    tenant_id INTEGER;
BEGIN