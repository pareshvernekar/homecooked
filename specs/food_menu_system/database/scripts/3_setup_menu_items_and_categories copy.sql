-- ============================================
-- Sample Data Insertion Script: Menu Items and Categories
-- File: 3_setup_menu_items_and_categories.sql
-- Description: This script creates sample menu items and categories for each restaurant
-- ============================================

-- Enable row-level security temporarily for tenant-specific tables
SET LOCAL row_security = off;

-- ============================================
-- Create sample menu categories
-- ============================================
-- Create tenant_restaurant_categories with explicit category_id and string tenant/restaurant IDs
INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, is_active, created_by) VALUES
    -- Burger Haven (restaurant_t1_1)
    ('cat_t1_r1_burgers','tenant_1','restaurant_t1_1','Burgers','Classic beef burgers with various toppings',0,TRUE,'user_t1_admin'),
    ('cat_t1_r1_sides','tenant_1','restaurant_t1_1','Sides','Sides to complement your burger',1,TRUE,'user_t1_admin'),
    ('cat_t1_r1_drinks','tenant_1','restaurant_t1_1','Drinks','Refreshing beverages',2,TRUE,'user_t1_admin'),
    ('cat_t1_r1_desserts','tenant_1','restaurant_t1_1','Desserts','Sweet treats to end your meal',3,TRUE,'user_t1_admin'),
    ('cat_t1_r1_kids','tenant_1','restaurant_t1_1','Kids Menu','Kid-friendly options',4,TRUE,'user_t1_admin'),
    -- Burger Haven (restaurant_t1_2)
    ('cat_t1_r2_burgers','tenant_1','restaurant_t1_2','Burgers','Classic beef burgers with various toppings',0,TRUE,'user_t1_admin'),
    ('cat_t1_r2_sides','tenant_1','restaurant_t1_2','Sides','Sides to complement your burger',1,TRUE,'user_t1_admin'),
    ('cat_t1_r2_drinks','tenant_1','restaurant_t1_2','Drinks','Refreshing beverages',2,TRUE,'user_t1_admin'),
    ('cat_t1_r2_desserts','tenant_1','restaurant_t1_2','Desserts','Sweet treats to end your meal',3,TRUE,'user_t1_admin'),
    ('cat_t1_r2_salads','tenant_1','restaurant_t1_2','Salads','Fresh salads',4,TRUE,'user_t1_admin')
ON CONFLICT (category_id) DO NOTHING;

INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, is_active, created_by) VALUES
    -- Sushi Palace
    ('cat_t2_r1_sushi','tenant_2','restaurant_t2_1','Sushi','Freshly prepared sushi rolls',0,TRUE,'user_t2_admin'),
    ('cat_t2_r1_sashimi','tenant_2','restaurant_t2_1','Sashimi','Thinly sliced raw fish',1,TRUE,'user_t2_admin'),
    ('cat_t2_r1_tempura','tenant_2','restaurant_t2_1','Tempura','Lightly battered and fried dishes',2,TRUE,'user_t2_admin'),
    ('cat_t2_r1_ramen','tenant_2','restaurant_t2_1','Ramen','Rich noodle soups',3,TRUE,'user_t2_admin'),
    ('cat_t2_r1_appetizers','tenant_2','restaurant_t2_1','Appetizers','Small dishes to start your meal',4,TRUE,'user_t2_admin'),
    ('cat_t2_r1_drinks','tenant_2','restaurant_t2_1','Drinks','Traditional Japanese beverages',5,TRUE,'user_t2_admin'),
    ('cat_t2_r2_sushi','tenant_2','restaurant_t2_2','Sushi','Freshly prepared sushi rolls',0,TRUE,'user_t2_admin'),
    ('cat_t2_r2_sashimi','tenant_2','restaurant_t2_2','Sashimi','Thinly sliced raw fish',1,TRUE,'user_t2_admin'),
    ('cat_t2_r2_udon','tenant_2','restaurant_t2_2','Udon','Thick wheat noodles',2,TRUE,'user_t2_admin'),
    ('cat_t2_r2_drinks','tenant_2','restaurant_t2_2','Drinks','Traditional Japanese beverages',3,TRUE,'user_t2_admin'),
    ('cat_t2_r2_desserts','tenant_2','restaurant_t2_2','Desserts','Japanese sweets',4,TRUE,'user_t2_admin')
ON CONFLICT (category_id) DO NOTHING;

INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, is_active, created_by) VALUES
    -- Veggie Delight
    ('cat_t3_r1_vegetarian','tenant_3','restaurant_t3_1','Vegetarian','100% plant-based meals',0,TRUE,'user_t3_admin'),
    ('cat_t3_r1_vegan','tenant_3','restaurant_t3_1','Vegan','Completely plant-based options',1,TRUE,'user_t3_admin'),
    ('cat_t3_r1_salads','tenant_3','restaurant_t3_1','Salads','Fresh and healthy salads',2,TRUE,'user_t3_admin'),
    ('cat_t3_r1_soups','tenant_3','restaurant_t3_1','Soups','Warm and comforting soups',3,TRUE,'user_t3_admin'),
    ('cat_t3_r1_sandwiches','tenant_3','restaurant_t3_1','Sandwiches','Freshly made sandwiches',4,TRUE,'user_t3_admin'),
    ('cat_t3_r1_desserts','tenant_3','restaurant_t3_1','Desserts','Healthy sweet treats',5,TRUE,'user_t3_admin'),
    ('cat_t3_r2_vegetarian','tenant_3','restaurant_t3_2','Vegetarian','100% plant-based meals',0,TRUE,'user_t3_admin'),
    ('cat_t3_r2_vegan','tenant_3','restaurant_t3_2','Vegan','Completely plant-based options',1,TRUE,'user_t3_admin'),
    ('cat_t3_r2_burgers','tenant_3','restaurant_t3_2','Burgers','Veggie burgers',2,TRUE,'user_t3_admin'),
    ('cat_t3_r2_drinks','tenant_3','restaurant_t3_2','Drinks','Healthy beverage options',3,TRUE,'user_t3_admin'),
    ('cat_t3_r2_breakfast','tenant_3','restaurant_t3_2','Breakfast','Morning meals',4,TRUE,'user_t3_admin')
ON CONFLICT (category_id) DO NOTHING;

INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, is_active, created_by) VALUES
    -- Italian Eats
    ('cat_t4_r1_pasta','tenant_4','restaurant_t4_1','Pasta','Classic Italian pasta dishes',0,TRUE,'user_t4_admin'),
    ('cat_t4_r1_pizza','tenant_4','restaurant_t4_1','Pizza','Authentic Italian pizza',1,TRUE,'user_t4_admin'),
    ('cat_t4_r1_antipasti','tenant_4','restaurant_t4_1','Antipasti','Italian starters',2,TRUE,'user_t4_admin'),
    ('cat_t4_r1_risotto','tenant_4','restaurant_t4_1','Risotto','Creamy rice dishes',3,TRUE,'user_t4_admin'),
    ('cat_t4_r1_desserts','tenant_4','restaurant_t4_1','Desserts','Italian sweets',4,TRUE,'user_t4_admin'),
    ('cat_t4_r1_drinks','tenant_4','restaurant_t4_1','Drinks','Italian wines and beverages',5,TRUE,'user_t4_admin'),
    ('cat_t4_r2_pasta','tenant_4','restaurant_t4_2','Pasta','Classic Italian pasta dishes',0,TRUE,'user_t4_admin'),
    ('cat_t4_r2_pizza','tenant_4','restaurant_t4_2','Pizza','Authentic Italian pizza',1,TRUE,'user_t4_admin'),
    ('cat_t4_r2_appetizers','tenant_4','restaurant_t4_2','Appetizers','Italian starters',2,TRUE,'user_t4_admin'),
    ('cat_t4_r2_drinks','tenant_4','restaurant_t4_2','Drinks','Italian wines and beverages',3,TRUE,'user_t4_admin'),
    ('cat_t4_r2_desserts','tenant_4','restaurant_t4_2','Desserts','Italian sweets',4,TRUE,'user_t4_admin')
ON CONFLICT (category_id) DO NOTHING;

INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, is_active, created_by) VALUES
    -- Spicy Wings
    ('cat_t5_r1_wings','tenant_5','restaurant_t5_1','Wings','Spicy and crispy wings',0,TRUE,'user_t5_admin'),
    ('cat_t5_r1_sandwiches','tenant_5','restaurant_t5_1','Sandwiches','Meat-filled sandwiches',1,TRUE,'user_t5_admin'),
    ('cat_t5_r1_burgers','tenant_5','restaurant_t5_1','Burgers','Spicy beef burgers',2,TRUE,'user_t5_admin'),
    ('cat_t5_r1_sides','tenant_5','restaurant_t5_1','Sides','Spicy and tangy sides',3,TRUE,'user_t5_admin'),
    ('cat_t5_r1_drinks','tenant_5','restaurant_t5_1','Drinks','Cold beverages',4,TRUE,'user_t5_admin'),
    ('cat_t5_r1_desserts','tenant_5','restaurant_t5_1','Desserts','Sweet treats',5,TRUE,'user_t5_admin'),
    ('cat_t5_r2_wings','tenant_5','restaurant_t5_2','Wings','Spicy and crispy wings',0,TRUE,'user_t5_admin'),
    ('cat_t5_r2_sandwiches','tenant_5','restaurant_t5_2','Sandwiches','Meat-filled sandwiches',1,TRUE,'user_t5_admin'),
    ('cat_t5_r2_burgers','tenant_5','restaurant_t5_2','Burgers','Spicy beef burgers',2,TRUE,'user_t5_admin'),
    ('cat_t5_r2_drinks','tenant_5','restaurant_t5_2','Drinks','Cold beverages',3,TRUE,'user_t5_admin'),
    ('cat_t5_r2_desserts','tenant_5','restaurant_t5_2','Desserts','Sweet treats',4,TRUE,'user_t5_admin')
ON CONFLICT (category_id) DO NOTHING;

INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, is_active, created_by) VALUES
    -- Healthy Bites
    ('cat_t6_r1_salads','tenant_6','restaurant_t6_1','Salads','Fresh and healthy salads',0,TRUE,'user_t6_admin'),
    ('cat_t6_r1_soups','tenant_6','restaurant_t6_1','Soups','Light and nutritious soups',1,TRUE,'user_t6_admin'),
    ('cat_t6_r1_sandwiches','tenant_6','restaurant_t6_1','Sandwiches','Whole grain sandwiches',2,TRUE,'user_t6_admin'),
    ('cat_t6_r1_bowls','tenant_6','restaurant_t6_1','Bowls','Nutritious grain bowls',3,TRUE,'user_t6_admin'),
    ('cat_t6_r1_drinks','tenant_6','restaurant_t6_1','Drinks','Healthy beverage options',4,TRUE,'user_t6_admin'),
    ('cat_t6_r1_desserts','tenant_6','restaurant_t6_1','Desserts','Low-sugar desserts',5,TRUE,'user_t6_admin'),
    ('cat_t6_r2_salads','tenant_6','restaurant_t6_2','Salads','Fresh and healthy salads',0,TRUE,'user_t6_admin'),
    ('cat_t6_r2_soups','tenant_6','restaurant_t6_2','Soups','Light and nutritious soups',1,TRUE,'user_t6_admin'),
    ('cat_t6_r2_sandwiches','tenant_6','restaurant_t6_2','Sandwiches','Whole grain sandwiches',2,TRUE,'user_t6_admin'),
    ('cat_t6_r2_drinks','tenant_6','restaurant_t6_2','Drinks','Healthy beverage options',3,TRUE,'user_t6_admin'),
    ('cat_t6_r2_breakfast','tenant_6','restaurant_t6_2','Breakfast','Healthy morning meals',4,TRUE,'user_t6_admin')
ON CONFLICT (category_id) DO NOTHING;

INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, is_active, created_by) VALUES
    -- Dessert Den
    ('cat_t7_r1_cakes','tenant_7','restaurant_t7_1','Cakes','Beautifully decorated cakes',0,TRUE,'user_t7_admin'),
    ('cat_t7_r1_cookies','tenant_7','restaurant_t7_1','Cookies','Freshly baked cookies',1,TRUE,'user_t7_admin'),
    ('cat_t7_r1_icecream','tenant_7','restaurant_t7_1','Ice Cream','Premium ice cream',2,TRUE,'user_t7_admin'),
    ('cat_t7_r1_pastries','tenant_7','restaurant_t7_1','Pastries','Delicious pastries',3,TRUE,'user_t7_admin'),
    ('cat_t7_r1_coffee','tenant_7','restaurant_t7_1','Coffee','Specialty coffee drinks',4,TRUE,'user_t7_admin'),
    ('cat_t7_r1_tea','tenant_7','restaurant_t7_1','Tea','Fine teas',5,TRUE,'user_t7_admin'),
    ('cat_t7_r2_cakes','tenant_7','restaurant_t7_2','Cakes','Beautifully decorated cakes',0,TRUE,'user_t7_admin'),
    ('cat_t7_r2_cookies','tenant_7','restaurant_t7_2','Cookies','Freshly baked cookies',1,TRUE,'user_t7_admin'),
    ('cat_t7_r2_icecream','tenant_7','restaurant_t7_2','Ice Cream','Premium ice cream',2,TRUE,'user_t7_admin'),
    ('cat_t7_r2_pastries','tenant_7','restaurant_t7_2','Pastries','Delicious pastries',3,TRUE,'user_t7_admin'),
    ('cat_t7_r2_drinks','tenant_7','restaurant_t7_2','Drinks','Refreshing beverages',4,TRUE,'user_t7_admin')
