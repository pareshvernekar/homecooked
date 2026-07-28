-- ============================================
-- Sample Data Insertion Script: Orders and Order Items
-- File: 4_setup_orders_and_order_items.sql
-- Description: This script creates sample orders and order items for testing.
-- ============================================

-- Enable row-level security temporarily for tenant-specific tables
SET LOCAL row_security = off;

-- ============================================
-- Create sample orders for Burger Haven Downtown (Restaurant 1)
-- ============================================
-- Order 1: Classic Cheeseburger Combo
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_1', 'restaurant_t1_1', 'user_t1_customer1', '2026-04-20 12:00:00', 'completed', 16.97, 'Loyalty discount applied for returning customer');

-- Order 2: Spicy Chicken Burger with sides
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_1', 'restaurant_t1_1', 'user_t1_manager', '2026-04-21 18:30:00', 'completed', 19.97, 'Delivery order with tip included');

-- Order 3: Veggie Burger with sides
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_1', 'restaurant_t1_1', 'user_t1_customer1', '2026-04-22 13:15:00', 'completed', 15.97, 'Happy Hour special');

-- Order 4: Kids Meal
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_1', 'restaurant_t1_1', 'user_t1_customer1', '2026-04-23 16:30:00', 'completed', 12.98, 'Family meal with two kids');

-- ============================================
-- Create sample orders for Burger Haven Midtown (Restaurant 2)
-- ============================================
-- Order 5: Gourmet Burger with Caesar Salad
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_1', 'restaurant_t1_2', 'user_t1_customer1', '2026-04-24 19:00:00', 'completed', 22.98, 'Dine-in with premium wine pairing');

-- Order 6: Greek Salad with side fries
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_1', 'restaurant_t1_2', 'user_t1_customer1', '2026-04-25 17:30:00', 'completed', 12.98, 'Health-conscious meal');

-- ============================================
-- Create sample orders for Sushi Palace Tokyo (Restaurant 1)
-- ============================================
-- Order 7: Dragon Roll with Edamame
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_2', 'restaurant_t2_1', 'user_t2_admin', '2026-04-26 19:00:00', 'completed', 19.98, 'Takeout order');

-- Order 8: Rainbow Roll with Green Tea
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_2', 'restaurant_t2_1', 'user_t2_manager', '2026-04-27 20:30:00', 'completed', 17.98, 'Dine-in with miso soup');

-- ============================================
-- Create sample orders for Sushi Palace Osaka (Restaurant 2)
-- ============================================
-- Order 9: Veggie Roll with Sake
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_2', 'restaurant_t2_2', 'user_t2_customer1', '2026-04-28 18:00:00', 'completed', 16.98, 'Special occasion meal');

-- Order 10: Tuna Sashimi with Green Tea
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_2', 'restaurant_t2_2', 'user_t2_customer1', '2026-04-29 19:30:00', 'completed', 20.98, 'Business lunch');

-- ============================================
-- Add order items to orders
-- ============================================
-- Burger Haven Downtown (Restaurant 1)
-- Order 1: Classic Cheeseburger Combo
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_1', 1, 'mi_t1_r1_classic_cheeseburger', 1, 9.99, 'Main item'),
    ('tenant_1', 1, 'mi_t1_r1_classic_coke', 1, 2.49, 'Beverage'),
    ('tenant_1', 1, 'mi_t1_r1_crispy_fries', 1, 3.99, 'Side');

-- Order 2: Spicy Chicken Burger with sides
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_1', 2, 'mi_t1_r1_spicy_chicken_burger', 1, 10.99, 'Main item'),
    ('tenant_1', 2, 'mi_t1_r1_sweet_potato_fries', 1, 4.49, 'Side'),
    ('tenant_1', 2, 'mi_t1_r1_lemonade', 1, 2.99, 'Beverage');

-- Order 3: Veggie Burger with sides
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_1', 3, 'mi_t1_r1_veggie_burger', 1, 8.99, 'Main item'),
    ('tenant_1', 3, 'mi_t1_r1_onion_rings', 1, 4.99, 'Side'),
    ('tenant_1', 3, 'mi_t1_r1_iced_tea', 1, 1.99, 'Beverage');

-- Order 4: Kids Meal
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_1', 4, 'mi_t1_r1_mini_cheeseburger', 1, 6.99, 'Kids main'),
    ('tenant_1', 4, 'mi_t1_r1_chicken_tenders', 1, 5.99, 'Kids side'),
    ('tenant_1', 4, 'mi_t1_r1_classic_coke', 1, 2.49, 'Beverage');

-- Burger Haven Midtown (Restaurant 2)
-- Order 5: Gourmet Burger with Caesar Salad
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_1', 5, 'mi_t1_r2_gourmet_burger', 1, 14.99, 'Main item'),
    ('tenant_1', 5, 'mi_t1_r2_caesar_salad', 1, 7.99, 'Side');

