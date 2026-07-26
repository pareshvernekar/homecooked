-- Enable Row-Level Security on all tenant-specific tables
-- Updated Version: 1.1 - Alphanumeric Tenant IDs (VARCHAR)

-- ============================================================================
-- IMPORTANT: Ensure tables are created before running this script!
-- This should be run AFTER database_schema.sql to add RLS to existing tables
-- ============================================================================

-- Tenants table (base table, no RLS needed as it's the source of truth)
-- Tenants are managed by superusers or admin roles that bypass RLS

-- User Roles
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON user_roles;
CREATE POLICY tenant_isolation_policy ON user_roles
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- User Role Assignments
ALTER TABLE user_role_assignments ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON user_role_assignments;
CREATE POLICY tenant_isolation_policy ON user_role_assignments
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Restaurants
ALTER TABLE tenant_restaurants ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_restaurants;
CREATE POLICY tenant_isolation_policy ON tenant_restaurants
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Restaurant Categories
ALTER TABLE tenant_restaurant_categories ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_restaurant_categories;
CREATE POLICY tenant_isolation_policy ON tenant_restaurant_categories
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Menu Items
ALTER TABLE tenant_menu_items ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_menu_items;
CREATE POLICY tenant_isolation_policy ON tenant_menu_items
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Menu Item Images
ALTER TABLE tenant_menu_item_images ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_menu_item_images;
CREATE POLICY tenant_isolation_policy ON tenant_menu_item_images
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Inventory Categories
ALTER TABLE tenant_inventory_categories ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_inventory_categories;
CREATE POLICY tenant_isolation_policy ON tenant_inventory_categories
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Inventory Items
ALTER TABLE tenant_inventory_items ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_inventory_items;
CREATE POLICY tenant_isolation_policy ON tenant_inventory_items
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Inventory Stock
ALTER TABLE tenant_inventory_stock ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_inventory_stock;
CREATE POLICY tenant_isolation_policy ON tenant_inventory_stock
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Orders
ALTER TABLE tenant_orders ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_orders;
CREATE POLICY tenant_isolation_policy ON tenant_orders
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Order Items
ALTER TABLE tenant_order_items ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_order_items;
CREATE POLICY tenant_isolation_policy ON tenant_order_items
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Payments
ALTER TABLE tenant_payments ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_payments;
CREATE POLICY tenant_isolation_policy ON tenant_payments
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Reviews
ALTER TABLE tenant_reviews ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_reviews;
CREATE POLICY tenant_isolation_policy ON tenant_reviews
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- ============================================================================
-- CONFIGURATION: Session-level functions to set the current tenant ID
-- This should be called in your application middleware before any queries
-- ============================================================================

-- Function to set the current tenant ID in session
CREATE OR REPLACE FUNCTION app.set_current_tenant_id(p tenant_id) RETURNS VOID AS $$BEGIN
    SET LOCAL 'app.current_tenant_id' = p::TEXT;
END;$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Helper function to get current tenant ID (for SELECT statements)
CREATE OR REPLACE FUNCTION app.get_current_tenant_id() RETURNS TEXT AS $$DECLARE
    tenant_id TEXT;
BEGIN
    SELECT value INTO tenant_id FROM pg_settings WHERE name = 'app.current_tenant_id';
    RETURN COALESCE(tenant_id, '1'); -- Default to tenant ID '1' if not set
END;$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================================================
-- ROLE-BASED ACCESS CONTROL
-- Configure Role-Based Access Control
-- Create a role that can manage tenants (superuser equivalent)
-- ============================================================================

CREATE ROLE tenant_admin NOLOGIN;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO tenant_admin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_admin;

-- Create a role for individual tenant users
CREATE ROLE tenant_user NOLOGIN;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_restaurants TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_restaurant_categories TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_menu_items TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_menu_item_images TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_inventory_categories TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_inventory_items TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_inventory_stock TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_orders TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_order_items TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_payments TO tenant_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_reviews TO tenant_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_user;