ON CONFLICT (category_id) DO NOTHING;

INSERT INTO tenant_restaurant_categories (category_id, tenant_id, restaurant_id, name, description, display_order, is_active, created_by) VALUES
    -- Breakfast Club
    ('cat_t8_r1_breakfast','tenant_8','restaurant_t8_1','Breakfast','Classic breakfast items',0,TRUE,'user_t8_admin'),
    ('cat_t8_r1_bacon','tenant_8','restaurant_t8_1','Bacon','Various bacon dishes',1,TRUE,'user_t8_admin'),
    ('cat_t8_r1_eggs','tenant_8','restaurant_t8_1','Eggs','Egg-based dishes',2,TRUE,'user_t8_admin'),
    ('cat_t8_r1_pancakes','tenant_8','restaurant_t8_1','Pancakes','Fluffy pancakes',3,TRUE,'user_t8_admin'),
    ('cat_t8_r1_drinks','tenant_8','restaurant_t8_1','Drinks','Morning beverages',4,TRUE,'user_t8_admin'),
    ('cat_t8_r1_desserts','tenant_8','restaurant_t8_1','Desserts','Sweet breakfast treats',5,TRUE,'user_t8_admin'),
    ('cat_t8_r2_breakfast','tenant_8','restaurant_t8_2','Breakfast','Classic breakfast items',0,TRUE,'user_t8_admin'),
    ('cat_t8_r2_bacon','tenant_8','restaurant_t8_2','Bacon','Various bacon dishes',1,TRUE,'user_t8_admin'),
    ('cat_t8_r2_eggs','tenant_8','restaurant_t8_2','Eggs','Egg-based dishes',2,TRUE,'user_t8_admin'),
    ('cat_t8_r2_pancakes','tenant_8','restaurant_t8_2','Pancakes','Fluffy pancakes',3,TRUE,'user_t8_admin'),
    ('cat_t8_r2_drinks','tenant_8','restaurant_t8_2','Drinks','Morning beverages',4,TRUE,'user_t8_admin'),
    ('cat_t8_r2_desserts','tenant_8','restaurant_t8_2','Desserts','Sweet breakfast treats',5,TRUE,'user_t8_admin')
ON CONFLICT (category_id) DO NOTHING;

