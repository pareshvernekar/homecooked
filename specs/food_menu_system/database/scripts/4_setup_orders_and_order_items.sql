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
    (1, 1, 3, '2026-04-20 12:00:00', 'completed', 16.97, 'Loyalty discount applied for returning customer');

-- Order 2: Spicy Chicken Burger with sides
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (1, 1, 2, '2026-04-21 18:30:00', 'completed', 19.97, 'Delivery order with tip included');

-- Order 3: Veggie Burger with sides
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (1, 1, 3, '2026-04-22 13:15:00', 'completed', 15.97, 'Happy Hour special');

-- Order 4: Kids Meal
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (1, 1, 4, '2026-04-23 16:30:00', 'completed', 12.98, 'Family meal with two kids');

-- ============================================
-- Create sample orders for Burger Haven Midtown (Restaurant 2)
-- ============================================
-- Order 5: Gourmet Burger with Caesar Salad
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (1, 2, 5, '2026-04-24 19:00:00', 'completed', 22.98, 'Dine-in with premium wine pairing');

-- Order 6: Greek Salad with side fries
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (1, 2, 6, '2026-04-25 17:30:00', 'completed', 12.98, 'Health-conscious meal');

-- ============================================
-- Create sample orders for Sushi Palace Tokyo (Restaurant 1)
-- ============================================
-- Order 7: Dragon Roll with Edamame
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (2, 1, 7, '2026-04-26 19:00:00', 'completed', 19.98, 'Takeout order');

-- Order 8: Rainbow Roll with Green Tea
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (2, 1, 8, '2026-04-27 20:30:00', 'completed', 17.98, 'Dine-in with miso soup');

-- ============================================
-- Create sample orders for Sushi Palace Osaka (Restaurant 2)
-- ============================================
-- Order 9: Veggie Roll with Sake
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (2, 2, 9, '2026-04-28 18:00:00', 'completed', 16.98, 'Special occasion meal');

-- Order 10: Tuna Sashimi with Green Tea
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (2, 2, 10, '2026-04-29 19:30:00', 'completed', 20.98, 'Business lunch');

