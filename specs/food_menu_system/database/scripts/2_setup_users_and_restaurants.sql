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
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    (1, 'admin@burgerhaven.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'John', 'Doe', '+15551234567', TRUE, TRUE),
    (1, 'manager@burgerhaven.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Sarah', 'Johnson', '+15552345678', TRUE, FALSE),
    (1, 'customer1@burgerhaven.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Michael', 'Williams', '+15553456789', TRUE, FALSE);

-- Sushi Palace
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    (2, 'admin@sushipalace.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Kenji', 'Tanaka', '+81312345678', TRUE, TRUE),
    (2, 'manager@sushipalace.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Yuki', 'Kimura', '+81398765432', TRUE, FALSE),
    (2, 'customer1@sushipalace.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Akira', 'Sato', '+8135551234', TRUE, FALSE);

-- Veggie Delight
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    (3, 'admin@veggiedelight.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Emma', 'Wilson', '+15554567890', TRUE, TRUE),
    (3, 'manager@veggiedelight.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Olivia', 'Martinez', '+15556789012', TRUE, FALSE),
    (3, 'customer1@veggiedelight.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Sophia', 'Lee', '+15557890123', TRUE, FALSE);

-- Italian Eats
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    (4, 'admin@italianeats.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Marco', 'Rossi', '+39061234567', TRUE, TRUE),
    (4, 'manager@italianeats.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Lucia', 'Bianchi', '+39067890123', TRUE, FALSE),
    (4, 'customer1@italianeats.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Giuseppe', 'Ferrari', '+39064567890', TRUE, FALSE);

-- Spicy Wings
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    (5, 'admin@spicywings.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'James', 'Smith', '+15559876543', TRUE, TRUE),
    (5, 'manager@spicywings.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Jessica', 'Taylor', '+15558765432', TRUE, FALSE),
    (5, 'customer1@spicywings.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'David', 'Anderson', '+15557654321', TRUE, FALSE);

-- Healthy Bites
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    (6, 'admin@healthybites.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Emily', 'Davis', '+15556543210', TRUE, TRUE),
    (6, 'manager@healthybites.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Sarah', 'Miller', '+15555432109', TRUE, FALSE),
    (6, 'customer1@healthybites.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Daniel', 'Wilson', '+15554321098', TRUE, FALSE);

-- Dessert Den
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    (7, 'admin@dessertden.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Sophia', 'Garcia', '+15553210987', TRUE, TRUE),
    (7, 'manager@dessertden.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Olivia', 'Rodriguez', '+15552109876', TRUE, FALSE),
    (7, 'customer1@dessertden.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'James', 'Clark', '+15551098765', TRUE, FALSE);

-- Breakfast Club
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, phone, is_active, is_admin) VALUES
    (8, 'admin@breakfastclub.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'William', 'Brown', '+15552345670', TRUE, TRUE),
    (8, 'manager@breakfastclub.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Amanda', 'Lee', '+15553456781', TRUE, FALSE),
    (8, 'customer1@breakfastclub.example.com', '$2a$10$N9qo8uLOzTmX7ZMMZoP6eOePyWxjvB3Z3XxXxXxXxXxXxXxXxXxX', 'Michael', 'Taylor', '+15554567892', TRUE, FALSE);

-- ============================================
-- Assign roles to users
-- ============================================
-- Burger Haven
INSERT INTO user_roles (tenant_id, name, description) VALUES
    (1, 'burger_haven_restaurant_manager', 'Restaurant manager for Burger Haven'),
    (1, 'burger_haven_menu_editor', 'Menu editor for Burger Haven');

INSERT INTO user_role_assignments (user_id, role_id) VALUES
    (1, (SELECT role_id FROM user_roles WHERE tenant_id = 1 AND name = 'burger_haven_admin')),
    (2, (SELECT role_id FROM user_roles WHERE tenant_id = 1 AND name = 'burger_haven_restaurant_manager')),
    (3, (SELECT role_id FROM user_roles WHERE tenant_id = 1 AND name = 'burger_haven_menu_editor'));

-- Sushi Palace
INSERT INTO user_roles (tenant_id, name, description) VALUES
    (2, 'sushi_palace_restaurant_manager', 'Restaurant manager for Sushi Palace'),
    (2, 'sushi_palace_menu_editor', 'Menu editor for Sushi Palace');

INSERT INTO user_role_assignments (user_id, role_id) VALUES
    (4, (SELECT role_id FROM user_roles WHERE tenant_id = 2 AND name = 'sushi_palace_admin')),
    (5, (SELECT role_id FROM user_roles WHERE tenant_id = 2 AND name = 'sushi_palace_restaurant_manager')),
    (6, (SELECT role_id FROM user_roles WHERE tenant_id = 2 AND name = 'sushi_palace_menu_editor'));

-- ============================================
-- Create sample restaurants for each tenant
-- ============================================
-- Burger Haven
INSERT INTO restaurants (tenant_id, name, address, city, state, zip_code, country, phone, email, website, logo_url, is_active) VALUES
    (1, 'Burger Haven Downtown', '123 Main St', 'New York', 'NY', '10001', 'USA', '+15551234567', 'info@burgerhaven.example.com', 'https://burgerhaven.example.com', 'https://burgerhaven.example.com/logo.png', TRUE),
    (1, 'Burger Haven Midtown', '456 Park Ave', 'New York', 'NY', '10021', 'USA', '+15557890123', 'info@burgerhaven-midtown.example.com', 'https://burgerhaven-midtown.example.com', 'https://burgerhaven.example.com/logo-midtown.png', TRUE);

-- Sushi Palace
INSERT INTO restaurants (tenant_id, name, address, city, state, zip_code, country, phone, email, website, logo_url, is_active) VALUES
    (2, 'Sushi Palace Tokyo', '789 Ginza St', 'Tokyo', '', '10400432', 'Japan', '+81312345678', 'info@sushipalace-tokyo.example.com', 'https://sushipalace-tokyo.example.com', 'https://sushipalace.example.com/logo-tokyo.png', TRUE),
    (2, 'Sushi Palace Osaka', '321 Dotonbori Ave', 'Osaka', '', '5420001', 'Japan', '+81612345678', 'info@sushipalace-osaka.example.com', 'https://sushipalace-osaka.example.com', 'https://sushipalace.example.com/logo-osaka.png', TRUE);

-- Veggie Delight
INSERT INTO restaurants (tenant_id, name, address, city, state, zip_code, country, phone, email, website, logo_url, is_active) VALUES
    (3, 'Veggie Delight Downtown', '321 Oak St', 'Chicago', 'IL', '60601', 'USA', '+15552345678', 'info@veggiedelight.example.com', 'https://veggiedelight.example.com', 'https://veggiedelight.example.com/logo.png', TRUE),
    (3, 'Veggie Delight Lakeshore', '654 Lake Shore Dr', 'Chicago', 'IL', '60611', 'USA', '+15553456789', 'info@veggiedelight-lakeshore.example.com', 'https://veggiedelight-lakeshore.example.com', 'https://veggiedelight.example.com/logo-lakeshore.png', TRUE);

-- Italian Eats
INSERT INTO restaurants (tenant_id, name, address, city, state, zip_code, country, phone, email, website, logo_url, is_active) VALUES
    (4, 'Italian Eats Rome', '159 Via dei Condotti', 'Rome', '', '00186', 'Italy', '+39064888888', 'info@italianeats-rome.example.com', 'https://italianeats-rome.example.com', 'https://italianeats.example.com/logo-rome.png', TRUE),
    (4, 'Italian Eats Milan', '246 Via Montenapoleone', 'Milan', '', '20121', 'Italy', '+39027899999', 'info@italianeats-milan.example.com', 'https://italianeats-milan.example.com', 'https://italianeats.example.com/logo-milan.png', TRUE);

-- Spicy Wings
INSERT INTO restaurants (tenant_id, name, address, city, state, zip_code, country, phone, email, website, logo_url, is_active) VALUES
    (5, 'Spicy Wings Dallas', '789 Main St', 'Dallas', 'TX', '75201', 'USA', '+15554567890', 'info@spicywings-dallas.example.com', 'https://spicywings-dallas.example.com', 'https://spicywings.example.com/logo-dallas.png', TRUE),
    (5, 'Spicy Wings Austin', '321 Congress Ave', 'Austin', 'TX', '78701', 'USA', '+15555678901', 'info@spicywings-austin.example.com', 'https://spicywings-austin.example.com', 'https://spicywings.example.com/logo-austin.png', TRUE);

-- Healthy Bites
INSERT INTO restaurants (tenant_id, name, address, city, state, zip_code, country, phone, email, website, logo_url, is_active) VALUES
    (6, 'Healthy Bites San Francisco', '456 Market St', 'San Francisco', 'CA', '94105', 'USA', '+15556789012', 'info@healthybites-sf.example.com', 'https://healthybites-sf.example.com', 'https://healthybites.example.com/logo-sf.png', TRUE),
    (6, 'Healthy Bites Berkeley', '789 University Ave', 'Berkeley', 'CA', '94701', 'USA', '+15557890123', 'info@healthybites-berkeley.example.com', 'https://healthybites-berkeley.example.com', 'https://healthybites.example.com/logo-berkeley.png', TRUE);

-- Dessert Den
INSERT INTO restaurants (tenant_id, name, address, city, state, zip_code, country, phone, email, website, logo_url, is_active) VALUES
    (7, 'Dessert Den New York', '987 Fifth Ave', 'New York', 'NY', '10017', 'USA', '+15558901234', 'info@dessertden-ny.example.com', 'https://dessertden-ny.example.com', 'https://dessertden.example.com/logo-ny.png', TRUE),
    (7, 'Dessert Den Chicago', '135 Michigan Ave', 'Chicago', 'IL', '60601', 'USA', '+15559012345', 'info@dessertden-chicago.example.com', 'https://dessertden-chicago.example.com', 'https://dessertden.example.com/logo-chicago.png', TRUE);

-- Breakfast Club
INSERT INTO restaurants (tenant_id, name, address, city, state, zip_code, country, phone, email, website, logo_url, is_active) VALUES
    (8, 'Breakfast Club Seattle', '654 Pike St', 'Seattle', 'WA', '98101', 'USA', '+15550123456', 'info@breakfastclub-seattle.example.com', 'https://breakfastclub-seattle.example.com', 'https://breakfastclub.example.com/logo-seattle.png', TRUE),
    (8, 'Breakfast Club Portland', '321 NW 23rd Ave', 'Portland', 'OR', '97210', 'USA', '+15551234567', 'info@breakfastclub-portland.example.com', 'https://breakfastclub-portland.example.com', 'https://breakfastclub.example.com/logo-portland.png', TRUE);

-- ============================================
-- Create restaurant hours for each restaurant
-- ============================================
-- Burger Haven
INSERT INTO restaurant_hours (restaurant_id, day_of_week, open_time, close_time) VALUES
    (1, 'Monday', '08:00:00', '22:00:00'),
    (1, 'Tuesday', '08:00:00', '22:00:00'),
    (1, 'Wednesday', '08:00:00', '22:00:00'),
    (1, 'Thursday', '08:00:00', '22:00:00'),
    (1, 'Friday', '08:00:00', '23:00:00'),
    (1, 'Saturday', '09:00:00', '23:00:00'),
    (1, 'Sunday', '10:00:00', '21:00:00');

INSERT INTO restaurant_hours (restaurant_id, day_of_week, open_time, close_time) VALUES
    (2, 'Monday', '08:00:00', '22:00:00'),
    (2, 'Tuesday', '08:00:00', '22:00:00'),
    (2, 'Wednesday', '08:00:00', '22:00:00'),
    (2, 'Thursday', '08:00:00', '22:00:00'),
    (2, 'Friday', '08:00:00', '23:00:00'),
    (2, 'Saturday', '09:00:00', '23:00:00'),
    (2, 'Sunday', '10:00:00', '21:00:00');

-- ============================================
-- Re-enable row-level security
-- ============================================
SET LOCAL row_security = on;