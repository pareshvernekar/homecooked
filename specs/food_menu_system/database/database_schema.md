# Food Menu System Database Schema

## 1. Overview
This document defines the complete database schema for the food menu system with multi-tenancy support. The schema includes all tables, relationships, indexes, and constraints required for a fully functional food menu system.

### 1.1 Schema Architecture
- **Database**: PostgreSQL 14+ (recommended)
- **Multi-tenancy**: Database-per-tenant approach
- **Schema**: Single schema with tenant-aware tables
- **Character Set**: UTF-8
- **Collation**: en_US.UTF-8

### 1.2 Table Structure
All tables follow this naming convention:
- `tenant_<table_name>` for tenant-specific data
- `<table_name>` for shared data
- `tenant_id` column for tenant association

## 2. Core Tables

### 2.1 Tenants
```sql
CREATE TABLE tenants (
    tenant_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);
```

### 2.2 Tenant Configuration
```sql
CREATE TABLE tenant_configurations (
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE PRIMARY KEY,
    default_currency VARCHAR(3) DEFAULT 'USD',
    default_language VARCHAR(10) DEFAULT 'en',
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    delivery_fee DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 3. User Management

### 3.1 Users
```sql
CREATE TABLE users (
    user_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
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
```

### 3.2 User Roles
```sql
CREATE TABLE user_roles (
    role_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 3.3 User Role Assignments
```sql
CREATE TABLE user_role_assignments (
    assignment_id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) REFERENCES users(user_id) ON DELETE CASCADE,
    role_id VARCHAR(50) REFERENCES user_roles(role_id) ON DELETE CASCADE,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 4. Restaurant Tables

### 4.1 Restaurants
```sql
CREATE TABLE tenant_restaurants (
    restaurant_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
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
```

### 4.2 Restaurant Categories
```sql
CREATE TABLE tenant_restaurant_categories (
    category_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    restaurant_id VARCHAR(50) REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_category_id INTEGER REFERENCES tenant_restaurant_categories(category_id) ON DELETE SET NULL,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL
);
```

## 5. Menu Items

### 5.1 Menu Items
```sql
CREATE TABLE tenant_menu_items (
    menu_item_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    restaurant_id VARCHAR(50) REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    category_id VARCHAR(50) REFERENCES tenant_restaurant_categories(category_id) ON DELETE SET NULL,
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
```

### 5.2 Menu Item Images
```sql
CREATE TABLE tenant_menu_item_images (
    image_id VARCHAR(50) PRIMARY KEY,
    menu_item_id VARCHAR(50) REFERENCES tenant_menu_items(menu_item_id) ON DELETE CASCADE,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    url VARCHAR(512) NOT NULL,
    alt_text VARCHAR(255),
    is_primary BOOLEAN DEFAULT FALSE,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 5.x Spec-aligned Tables (singular names)

The following tables mirror the `specs/food_menu_system/spec.md` design. Primary keys use `VARCHAR(50)`.

### Food Category
```sql
CREATE TABLE food_category (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Food Item
```sql
CREATE TABLE food_item (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    category_id VARCHAR(50) REFERENCES food_category(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Menu Item
```sql
CREATE TABLE menu_item (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    food_item_id VARCHAR(50) REFERENCES food_item(id) ON DELETE CASCADE,
    menu_id VARCHAR(50) NOT NULL,
    menu_type VARCHAR(20) NOT NULL,
    description VARCHAR(100),
    size VARCHAR(50),
    price DECIMAL(10,2) NOT NULL,
    sequence INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Weekly Menu
```sql
CREATE TABLE weekly_menu (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Catering Menu
```sql
CREATE TABLE catering_menu (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    event_date TIMESTAMP WITH TIME ZONE NOT NULL,
    event_location VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Food Catalog
```sql
CREATE TABLE food_catalog (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Order
```sql
CREATE TABLE "order" (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    user_id VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL,
    restaurant_id VARCHAR(50) REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
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
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Order Item
```sql
CREATE TABLE order_item (
    id VARCHAR(50) PRIMARY KEY,
    order_id VARCHAR(50) REFERENCES "order"(id) ON DELETE CASCADE,
    menu_item_id VARCHAR(50) REFERENCES menu_item(id) ON DELETE SET NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    total_price DECIMAL(10,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Notification
```sql
CREATE TABLE notification (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    user_id VARCHAR(50) REFERENCES users(user_id) ON DELETE CASCADE,
    order_id VARCHAR(50) REFERENCES "order"(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 6. Inventory Management

### 6.1 Inventory Categories
```sql
CREATE TABLE tenant_inventory_categories (
    category_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_category_id INTEGER REFERENCES tenant_inventory_categories(category_id) ON DELETE SET NULL,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL
);
```

### 6.2 Inventory Items
```sql
CREATE TABLE tenant_inventory_items (
    item_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    category_id VARCHAR(50) REFERENCES tenant_inventory_categories(category_id) ON DELETE SET NULL,
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
```

### 6.3 Inventory Stock
```sql
CREATE TABLE tenant_inventory_stock (
    stock_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    item_id VARCHAR(50) REFERENCES tenant_inventory_items(item_id) ON DELETE CASCADE,
    location_id INTEGER,
    quantity DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    reorder_threshold DECIMAL(15,2) DEFAULT 0.00,
    last_stock_take_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 7. Orders

### 7.1 Order Statuses
```sql
CREATE TABLE order_statuses (
    status_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_terminal BOOLEAN DEFAULT FALSE
);
```

### 7.2 Orders
```sql
CREATE TABLE tenant_orders (
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
    estimated_delivery_time TIMESTAMP WITH TIME ZONE,
    actual_delivery_time TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL,
    updated_by VARCHAR(50) REFERENCES users(user_id) ON DELETE SET NULL
);
```

### 7.3 Order Items
```sql
CREATE TABLE tenant_order_items (
    order_item_id VARCHAR(50) PRIMARY KEY,
    order_id VARCHAR(50) REFERENCES tenant_orders(order_id) ON DELETE CASCADE,
    menu_item_id VARCHAR(50) REFERENCES tenant_menu_items(menu_item_id) ON DELETE SET NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    tax_rate DECIMAL(5,2) DEFAULT 0.00,
    total_price DECIMAL(10,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 8. Payments

### 8.1 Payment Methods
```sql
CREATE TABLE payment_methods (
    method_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);
```

### 8.2 Payments
```sql
CREATE TABLE tenant_payments (
    payment_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    order_id VARCHAR(50) REFERENCES tenant_orders(order_id) ON DELETE SET NULL,
    payment_method_id VARCHAR(50) REFERENCES payment_methods(method_id) ON DELETE SET NULL,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    transaction_id VARCHAR(100),
    payment_gateway VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',
    payment_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 9. Reviews and Ratings

### 9.1 Reviews
```sql
CREATE TABLE tenant_reviews (
    review_id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    user_id VARCHAR(50) REFERENCES users(user_id) ON DELETE CASCADE,
    restaurant_id VARCHAR(50) REFERENCES tenant_restaurants(restaurant_id) ON DELETE CASCADE,
    menu_item_id VARCHAR(50) REFERENCES tenant_menu_items(menu_item_id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title VARCHAR(255),
    review_text TEXT,
    is_approved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 10. Indexes

### 10.1 Core Indexes
```sql
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
```

## 11. Triggers

### 11.1 Update Timestamps
```sql
CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables
CREATE TRIGGER update_tenants_timestamp
BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_users_timestamp
BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_restaurants_timestamp
BEFORE UPDATE ON tenant_restaurants FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_menu_items_timestamp
BEFORE UPDATE ON tenant_menu_items FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_orders_timestamp
BEFORE UPDATE ON tenant_orders FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_payments_timestamp
BEFORE UPDATE ON tenant_payments FOR EACH ROW EXECUTE FUNCTION update_timestamp();
```

## 12. Views

### 12.1 Restaurant Menu View
```sql
CREATE VIEW tenant_restaurant_menu AS
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
    r.tenant_id = CURRENT_TENANT_ID
    AND mi.is_active = TRUE
ORDER BY
    r.restaurant_id, c.display_order, mi.display_order;
```

### 12.2 User Orders View
```sql
CREATE VIEW user_orders AS
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
    o.tenant_id = CURRENT_TENANT_ID
    AND o.user_id = CURRENT_USER_ID
ORDER BY
    o.created_at DESC;
```

## 13. Functions

### 13.1 Tenant-Specific Functions
```sql
-- Function to get current tenant ID
CREATE OR REPLACE FUNCTION get_current_tenant_id() RETURNS VARCHAR(50) AS $$
DECLARE
    tenant_id VARCHAR(50);
BEGIN
    -- In a real application, this would be set via middleware or connection parameters
    -- For this schema, we'll assume it's passed via a session variable (session setting)
    tenant_id := current_setting('current_tenant_id', true);
    RETURN tenant_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get current user ID
CREATE OR REPLACE FUNCTION get_current_user_id() RETURNS VARCHAR(50) AS $$
DECLARE
    user_id VARCHAR(50);
BEGIN
    -- In a real application, this would be set via authentication middleware
    user_id := current_setting('current_user_id', true);
    RETURN user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create a view that includes tenant context
CREATE OR REPLACE VIEW current_tenant_context AS
SELECT
    current_setting('current_tenant_id', true) AS tenant_id,
    current_setting('current_user_id', true) AS user_id,
    (SELECT name FROM tenants WHERE tenant_id = current_setting('current_tenant_id', true)) AS tenant_name;
```

## 14. Security

### 14.1 Row-Level Security
```sql
-- Enable row-level security for tenant-specific tables
ALTER TABLE tenant_restaurants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_restaurant_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_menu_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_menu_item_images ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_inventory_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_inventory_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_inventory_stock ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_order_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_reviews ENABLE ROW LEVEL SECURITY;

-- Create policies for row-level security
CREATE POLICY tenant_restaurants_policy ON tenant_restaurants FOR ALL USING (tenant_id = get_current_tenant_id()::VARCHAR(50));
CREATE POLICY tenant_menu_items_policy ON tenant_menu_items FOR ALL USING (tenant_id = get_current_tenant_id()::VARCHAR(50));
CREATE POLICY tenant_orders_policy ON tenant_orders FOR ALL USING (tenant_id = get_current_tenant_id()::VARCHAR(50));
```

### 14.2 Default Security Roles
```sql
-- Create roles for different access levels
CREATE ROLE tenant_admin WITH NOLOGIN;
CREATE ROLE tenant_manager WITH NOLOGIN;
CREATE ROLE tenant_user WITH NOLOGIN;

-- Grant permissions to roles
GRANT CONNECT ON DATABASE your_database_name TO tenant_admin, tenant_manager, tenant_user;

-- Example permissions for tenant_admin
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO tenant_admin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_admin;

-- Example permissions for tenant_manager
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO tenant_manager;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_manager;

-- Example permissions for tenant_user
GRANT SELECT ON ALL TABLES IN SCHEMA public TO tenant_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_user;
```

## 15. Data Migration Strategy

### 15.1 Initial Setup Script
```sql
-- Create the database
CREATE DATABASE homecooked;

-- Connect to the database and run the schema creation scripts
\\conn homecooked
\\i database_schema.sql
```

### 15.2 Sample Data Insertion
```sql
-- Insert sample tenants (use string IDs)
INSERT INTO tenants (tenant_id, name, domain, status) VALUES
('tenant_1', 'Restaurant Chain 1', 'restaurant1.example.com', 'active'),
('tenant_2', 'Restaurant Chain 2', 'restaurant2.example.com', 'active'),
('tenant_3', 'Food Delivery Service', 'foodservice.example.com', 'active');

-- Insert sample configurations (tenant IDs as strings)
INSERT INTO tenant_configurations (tenant_id, default_currency, default_language, tax_rate, delivery_fee) VALUES
('tenant_1', 'USD', 'en', 0.08, 2.99),
('tenant_2', 'EUR', 'fr', 0.20, 3.99),
('tenant_3', 'USD', 'es', 0.07, 1.99);
```

## 16. Maintenance Procedures

### 16.1 Regular Maintenance
```sql
-- Create a function to analyze and rebuild indexes
CREATE OR REPLACE FUNCTION analyze_and_rebuild_indexes() RETURNS VOID AS $$
DECLARE
    query TEXT;
BEGIN
    -- Analyze all tables
    query := 'ANALYZE;
             REINDEX DATABASE food_menu_system;
             VACUUM FULL food_menu_system;
             VACUUM ANALYZE food_menu_system;
             REINDEX TABLE tenants;
             REINDEX TABLE tenant_restaurants;
             REINDEX TABLE tenant_menu_items;
             REINDEX TABLE tenant_orders;
             REINDEX TABLE tenant_payments;
             REINDEX TABLE tenant_reviews;
             REINDEX TABLE current_tenant_context;
             ';
    EXECUTE query;
END;
$$ LANGUAGE plpgsql;
```

## 17. Backup Strategy

### 17.1 Backup Script
```sql
-- Create a function to perform backups
CREATE OR REPLACE FUNCTION perform_backup() RETURNS VOID AS $$
DECLARE
    backup_name TEXT := 'food_menu_system_' || TO_CHAR(CURRENT_DATE, 'YYYY-MM-DD_HH24MISS') || '.sql';
    backup_path TEXT := '/backups/' || backup_name;
BEGIN
    -- Create backup
    EXECUTE format('pg_dump -U postgres -Fc -d food_menu_system -f %I', backup_path);
    
    -- Optional: Compress backup
    -- EXECUTE format('gzip %I', backup_path);
    
    -- Log backup information
    INSERT INTO backup_logs (backup_name, backup_path, backup_date, backup_size, status) 
    VALUES (backup_name, backup_path, CURRENT_TIMESTAMP, (SELECT pg_size_pretty(pg_total_relation_size('food_menu_system'))), 'completed');
END;
$$ LANGUAGE plpgsql;
```

## 18. Documentation Notes

### 18.1 Schema Versioning
- **Version**: 1.0
- **Last Updated**: 2023-11-15
- **Next Major Version**: 2.0 (when significant architectural changes are made)

### 18.2 Migration Path
1. **Initial Setup**: Create database and apply schema
2. **Data Migration**: Populate initial data using provided scripts
3. **Application Integration**: Configure application to use tenant context
4. **Testing**: Thorough testing of all tenant-specific functionality
5. **Go Live**: Gradual rollout with monitoring

### 18.3 Support Information
- **Database**: PostgreSQL 14+ recommended
- **Extensions**: pg_trgm for text search, pg_cron for scheduling
- **Connection Pooling**: PgBouncer recommended for production
- **Monitoring**: Enable PostgreSQL logging for query performance

## 19. Appendix

### 19.1 Sample Queries
```sql
-- Get all active restaurants for a tenant (tenant IDs are strings)
SELECT * FROM tenant_restaurants WHERE tenant_id = 'tenant_1' AND is_active = TRUE;

-- Get menu items for a specific restaurant (string IDs)
SELECT * FROM tenant_menu_items WHERE restaurant_id = 'restaurant_123' AND is_active = TRUE;

-- Get order history for a user (string user IDs)
SELECT * FROM tenant_orders WHERE user_id = 'user_456' ORDER BY created_at DESC;

-- Get restaurant ratings (tenant ID as string)
SELECT r.restaurant_id, r.name, AVG(rr.rating) AS avg_rating
FROM tenant_restaurants r
JOIN tenant_reviews rr ON r.restaurant_id = rr.restaurant_id
WHERE r.tenant_id = 'tenant_1'
GROUP BY r.restaurant_id, r.name;
```

### 19.2 Performance Considerations
1. **Indexing**: Ensure proper indexing for frequently queried columns
2. **Partitioning**: Consider partitioning large tables like orders by date ranges
3. **Caching**: Implement application-level caching for frequently accessed data
4. **Connection Management**: Use connection pooling to manage database connections
5. **Query Optimization**: Review slow queries using EXPLAIN ANALYZE

### 19.3 Security Best Practices
1. **Principle of Least Privilege**: Grant only necessary permissions
2. **Regular Audits**: Perform regular security audits and reviews
3. **Data Encryption**: Encrypt sensitive data at rest and in transit
4. **Backup Security**: Secure backup storage and access
5. **Access Controls**: Implement strong authentication and authorization

### 19.4 Future Enhancements
1. **Sharding**: Consider database sharding for very large-scale deployments
2. **NoSQL Integration**: Add NoSQL support for unstructured data
3. **Machine Learning**: Integrate recommendation algorithms
4. **Real-time Analytics**: Add real-time data processing capabilities
5. **Multi-cloud Support**: Extend to support multiple cloud providers