-- ============================================
-- Add order items to orders
-- ============================================
-- Burger Haven Downtown (Restaurant 1)
-- Order 1: Classic Cheeseburger Combo
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (1, 1, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Classic Cheeseburger'), 1, 9.99, 'Main item'),
    (1, 1, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Classic Coke'), 1, 2.49, 'Beverage'),
    (1, 1, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Crispy Fries'), 1, 3.99, 'Side');

-- Order 2: Spicy Chicken Burger with sides
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (1, 2, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Spicy Chicken Burger'), 1, 10.99, 'Main item'),
    (1, 2, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Sweet Potato Fries'), 1, 4.49, 'Side'),
    (1, 2, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Lemonade'), 1, 2.99, 'Beverage');

-- Order 3: Veggie Burger with sides
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (1, 3, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Veggie Burger'), 1, 8.99, 'Main item'),
    (1, 3, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Onion Rings'), 1, 4.99, 'Side'),
    (1, 3, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Iced Tea'), 1, 1.99, 'Beverage');

-- Order 4: Kids Meal
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (1, 4, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Mini Cheeseburger'), 1, 6.99, 'Kids main'),
    (1, 4, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Chicken Tenders'), 1, 5.99, 'Kids side'),
    (1, 4, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Classic Coke'), 1, 2.49, 'Beverage');

-- Burger Haven Midtown (Restaurant 2)
-- Order 5: Gourmet Burger with Caesar Salad
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (1, 5, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 2 AND name = 'Gourmet Burger'), 1, 14.99, 'Main item'),
    (1, 5, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 2 AND name = 'Caesar Salad'), 1, 7.99, 'Side'),
    (1, 5, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 2 AND name = 'Sake'), 1, 5.99, 'Beverage');

-- Order 6: Greek Salad with side fries
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (1, 6, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 2 AND name = 'Greek Salad'), 1, 8.99, 'Main item'),
    (1, 6, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Sweet Potato Fries'), 1, 4.49, 'Side'),
    (1, 6, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Green Tea'), 1, 2.99, 'Beverage');

-- Sushi Palace Tokyo (Restaurant 1)
-- Order 7: Dragon Roll with Edamame
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (2, 7, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Dragon Roll'), 1, 12.99, 'Main item'),
    (2, 7, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Edamame'), 1, 6.99, 'Side'),
    (2, 7, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Green Tea'), 1, 2.99, 'Beverage');

-- Order 8: Rainbow Roll with Green Tea
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (2, 8, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Rainbow Roll'), 1, 14.99, 'Main item'),
    (2, 8, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Spicy Miso Ramen'), 1, 12.99, 'Side'),
    (2, 8, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Sake'), 1, 5.99, 'Beverage');

-- Sushi Palace Osaka (Restaurant 2)
-- Order 9: Veggie Roll with Sake
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (2, 9, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 2 AND name = 'Veggie Roll'), 1, 11.99, 'Main item'),
    (2, 9, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Tuna Sashimi'), 1, 18.99, 'Side'),
    (2, 9, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 2 AND name = 'Sake'), 1, 5.99, 'Beverage');

-- Order 10: Tuna Sashimi with Green Tea
INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (2, 10, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 2 AND name = 'Tuna Sashimi'), 1, 18.99, 'Main item'),
    (2, 10, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Salmon Sashimi'), 1, 16.99, 'Side'),
    (2, 10, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Green Tea'), 1, 2.99, 'Beverage');
    (2, 4, 1, 3.99, 'Onion Rings'),
    (2, 6, 1, 2.99, 'Iced Tea');

-- Order items for Sushi Palace Tokyo (Order ID 3)
INSERT INTO order_items (order_id, menu_item_id, quantity, price, notes)
VALUES
    (3, 7, 3, 12.00, 'Sushi Platter'),
    (3, 8, 1, 8.50, 'Edamame'),
    (3, 9, 1, 5.00, 'Miso Soup');

-- Order items for Sushi Palace Osaka (Order ID 4)
INSERT INTO order_items (order_id, menu_item_id, quantity, price, notes)
VALUES
    (4, 10, 2, 14.00, 'Temaki Rolls'),
    (4, 11, 1, 9.00, 'Grilled Salmon'),
    (4, 12, 1, 4.00, 'Green Tea');

-- ============================================
-- Add some orders with multiple items
-- ============================================
-- Order 11: Combo Meal at Burger Haven Downtown
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (1, 1, 11, '2026-04-30 11:00:00', 'completed', 25.95, 'Happy Hour combo');

INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (1, 11, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Classic Cheeseburger'), 1, 9.99, 'Main item'),
    (1, 11, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Sweet Potato Fries'), 1, 4.49, 'Side'),
    (1, 11, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Chocolate Chip Cookie'), 1, 2.99, 'Dessert'),
    (1, 11, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Classic Coke'), 1, 2.49, 'Beverage');

-- Order 12: Sushi Platter at Sushi Palace Tokyo
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (2, 1, 12, '2026-05-01 18:00:00', 'completed', 45.95, 'Group order for 4 people');

INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (2, 12, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Dragon Roll'), 2, 12.99, 'Main item'),
    (2, 12, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Rainbow Roll'), 2, 14.99, 'Main item'),
    (2, 12, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Tuna Sashimi'), 1, 18.99, 'Side'),
    (2, 12, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Salmon Sashimi'), 1, 16.99, 'Side'),
    (2, 12, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Edamame'), 1, 6.99, 'Side'),
    (2, 12, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Green Tea'), 4, 2.99, 'Beverage');

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
    (3, 1, 13, '2026-05-02 12:30:00', 'completed', 17.98, 'Vegetarian meal');

INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (3, 13, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Portobello Mushroom Burger'), 1, 10.99, 'Main item'),
    (3, 13, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Kale Caesar'), 1, 8.99, 'Side'),
    (3, 13, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Iced Tea'), 1, 1.99, 'Beverage');

-- Order 14: Italian Eats - Spicy Chicken Burger with Pizza
INSERT INTO orders (tenant_id, restaurant_id, user_id, order_date, status, total_amount, notes)
VALUES
    (4, 1, 14, '2026-05-03 19:00:00', 'completed', 28.98, 'Dine-in with pasta');

INSERT INTO order_items (tenant_id, order_id, menu_item_id, quantity, price, notes)
VALUES
    (4, 14, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Spicy Chicken Burger'), 1, 10.99, 'Main item'),
    (4, 14, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Margherita Pizza'), 1, 12.99, 'Side'),
    (4, 14, (SELECT menu_item_id FROM menu_items WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Italian Soda'), 1, 2.99, 'Beverage');