-- ============================================
-- Create sample menu items
-- ============================================
-- Insert tenant_menu_items with explicit IDs and reference the category_ids inserted above
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    -- Burger Haven - Restaurant 1
    ('mi_t1_r1_classic_cheeseburger','tenant_1','restaurant_t1_1','cat_t1_r1_burgers','Classic Cheeseburger','Juicy beef patty with melted cheddar, lettuce, tomato, onion, pickles, and special sauce on a brioche bun',9.99,9.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,10,'user_t1_admin'),
    ('mi_t1_r1_spicy_chicken_burger','tenant_1','restaurant_t1_1','cat_t1_r1_burgers','Spicy Chicken Burger','Grilled chicken patty with spicy mayo, pepper jack cheese, bacon, and sriracha aioli',10.99,10.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,TRUE,12,'user_t1_admin'),
    ('mi_t1_r1_veggie_burger','tenant_1','restaurant_t1_1','cat_t1_r1_burgers','Veggie Burger','Grilled portobello mushroom patty with avocado, spinach, and vegan cheese',8.99,8.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t1_admin'),
    ('mi_t1_r1_crispy_fries','tenant_1','restaurant_t1_1','cat_t1_r1_sides','Crispy Fries','Thick-cut fries with sea salt',3.99,3.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t1_admin'),
    ('mi_t1_r1_sweet_potato_fries','tenant_1','restaurant_t1_1','cat_t1_r1_sides','Sweet Potato Fries','Sweet potato fries with garlic parmesan',4.49,4.49,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,9,'user_t1_admin'),
    ('mi_t1_r1_onion_rings','tenant_1','restaurant_t1_1','cat_t1_r1_sides','Onion Rings','Crispy onion rings with ranch',4.99,4.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t1_admin'),
    ('mi_t1_r1_classic_coke','tenant_1','restaurant_t1_1','cat_t1_r1_drinks','Classic Coke','Original Coca-Cola',2.49,2.49,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,2,'user_t1_admin'),
    ('mi_t1_r1_iced_tea','tenant_1','restaurant_t1_1','cat_t1_r1_drinks','Iced Tea','Sweet iced tea',1.99,1.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,3,'user_t1_admin'),
    ('mi_t1_r1_lemonade','tenant_1','restaurant_t1_1','cat_t1_r1_drinks','Lemonade','Fresh-squeezed lemonade',2.99,2.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,4,'user_t1_admin'),
    ('mi_t1_r1_chocolate_chip_cookie','tenant_1','restaurant_t1_1','cat_t1_r1_desserts','Chocolate Chip Cookie','Warm chocolate chip cookie',2.99,2.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t1_admin'),
    ('mi_t1_r1_brownies','tenant_1','restaurant_t1_1','cat_t1_r1_desserts','Brownies','Fudgy chocolate brownies',3.49,3.49,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,6,'user_t1_admin'),
    ('mi_t1_r1_mini_cheeseburger','tenant_1','restaurant_t1_1','cat_t1_r1_kids','Mini Cheeseburger','Smaller version of our classic cheeseburger',6.99,6.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,8,'user_t1_admin'),
    ('mi_t1_r1_chicken_tenders','tenant_1','restaurant_t1_1','cat_t1_r1_kids','Chicken Tenders','Crispy chicken tenders with honey mustard',5.99,5.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,10,'user_t1_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Additional items for Burger Haven restaurant 2
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    ('mi_t1_r2_gourmet_burger','tenant_1','restaurant_t1_2','cat_t1_r2_burgers','Gourmet Burger','Premium beef patty with truffle aioli, caramelized onions, and blue cheese',14.99,14.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,12,'user_t1_admin'),
    ('mi_t1_r2_caesar_salad','tenant_1','restaurant_t1_2','cat_t1_r2_salads','Caesar Salad','Romaine lettuce with Caesar dressing, croutons, and parmesan',7.99,7.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,8,'user_t1_admin'),
    ('mi_t1_r2_greek_salad','tenant_1','restaurant_t1_2','cat_t1_r2_salads','Greek Salad','Mixed greens with cucumber, tomato, olives, feta, and Greek dressing',8.99,8.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t1_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Sushi Palace - Restaurant 1 and 2 (tenant_2)
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    -- Restaurant 1
    ('mi_t2_r1_dragon_roll','tenant_2','restaurant_t2_1','cat_t2_r1_sushi','Dragon Roll','Avocado, cucumber, and tempura crunch wrapped in nori with spicy mayo',12.99,12.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,15,'user_t2_admin'),
    ('mi_t2_r1_rainbow_roll','tenant_2','restaurant_t2_1','cat_t2_r1_sushi','Rainbow Roll','Assorted fish and avocado with eel sauce',14.99,14.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,18,'user_t2_admin'),
    ('mi_t2_r1_veggie_roll','tenant_2','restaurant_t2_1','cat_t2_r1_sushi','Veggie Roll','Avocado, cucumber, carrot, and pickled radish with spicy mayo',11.99,11.99,'USD',TRUE,FALSE,TRUE,TRUE,FALSE,FALSE,15,'user_t2_admin'),
    ('mi_t2_r1_tuna_sashimi','tenant_2','restaurant_t2_1','cat_t2_r1_sashimi','Tuna Sashimi','Fresh sashimi-grade tuna',18.99,18.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t2_admin'),
    ('mi_t2_r1_salmon_sashimi','tenant_2','restaurant_t2_1','cat_t2_r1_sashimi','Salmon Sashimi','Fresh sashimi-grade salmon',16.99,16.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t2_admin'),
    ('mi_t2_r1_tempura_shrimp','tenant_2','restaurant_t2_1','cat_t2_r1_tempura','Tempura Shrimp','Lightly battered shrimp with dipping sauce',14.99,14.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,12,'user_t2_admin'),
    ('mi_t2_r1_spicy_miso_ramen','tenant_2','restaurant_t2_1','cat_t2_r1_ramen','Spicy Miso Ramen','Rich miso broth with chashu pork, soft-boiled egg, and nori',12.99,12.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,20,'user_t2_admin'),
    ('mi_t2_r1_edamame','tenant_2','restaurant_t2_1','cat_t2_r1_appetizers','Edamame','Steamed soybeans with sea salt',6.99,6.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t2_admin'),
    ('mi_t2_r1_green_tea','tenant_2','restaurant_t2_1','cat_t2_r1_drinks','Green Tea','Freshly brewed green tea',2.99,2.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,3,'user_t2_admin'),
    ('mi_t2_r1_sake','tenant_2','restaurant_t2_1','cat_t2_r1_drinks','Sake','Premium Japanese sake',5.99,5.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t2_admin'),
    -- Restaurant 2
    ('mi_t2_r2_sushi_udon_special','tenant_2','restaurant_t2_2','cat_t2_r2_sushi','Sushi Assortment','Chef selection of assorted sushi rolls',13.99,13.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,15,'user_t2_admin'),
    ('mi_t2_r2_udon_noodles','tenant_2','restaurant_t2_2','cat_t2_r2_udon','Udon Noodles','Thick wheat noodles in savory broth',10.99,10.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,12,'user_t2_admin'),
    ('mi_t2_r2_dessert_mochi','tenant_2','restaurant_t2_2','cat_t2_r2_desserts','Mochi Ice Cream','Assorted mochi ice cream',4.99,4.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t2_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Veggie Delight - Restaurant 1 and 2 (tenant_3)
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    ('mi_t3_r1_portobello_burger','tenant_3','restaurant_t3_1','cat_t3_r1_vegetarian','Portobello Mushroom Burger','Grilled portobello with avocado, spinach, and vegan cheese',10.99,10.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,12,'user_t3_admin'),
    ('mi_t3_r1_black_bean_burger','tenant_3','restaurant_t3_1','cat_t3_r1_vegetarian','Black Bean Burger','Spiced black bean patty with roasted peppers and vegan mayo',9.99,9.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t3_admin'),
    ('mi_t3_r1_beyond_burger','tenant_3','restaurant_t3_1','cat_t3_r1_vegan','Beyond Burger','Plant-based beef patty with all the fixings',10.99,10.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t3_admin'),
    ('mi_t3_r1_kale_caesar','tenant_3','restaurant_t3_1','cat_t3_r1_salads','Kale Caesar','Kale with Caesar dressing, croutons, and vegan parmesan',8.99,8.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t3_admin'),
    ('mi_t3_r1_quinoa_salad','tenant_3','restaurant_t3_1','cat_t3_r1_salads','Quinoa Salad','Quinoa with roasted veggies and lemon vinaigrette',9.99,9.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t3_admin'),
    ('mi_t3_r1_minestrone','tenant_3','restaurant_t3_1','cat_t3_r1_soups','Minestrone','Vegetable soup with beans and pasta',5.99,5.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,15,'user_t3_admin'),
    ('mi_t3_r1_veggie_sandwich','tenant_3','restaurant_t3_1','cat_t3_r1_sandwiches','Veggie Sandwich','Whole grain bread with hummus, roasted veggies, and avocado',8.99,8.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t3_admin'),
    ('mi_t3_r1_vegan_chocolate_cake','tenant_3','restaurant_t3_1','cat_t3_r1_desserts','Vegan Chocolate Cake','Rich chocolate cake with vegan frosting',5.99,5.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t3_admin'),
    ('mi_t3_r2_veggie_burger','tenant_3','restaurant_t3_2','cat_t3_r2_burgers','Veggie Burger (r2)','Veggie burger for restaurant 2',9.49,9.49,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t3_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Italian Eats - Restaurant 1 and 2 (tenant_4)
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    ('mi_t4_r1_spaghetti_carbonara','tenant_4','restaurant_t4_1','cat_t4_r1_pasta','Spaghetti Carbonara','Spaghetti with egg, cheese, pancetta, and black pepper',14.99,14.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,15,'user_t4_admin'),
    ('mi_t4_r1_penne_arrabbiata','tenant_4','restaurant_t4_1','cat_t4_r1_pasta','Penne Arrabbiata','Spicy tomato sauce with penne pasta',12.99,12.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,TRUE,15,'user_t4_admin'),
    ('mi_t4_r1_margherita_pizza','tenant_4','restaurant_t4_1','cat_t4_r1_pizza','Margherita Pizza','Classic tomato sauce, mozzarella, and basil',12.99,12.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,20,'user_t4_admin'),
    ('mi_t4_r1_veggie_pizza','tenant_4','restaurant_t4_1','cat_t4_r1_pizza','Veggie Pizza','Tomato sauce, mozzarella, bell peppers, onions, olives, and mushrooms',13.99,13.99,'USD',TRUE,FALSE,TRUE,FALSE,FALSE,FALSE,20,'user_t4_admin'),
    ('mi_t4_r1_bruschetta','tenant_4','restaurant_t4_1','cat_t4_r1_antipasti','Bruschetta','Toasted bread with tomato, garlic, basil, and balsamic glaze',6.99,6.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t4_admin'),
    ('mi_t4_r1_risotto_funghi','tenant_4','restaurant_t4_1','cat_t4_r1_risotto','Risotto ai Funghi','Creamy risotto with wild mushrooms',13.99,13.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,20,'user_t4_admin'),
    ('mi_t4_r1_tiramisu','tenant_4','restaurant_t4_1','cat_t4_r1_desserts','Tiramisu','Classic Italian coffee dessert',7.99,7.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,15,'user_t4_admin'),
    ('mi_t4_r1_chianti','tenant_4','restaurant_t4_1','cat_t4_r1_drinks','Chianti','Red wine from Tuscany',8.99,8.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t4_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Spicy Wings - Restaurant 1 and 2 (tenant_5)
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    ('mi_t5_r1_buffalo_wings','tenant_5','restaurant_t5_1','cat_t5_r1_wings','Buffalo Wings','Spicy buffalo sauce with celery and ranch',10.99,10.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,15,'user_t5_admin'),
    ('mi_t5_r1_honey_sriracha_wings','tenant_5','restaurant_t5_1','cat_t5_r1_wings','Honey Sriracha Wings','Sweet and spicy honey sriracha glaze',11.99,11.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,15,'user_t5_admin'),
    ('mi_t5_r1_spicy_chicken_sandwich','tenant_5','restaurant_t5_1','cat_t5_r1_sandwiches','Spicy Chicken Sandwich','Grilled chicken with spicy mayo, lettuce, tomato, and bacon',9.99,9.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,12,'user_t5_admin'),
    ('mi_t5_r1_spicy_bbq_burger','tenant_5','restaurant_t5_1','cat_t5_r1_burgers','Spicy BBQ Burger','Beef patty with spicy BBQ sauce, bacon, and cheddar',11.99,11.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,12,'user_t5_admin'),
    ('mi_t5_r1_spicy_fries','tenant_5','restaurant_t5_1','cat_t5_r1_sides','Spicy Fries','Fries with spicy mayo',4.99,4.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t5_admin'),
    ('mi_t5_r1_spicy_margarita','tenant_5','restaurant_t5_1','cat_t5_r1_drinks','Spicy Margarita','Mango margarita with a spicy kick',7.99,7.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,5,'user_t5_admin'),
    ('mi_t5_r1_spicy_chocolate_cake','tenant_5','restaurant_t5_1','cat_t5_r1_desserts','Spicy Chocolate Cake','Chocolate cake with spicy chocolate frosting',5.99,5.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,10,'user_t5_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Healthy Bites - Restaurant 1 and 2 (tenant_6)
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    ('mi_t6_r1_greek_salad','tenant_6','restaurant_t6_1','cat_t6_r1_salads','Greek Salad','Mixed greens with cucumber, tomato, olives, feta, and Greek dressing',8.99,8.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t6_admin'),
    ('mi_t6_r1_caesar_salad','tenant_6','restaurant_t6_1','cat_t6_r1_salads','Caesar Salad','Romaine lettuce with Caesar dressing, croutons, and parmesan',7.99,7.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,8,'user_t6_admin'),
    ('mi_t6_r1_tomato_basil_soup','tenant_6','restaurant_t6_1','cat_t6_r1_soups','Tomato Basil Soup','Creamy tomato soup with fresh basil',5.99,5.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,15,'user_t6_admin'),
    ('mi_t6_r1_turkey_club','tenant_6','restaurant_t6_1','cat_t6_r1_sandwiches','Turkey Club','Turkey, bacon, lettuce, tomato, and avocado on whole grain',9.99,9.99,'USD',TRUE,FALSE,FALSE,FALSE,TRUE,FALSE,10,'user_t6_admin'),
    ('mi_t6_r1_veggie_club','tenant_6','restaurant_t6_1','cat_t6_r1_sandwiches','Veggie Club','Avocado, spinach, tomato, and vegan mayo on whole grain',8.99,8.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t6_admin'),
    ('mi_t6_r1_quinoa_bowl','tenant_6','restaurant_t6_1','cat_t6_r1_bowls','Quinoa Bowl','Quinoa with roasted veggies, chickpeas, and tahini dressing',10.99,10.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,15,'user_t6_admin'),
    ('mi_t6_r1_green_smoothie','tenant_6','restaurant_t6_1','cat_t6_r1_drinks','Green Smoothie','Spinach, banana, and almond milk',4.99,4.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t6_admin'),
    ('mi_t6_r1_fruit_salad','tenant_6','restaurant_t6_1','cat_t6_r1_desserts','Fruit Salad','Fresh mixed fruit with honey-lime dressing',5.99,5.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t6_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Dessert Den - Restaurant 1 and 2 (tenant_7)
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    ('mi_t7_r1_chocolate_cake','tenant_7','restaurant_t7_1','cat_t7_r1_cakes','Chocolate Cake','Rich chocolate cake with chocolate frosting',6.99,6.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,20,'user_t7_admin'),
    ('mi_t7_r1_vanilla_cake','tenant_7','restaurant_t7_1','cat_t7_r1_cakes','Vanilla Cake','Classic vanilla cake with vanilla frosting',6.49,6.49,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,20,'user_t7_admin'),
    ('mi_t7_r1_choc_chip_cookie','tenant_7','restaurant_t7_1','cat_t7_r1_cookies','Chocolate Chip Cookie','Warm chocolate chip cookie',2.99,2.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t7_admin'),
    ('mi_t7_r1_vanilla_ice_cream','tenant_7','restaurant_t7_1','cat_t7_r1_icecream','Vanilla Ice Cream','Classic vanilla ice cream',3.99,3.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t7_admin'),
    ('mi_t7_r1_chocolate_ice_cream','tenant_7','restaurant_t7_1','cat_t7_r1_icecream','Chocolate Ice Cream','Rich chocolate ice cream',4.49,4.49,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,5,'user_t7_admin'),
    ('mi_t7_r1_croissant','tenant_7','restaurant_t7_1','cat_t7_r1_pastries','Croissant','Flaky French pastry',3.99,3.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t7_admin'),
    ('mi_t7_r1_espresso','tenant_7','restaurant_t7_1','cat_t7_r1_coffee','Espresso','Strong Italian espresso',2.99,2.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,3,'user_t7_admin'),
    ('mi_t7_r1_earl_grey','tenant_7','restaurant_t7_1','cat_t7_r1_tea','Earl Grey','Black tea with bergamot',2.49,2.49,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,3,'user_t7_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- Breakfast Club - Restaurant 1 and 2 (tenant_8)
INSERT INTO tenant_menu_items (menu_item_id, tenant_id, restaurant_id, category_id, name, description, base_price, price, currency, is_active, is_featured, is_vegetarian, is_vegan, is_gluten_free, is_dairy_free, preparation_time_minutes, created_by) VALUES
    ('mi_t8_r1_classic_pancakes','tenant_8','restaurant_t8_1','cat_t8_r1_breakfast','Classic Pancakes','Fluffy buttermilk pancakes with maple syrup',6.99,6.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t8_admin'),
    ('mi_t8_r1_french_toast','tenant_8','restaurant_t8_1','cat_t8_r1_breakfast','French Toast','Thick-cut brioche French toast with berries',7.99,7.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,12,'user_t8_admin'),
    ('mi_t8_r1_bacon','tenant_8','restaurant_t8_1','cat_t8_r1_bacon','Bacon','Crispy bacon',3.99,3.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,5,'user_t8_admin'),
    ('mi_t8_r1_over_easy','tenant_8','restaurant_t8_1','cat_t8_r1_eggs','Over Easy','Classic over easy egg',3.99,3.99,'USD',TRUE,FALSE,FALSE,FALSE,FALSE,FALSE,5,'user_t8_admin'),
    ('mi_t8_r1_veggie_omelet','tenant_8','restaurant_t8_1','cat_t8_r1_eggs','Veggie Omelet','Eggs with spinach, mushrooms, and cheese',6.99,6.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,8,'user_t8_admin'),
    ('mi_t8_r1_blueberry_pancakes','tenant_8','restaurant_t8_1','cat_t8_r1_pancakes','Blueberry Pancakes','Fluffy pancakes with fresh blueberries',7.49,7.49,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t8_admin'),
    ('mi_t8_r1_orange_juice','tenant_8','restaurant_t8_1','cat_t8_r1_drinks','Orange Juice','Fresh-squeezed orange juice',3.49,3.49,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,5,'user_t8_admin'),
    ('mi_t8_r1_waffles','tenant_8','restaurant_t8_1','cat_t8_r1_desserts','Waffles','Belgian-style waffles with maple syrup',6.99,6.99,'USD',TRUE,FALSE,TRUE,TRUE,TRUE,FALSE,10,'user_t8_admin')
ON CONFLICT (menu_item_id) DO NOTHING;

-- ============================================
-- Re-enable row-level security
-- ============================================
SET LOCAL row_security = on;