-- Order 6: Greek Salad with side fries
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_1', 6, 'mi_t1_r2_greek_salad', 1, 8.99, 'Main item'),
    ('tenant_1', 6, 'mi_t1_r1_sweet_potato_fries', 1, 4.49, 'Side'),
    ('tenant_1', 6, 'mi_t1_r1_green_tea', 1, 2.99, 'Beverage');

-- Sushi Palace Tokyo (Restaurant 1)
-- Order 7: Dragon Roll with Edamame
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_2', 7, 'mi_t2_r1_dragon_roll', 1, 12.99, 'Main item'),
    ('tenant_2', 7, 'mi_t2_r1_edamame', 1, 6.99, 'Side'),
    ('tenant_2', 7, 'mi_t2_r1_green_tea', 1, 2.99, 'Beverage');

-- Order 8: Rainbow Roll with Green Tea
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_2', 8, 'mi_t2_r1_rainbow_roll', 1, 14.99, 'Main item'),
    ('tenant_2', 8, 'mi_t2_r1_spicy_miso_ramen', 1, 12.99, 'Side'),
    ('tenant_2', 8, 'mi_t2_r1_sake', 1, 5.99, 'Beverage');

-- Sushi Palace Osaka (Restaurant 2)
-- Order 9: Veggie Roll with Sake
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_2', 9, 'mi_t2_r2_sushi_udon_special', 1, 11.99, 'Main item'),
    ('tenant_2', 9, 'mi_t2_r2_udon_noodles', 1, 10.99, 'Side'),
    ('tenant_2', 9, 'mi_t2_r2_dessert_mochi', 1, 4.99, 'Dessert');

-- Order 10: Tuna Sashimi with Green Tea
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_2', 10, 'mi_t2_r2_tuna_sashimi', 1, 18.99, 'Main item'),
    ('tenant_2', 10, 'mi_t2_r1_salmon_sashimi', 1, 16.99, 'Side'),
    ('tenant_2', 10, 'mi_t2_r1_green_tea', 1, 2.99, 'Beverage');

-- Order items for Sushi Palace Tokyo (Order ID 3)
-- (Removed legacy numeric-order-item examples; use tenant-scoped inserts above)

-- ============================================
-- Add some orders with multiple items
-- ============================================
-- Order 11: Combo Meal at Burger Haven Downtown
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (1, 1, 11, '2026-04-30 11:00:00', 'completed', 25.95, 'Happy Hour combo');

INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_1', 11, 'mi_t1_r1_classic_cheeseburger', 1, 9.99, 'Main item'),
    ('tenant_1', 11, 'mi_t1_r1_sweet_potato_fries', 1, 4.49, 'Side'),
    ('tenant_1', 11, 'mi_t1_r1_chocolate_chip_cookie', 1, 2.99, 'Dessert'),
    ('tenant_1', 11, 'mi_t1_r1_classic_coke', 1, 2.49, 'Beverage');

-- Order 12: Sushi Platter at Sushi Palace Tokyo
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (2, 1, 12, '2026-05-01 18:00:00', 'completed', 45.95, 'Group order for 4 people');

INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_2', 12, 'mi_t2_r1_dragon_roll', 2, 12.99, 'Main item'),
    ('tenant_2', 12, 'mi_t2_r1_rainbow_roll', 2, 14.99, 'Main item'),
    ('tenant_2', 12, 'mi_t2_r1_tuna_sashimi', 1, 18.99, 'Side'),
    ('tenant_2', 12, 'mi_t2_r1_salmon_sashimi', 1, 16.99, 'Side'),
    ('tenant_2', 12, 'mi_t2_r1_edamame', 1, 6.99, 'Side'),
    ('tenant_2', 12, 'mi_t2_r1_green_tea', 4, 2.99, 'Beverage');

-- ============================================
-- Re-enable row-level security
-- ============================================
SET LOCAL row_security = on;

-- ============================================
-- Add a few more orders for variety
-- ============================================
-- Order 13: Veggie Delight - Portobello Burger with Salad
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_3', 'restaurant_t3_1', 'user_t3_admin', '2026-05-02 12:30:00', 'completed', 17.98, 'Vegetarian meal');

INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_3', 13, 'mi_t3_r1_portobello_burger', 1, 10.99, 'Main item'),
    ('tenant_3', 13, 'mi_t3_r1_kale_caesar', 1, 8.99, 'Side'),
    ('tenant_3', 13, 'mi_t3_r1_iced_tea', 1, 1.99, 'Beverage');

-- Order 14: Italian Eats - Spicy Chicken Burger with Pizza
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    ('tenant_4', 'restaurant_t4_1', 'user_t4_admin', '2026-05-03 19:00:00', 'completed', 28.98, 'Dine-in with pasta');

INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    ('tenant_4', 14, 'mi_t4_r1_spicy_chicken_burger', 1, 10.99, 'Main item'),
    ('tenant_4', 14, 'mi_t4_r1_margherita_pizza', 1, 12.99, 'Side');