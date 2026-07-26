-- Enable Row-Level Security on all tenant-specific tables

-- 1. Tenants Table (Base table - no RLS needed)
-- Note: Superusers/admins can access all tenants without RLS restrictions

-- 2. Tenant Configurations
ALTER TABLE tenant_configurations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_configurations FOR SELECT USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_configurations FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_configurations FOR UPDATE USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 3. Users Table
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON users FOR SELECT USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON users FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON users FOR UPDATE USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 4. User Roles
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON user_roles FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 5. User Role Assignments
ALTER TABLE user_role_assignments ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON user_role_assignments FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 6. Tenant Restaurants
ALTER TABLE tenant_restaurants ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_restaurants FOR SELECT USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_restaurants FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_restaurants FOR UPDATE USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 7. Tenant Restaurant Categories
ALTER TABLE tenant_restaurant_categories ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_restaurant_categories FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 8. Tenant Menu Items
ALTER TABLE tenant_menu_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_menu_items FOR SELECT USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_menu_items FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_menu_items FOR UPDATE USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 9. Tenant Menu Item Images
ALTER TABLE tenant_menu_item_images ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_menu_item_images FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 10. Tenant Inventory Categories
ALTER TABLE tenant_inventory_categories ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_inventory_categories FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 11. Tenant Inventory Items
ALTER TABLE tenant_inventory_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_inventory_items FOR SELECT USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_inventory_items FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_inventory_items FOR UPDATE USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 12. Tenant Inventory Stock
ALTER TABLE tenant_inventory_stock ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_inventory_stock FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 13. Orders
ALTER TABLE tenant_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_orders FOR SELECT USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_orders FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_orders FOR UPDATE USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 14. Order Items
ALTER TABLE tenant_order_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_order_items FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 15. Payments
ALTER TABLE tenant_payments ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_payments FOR SELECT USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_payments FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_payments FOR UPDATE USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- 16. Reviews
ALTER TABLE tenant_reviews ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_reviews FOR SELECT USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_reviews FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::TEXT);
CREATE POLICY tenant_isolation ON tenant_reviews FOR UPDATE USING (tenant_id = current_setting('app.current_tenant_id')::TEXT);

-- Create Roles for Different Access Levels
CREATE ROLE tenant_admin NOLOGIN; -- Full access to own tenant data
CREATE ROLE tenant_manager NOLOGIN; -- Limited access with management capabilities
CREATE ROLE tenant_user NOLOGIN; -- Basic read/write access
CREATE ROLE superadmin NOLOGIN; -- Superuser with full database access

-- Grant Connect Privilege for All Roles
GRANT CONNECT ON DATABASE food_menu_system TO tenant_admin, tenant_manager, tenant_user, superadmin;

-- Grant Usage on Schema
GRANT USAGE ON SCHEMA public TO tenant_admin, tenant_manager, tenant_user, superadmin;

-- Grant Sequence Access
GRANT SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO tenant_admin, tenant_manager, tenant_user, superadmin;

-- Grant Permissions for Tenant Admin (Full access to own tenant data)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO tenant_admin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_admin;

-- Grant Permissions for Tenant Manager (Limited write access)
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO tenant_manager;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_manager;

-- Grant Permissions for Tenant User (Read-only with some write access)
GRANT SELECT ON ALL TABLES IN SCHEMA public TO tenant_user;
GRANT INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO tenant_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_user;

-- Grant Superadmin Full Access (Bypass RLS for administrative tasks)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO superadmin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO superadmin;

-- Create Function to Set Tenant Context
CREATE OR REPLACE FUNCTION app.set_current_tenant_id(p_tenant_id TEXT) RETURNS void AS $$BEGIN
    SET LOCAL 'app.current_tenant_id' = p_tenant_id::TEXT;
END;$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create Function to Get Current Tenant ID
CREATE OR REPLACE FUNCTION app.get_current_tenant_id() RETURNS TEXT STABLE AS $$DECLARE
    result TEXT;
BEGIN
    SELECT value INTO result FROM pg_settings WHERE name = 'app.current_tenant_id';
    IF result IS NULL THEN
        -- Return 0 if no tenant is set (shouldn't happen with RLS policies)
        RETURN 0;
    END IF;
    RETURN result::TEXT;
END;$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create Function to Check if Current User Has Access
CREATE OR REPLACE FUNCTION app.has_tenant_access(p_user_id TEXT) RETURNS BOOLEAN STABLE AS $$DECLARE
    has_access BOOLEAN;
BEGIN
    SELECT COUNT(*) > 0 INTO has_access FROM users WHERE user_id = p_user_id AND tenant_id = current_setting('app.current_tenant_id')::TEXT;
    RETURN has_access;
END;$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create Function to Set User Context with Tenant
CREATE OR REPLACE FUNCTION app.set_user_context(p_user_id TEXT, p_tenant_id TEXT) RETURNS void AS $$BEGIN
    -- This function sets both user and tenant context for the session
    SET 'app.current_user_id' = p_user_id::TEXT;
    SET 'app.current_tenant_id' = p_tenant_id::TEXT;
END;$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Insert Default Tenants (these will be locked by RLS policies)
INSERT INTO tenants (name, domain, status, is_active) 
VALUES ('Demo Restaurant', 'demo.local', 'active', TRUE)
ON CONFLICT (domain) DO NOTHING;

-- Verify Configuration
SELECT table_name, row_security FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;
SELECT rolname, role_admin, role_member, role_submember FROM pg_roles WHERE rolname IN ('tenant_admin', 'tenant_manager', 'tenant_user', 'superadmin');