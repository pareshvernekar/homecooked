-- ============================================
-- Sample Data Insertion Script: Users and Restaurants
-- File: 2_setup_users_and_restaurants.sql
-- Description: This script creates sample users and restaurants for each tenant
-- ============================================

-- Enable row-level security temporarily for tenant-specific tables
SET LOCAL row_security = off;

-- ============================================
-- Create sample users for each tenant
-- ============================================
-- Burger Haven
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    ('user_t1_admin', 'tenant_1', 'admin@burgerhaven.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'John', 'Doe', '+15551234567', TRUE, TRUE),
    ('user_t1_manager', 'tenant_1', 'manager@burgerhaven.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Sarah', 'Johnson', '+15552345678', TRUE, FALSE),
    ('user_t1_customer1', 'tenant_1', 'customer1@burgerhaven.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Michael', 'Williams', '+15553456789', TRUE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Sushi Palace
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    ('user_t2_admin', 'tenant_2', 'admin@sushipalace.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Kenji', 'Tanaka', '+81312345678', TRUE, TRUE),
    ('user_t2_manager', 'tenant_2', 'manager@sushipalace.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Yuki', 'Kimura', '+81398765432', TRUE, FALSE),
    ('user_t2_customer1', 'tenant_2', 'customer1@sushipalace.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Akira', 'Sato', '+8135551234', TRUE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Veggie Delight
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    ('user_t3_admin', 'tenant_3', 'admin@veggiedelight.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Emma', 'Wilson', '+15554567890', TRUE, TRUE),
    ('user_t3_manager', 'tenant_3', 'manager@veggiedelight.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Olivia', 'Martinez', '+15556789012', TRUE, FALSE),
    ('user_t3_customer1', 'tenant_3', 'customer1@veggiedelight.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Sophia', 'Lee', '+15557890123', TRUE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Italian Eats
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    ('user_t4_admin', 'tenant_4', 'admin@italianeats.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Marco', 'Rossi', '+39061234567', TRUE, TRUE),
    ('user_t4_manager', 'tenant_4', 'manager@italianeats.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Lucia', 'Bianchi', '+39067890123', TRUE, FALSE),
    ('user_t4_customer1', 'tenant_4', 'customer1@italianeats.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Giuseppe', 'Ferrari', '+39064567890', TRUE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Spicy Wings
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    ('user_t5_admin', 'tenant_5', 'admin@spicywings.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'James', 'Smith', '+15559876543', TRUE, TRUE),
    ('user_t5_manager', 'tenant_5', 'manager@spicywings.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Jessica', 'Taylor', '+15558765432', TRUE, FALSE),
    ('user_t5_customer1', 'tenant_5', 'customer1@spicywings.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'David', 'Anderson', '+15557654321', TRUE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Healthy Bites
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    ('user_t6_admin', 'tenant_6', 'admin@healthybites.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Emily', 'Davis', '+15556543210', TRUE, TRUE),
    ('user_t6_manager', 'tenant_6', 'manager@healthybites.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Sarah', 'Miller', '+15555432109', TRUE, FALSE),
    ('user_t6_customer1', 'tenant_6', 'customer1@healthybites.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Daniel', 'Wilson', '+15554321098', TRUE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Dessert Den
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    ('user_t7_admin', 'tenant_7', 'admin@dessertden.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Sophia', 'Garcia', '+15553210987', TRUE, TRUE),
    ('user_t7_manager', 'tenant_7', 'manager@dessertden.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Olivia', 'Rodriguez', '+15552109876', TRUE, FALSE),
    ('user_t7_customer1', 'tenant_7', 'customer1@dessertden.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'James', 'Clark', '+15551098765', TRUE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- Breakfast Club
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    ('user_t8_admin', 'tenant_8', 'admin@breakfastclub.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'William', 'Brown', '+15552345670', TRUE, TRUE),
    ('user_t8_manager', 'tenant_8', 'manager@breakfastclub.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Amanda', 'Lee', '+15553456781', TRUE, FALSE),
    ('user_t8_customer1', 'tenant_8', 'customer1@breakfastclub.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Michael', 'Taylor', '+15554567892', TRUE, FALSE)
ON CONFLICT (user_id) DO NOTHING;

-- ============================================
-- Assign roles to users
-- ============================================
-- Burger Haven
INSERT INTO user_roles (role_id, tenant_id, name, description) VALUES
    ('role_burger_haven_admin','tenant_1','burger_haven_admin','Full access for Burger Haven'),
    ('role_burger_haven_restaurant_manager','tenant_1','burger_haven_restaurant_manager','Restaurant manager for Burger Haven'),
    ('role_burger_haven_menu_editor','tenant_1','burger_haven_menu_editor','Menu editor for Burger Haven')
ON CONFLICT (role_id) DO NOTHING;

INSERT INTO user_role_assignments (assignment_id, user_id, role_id, tenant_id) VALUES
    ('ura_t1_admin', 'user_t1_admin', 'role_burger_haven_admin', 'tenant_1'),
    ('ura_t1_manager', 'user_t1_manager', 'role_burger_haven_restaurant_manager', 'tenant_1'),
    ('ura_t1_editor', 'user_t1_customer1', 'role_burger_haven_menu_editor', 'tenant_1')
ON CONFLICT (assignment_id) DO NOTHING;

-- Sushi Palace
INSERT INTO user_roles (role_id, tenant_id, name, description) VALUES
    ('role_sushi_palace_admin','tenant_2','sushi_palace_admin','Full access for Sushi Palace'),
    ('role_sushi_palace_restaurant_manager','tenant_2','sushi_palace_restaurant_manager','Restaurant manager for Sushi Palace'),
    ('role_sushi_palace_menu_editor','tenant_2','sushi_palace_menu_editor','Menu editor for Sushi Palace')
ON CONFLICT (role_id) DO NOTHING;

INSERT INTO user_role_assignments (assignment_id, user_id, role_id, tenant_id) VALUES
    ('ura_t2_admin', 'user_t2_admin', 'role_sushi_palace_admin', 'tenant_2'),
    ('ura_t2_manager', 'user_t2_manager', 'role_sushi_palace_restaurant_manager', 'tenant_2'),
    ('ura_t2_editor', 'user_t2_customer1', 'role_sushi_palace_menu_editor', 'tenant_2')
ON CONFLICT (assignment_id) DO NOTHING;

-- ============================================
-- Create sample restaurants for each tenant
-- ============================================
-- Tenant restaurants (use `tenant_restaurants` and explicit PKs)
INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, description, address_line1, address_line2, city, state, postal_code, country, phone, email, website_url, logo_url, is_active, created_by) VALUES
    ('restaurant_t1_1', 'tenant_1', 'Burger Haven Downtown', NULL, '123 Main St', NULL, 'New York', 'NY', '10001', 'USA', '+15551234567', 'info@burgerhaven.example.com', 'https://burgerhaven.example.com', 'https://burgerhaven.example.com/logo.png', TRUE, 'user_t1_admin'),
    ('restaurant_t1_2', 'tenant_1', 'Burger Haven Midtown', NULL, '456 Park Ave', NULL, 'New York', 'NY', '10021', 'USA', '+15557890123', 'info@burgerhaven-midtown.example.com', 'https://burgerhaven-midtown.example.com', 'https://burgerhaven.example.com/logo-midtown.png', TRUE, 'user_t1_admin')
ON CONFLICT (restaurant_id) DO NOTHING;

INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, description, address_line1, address_line2, city, state, postal_code, country, phone, email, website_url, logo_url, is_active, created_by) VALUES
    ('restaurant_t2_1', 'tenant_2', 'Sushi Palace Tokyo', NULL, '789 Ginza St', NULL, 'Tokyo', NULL, '10400432', 'Japan', '+81312345678', 'info@sushipalace-tokyo.example.com', 'https://sushipalace-tokyo.example.com', 'https://sushipalace.example.com/logo-tokyo.png', TRUE, 'user_t2_admin'),
    ('restaurant_t2_2', 'tenant_2', 'Sushi Palace Osaka', NULL, '321 Dotonbori Ave', NULL, 'Osaka', NULL, '5420001', 'Japan', '+81612345678', 'info@sushipalace-osaka.example.com', 'https://sushipalace-osaka.example.com', 'https://sushipalace.example.com/logo-osaka.png', TRUE, 'user_t2_admin')
ON CONFLICT (restaurant_id) DO NOTHING;

INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, description, address_line1, address_line2, city, state, postal_code, country, phone, email, website_url, logo_url, is_active, created_by) VALUES
    ('restaurant_t3_1', 'tenant_3', 'Veggie Delight Downtown', NULL, '321 Oak St', NULL, 'Chicago', 'IL', '60601', 'USA', '+15552345678', 'info@veggiedelight.example.com', 'https://veggiedelight.example.com', 'https://veggiedelight.example.com/logo.png', TRUE, 'user_t3_admin'),
    ('restaurant_t3_2', 'tenant_3', 'Veggie Delight Lakeshore', NULL, '654 Lake Shore Dr', NULL, 'Chicago', 'IL', '60611', 'USA', '+15553456789', 'info@veggiedelight-lakeshore.example.com', 'https://veggiedelight-lakeshore.example.com', 'https://veggiedelight.example.com/logo-lakeshore.png', TRUE, 'user_t3_admin')
ON CONFLICT (restaurant_id) DO NOTHING;

INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, description, address_line1, address_line2, city, state, postal_code, country, phone, email, website_url, logo_url, is_active, created_by) VALUES
    ('restaurant_t4_1', 'tenant_4', 'Italian Eats Rome', NULL, '159 Via dei Condotti', NULL, 'Rome', NULL, '00186', 'Italy', '+39064888888', 'info@italianeats-rome.example.com', 'https://italianeats-rome.example.com', 'https://italianeats.example.com/logo-rome.png', TRUE, 'user_t4_admin'),
    ('restaurant_t4_2', 'tenant_4', 'Italian Eats Milan', NULL, '246 Via Montenapoleone', NULL, 'Milan', NULL, '20121', 'Italy', '+39027899999', 'info@italianeats-milan.example.com', 'https://italianeats-milan.example.com', 'https://italianeats.example.com/logo-milan.png', TRUE, 'user_t4_admin')
ON CONFLICT (restaurant_id) DO NOTHING;

INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, description, address_line1, address_line2, city, state, postal_code, country, phone, email, website_url, logo_url, is_active, created_by) VALUES
    ('restaurant_t5_1', 'tenant_5', 'Spicy Wings Dallas', NULL, '789 Main St', NULL, 'Dallas', 'TX', '75201', 'USA', '+15554567890', 'info@spicywings-dallas.example.com', 'https://spicywings-dallas.example.com', 'https://spicywings.example.com/logo-dallas.png', TRUE, 'user_t5_admin'),
    ('restaurant_t5_2', 'tenant_5', 'Spicy Wings Austin', NULL, '321 Congress Ave', NULL, 'Austin', 'TX', '78701', 'USA', '+15555678901', 'info@spicywings-austin.example.com', 'https://spicywings-austin.example.com', 'https://spicywings.example.com/logo-austin.png', TRUE, 'user_t5_admin')
ON CONFLICT (restaurant_id) DO NOTHING;

INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, description, address_line1, address_line2, city, state, postal_code, country, phone, email, website_url, logo_url, is_active, created_by) VALUES
    ('restaurant_t6_1', 'tenant_6', 'Healthy Bites San Francisco', NULL, '456 Market St', NULL, 'San Francisco', 'CA', '94105', 'USA', '+15556789012', 'info@healthybites-sf.example.com', 'https://healthybites-sf.example.com', 'https://healthybites.example.com/logo-sf.png', TRUE, 'user_t6_admin'),
    ('restaurant_t6_2', 'tenant_6', 'Healthy Bites Berkeley', NULL, '789 University Ave', NULL, 'Berkeley', 'CA', '94701', 'USA', '+15557890123', 'info@healthybites-berkeley.example.com', 'https://healthybites-berkeley.example.com', 'https://healthybites.example.com/logo-berkeley.png', TRUE, 'user_t6_admin')
ON CONFLICT (restaurant_id) DO NOTHING;

INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, description, address_line1, address_line2, city, state, postal_code, country, phone, email, website_url, logo_url, is_active, created_by) VALUES
    ('restaurant_t7_1', 'tenant_7', 'Dessert Den New York', NULL, '987 Fifth Ave', NULL, 'New York', 'NY', '10017', 'USA', '+15558901234', 'info@dessertden-ny.example.com', 'https://dessertden-ny.example.com', 'https://dessertden.example.com/logo-ny.png', TRUE, 'user_t7_admin'),
    ('restaurant_t7_2', 'tenant_7', 'Dessert Den Chicago', NULL, '135 Michigan Ave', NULL, 'Chicago', 'IL', '60601', 'USA', '+15559012345', 'info@dessertden-chicago.example.com', 'https://dessertden-chicago.example.com', 'https://dessertden.example.com/logo-chicago.png', TRUE, 'user_t7_admin')
ON CONFLICT (restaurant_id) DO NOTHING;

INSERT INTO tenant_restaurants (restaurant_id, tenant_id, name, description, address_line1, address_line2, city, state, postal_code, country, phone, email, website_url, logo_url, is_active, created_by) VALUES
    ('restaurant_t8_1', 'tenant_8', 'Breakfast Club Seattle', NULL, '654 Pike St', NULL, 'Seattle', 'WA', '98101', 'USA', '+15550123456', 'info@breakfastclub-seattle.example.com', 'https://breakfastclub-seattle.example.com', 'https://breakfastclub.example.com/logo-seattle.png', TRUE, 'user_t8_admin'),
    ('restaurant_t8_2', 'tenant_8', 'Breakfast Club Portland', NULL, '321 NW 23rd Ave', NULL, 'Portland', 'OR', '97210', 'USA', '+15551234567', 'info@breakfastclub-portland.example.com', 'https://breakfastclub-portland.example.com', 'https://breakfastclub.example.com/logo-portland.png', TRUE, 'user_t8_admin')
ON CONFLICT (restaurant_id) DO NOTHING;

-- (restaurant_hours inserts removed — no matching table in current schema)

-- ============================================
-- Re-enable row-level security
-- ============================================
SET LOCAL row_security = on;