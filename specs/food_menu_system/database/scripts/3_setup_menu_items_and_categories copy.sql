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
-- Burger Haven
INSERT INTO menu_categories (tenant_id, restaurant_id, name, description, is_active) VALUES
    (1, 1, 'Burgers', 'Classic beef burgers with various toppings', TRUE),
    (1, 1, 'Sides', 'Sides to complement your burger', TRUE),
    (1, 1, 'Drinks', 'Refreshing beverages', TRUE),
    (1, 1, 'Desserts', 'Sweet treats to end your meal', TRUE),
    (1, 1, 'Kids Menu', 'Kid-friendly options', TRUE),
    
    (1, 2, 'Burgers', 'Classic beef burgers with various toppings', TRUE),
    (1, 2, 'Sides', 'Sides to complement your burger', TRUE),
    (1, 2, 'Drinks', 'Refreshing beverages', TRUE),
    (1, 2, 'Desserts', 'Sweet treats to end your meal', TRUE),
    (1, 2, 'Salads', 'Fresh salads', TRUE);

-- Sushi Palace
INSERT INTO menu_categories (tenant_id, restaurant_id, name, description, is_active) VALUES
    (2, 1, 'Sushi', 'Freshly prepared sushi rolls', TRUE),
    (2, 1, 'Sashimi', 'Thinly sliced raw fish', TRUE),
    (2, 1, 'Tempura', 'Lightly battered and fried dishes', TRUE),
    (2, 1, 'Ramen', 'Rich noodle soups', TRUE),
    (2, 1, 'Appetizers', 'Small dishes to start your meal', TRUE),
    (2, 1, 'Drinks', 'Traditional Japanese beverages', TRUE),
    
    (2, 2, 'Sushi', 'Freshly prepared sushi rolls', TRUE),
    (2, 2, 'Sashimi', 'Thinly sliced raw fish', TRUE),
    (2, 2, 'Udon', 'Thick wheat noodles', TRUE),
    (2, 2, 'Drinks', 'Traditional Japanese beverages', TRUE),
    (2, 2, 'Desserts', 'Japanese sweets', TRUE);

-- Veggie Delight
INSERT INTO menu_categories (tenant_id, restaurant_id, name, description, is_active) VALUES
    (3, 1, 'Vegetarian', '100% plant-based meals', TRUE),
    (3, 1, 'Vegan', 'Completely plant-based options', TRUE),
    (3, 1, 'Salads', 'Fresh and healthy salads', TRUE),
    (3, 1, 'Soups', 'Warm and comforting soups', TRUE),
    (3, 1, 'Sandwiches', 'Freshly made sandwiches', TRUE),
    (3, 1, 'Desserts', 'Healthy sweet treats', TRUE),
    
    (3, 2, 'Vegetarian', '100% plant-based meals', TRUE),
    (3, 2, 'Vegan', 'Completely plant-based options', TRUE),
    (3, 2, 'Burgers', 'Veggie burgers', TRUE),
    (3, 2, 'Drinks', 'Healthy beverage options', TRUE),
    (3, 2, 'Breakfast', 'Morning meals', TRUE);

-- Italian Eats
INSERT INTO menu_categories (tenant_id, restaurant_id, name, description, is_active) VALUES
    (4, 1, 'Pasta', 'Classic Italian pasta dishes', TRUE),
    (4, 1, 'Pizza', 'Authentic Italian pizza', TRUE),
    (4, 1, 'Antipasti', 'Italian starters', TRUE),
    (4, 1, 'Risotto', 'Creamy rice dishes', TRUE),
    (4, 1, 'Desserts', 'Italian sweets', TRUE),
    (4, 1, 'Drinks', 'Italian wines and beverages', TRUE),
    
    (4, 2, 'Pasta', 'Classic Italian pasta dishes', TRUE),
    (4, 2, 'Pizza', 'Authentic Italian pizza', TRUE),
    (4, 2, 'Appetizers', 'Italian starters', TRUE),
    (4, 2, 'Drinks', 'Italian wines and beverages', TRUE),
    (4, 2, 'Desserts', 'Italian sweets', TRUE);

-- Spicy Wings
INSERT INTO menu_categories (tenant_id, restaurant_id, name, description, is_active) VALUES
    (5, 1, 'Wings', 'Spicy and crispy wings', TRUE),
    (5, 1, 'Sandwiches', 'Meat-filled sandwiches', TRUE),
    (5, 1, 'Burgers', 'Spicy beef burgers', TRUE),
    (5, 1, 'Sides', 'Spicy and tangy sides', TRUE),
    (5, 1, 'Drinks', 'Cold beverages', TRUE),
    (5, 1, 'Desserts', 'Sweet treats', TRUE),
    
    (5, 2, 'Wings', 'Spicy and crispy wings', TRUE),
    (5, 2, 'Sandwiches', 'Meat-filled sandwiches', TRUE),
    (5, 2, 'Burgers', 'Spicy beef burgers', TRUE),
    (5, 2, 'Drinks', 'Cold beverages', TRUE),
    (5, 2, 'Desserts', 'Sweet treats', TRUE);

-- Healthy Bites
INSERT INTO menu_categories (tenant_id, restaurant_id, name, description, is_active) VALUES
    (6, 1, 'Salads', 'Fresh and healthy salads', TRUE),
    (6, 1, 'Soups', 'Light and nutritious soups', TRUE),
    (6, 1, 'Sandwiches', 'Whole grain sandwiches', TRUE),
    (6, 1, 'Bowls', 'Nutritious grain bowls', TRUE),
    (6, 1, 'Drinks', 'Healthy beverage options', TRUE),
    (6, 1, 'Desserts', 'Low-sugar desserts', TRUE),
    
    (6, 2, 'Salads', 'Fresh and healthy salads', TRUE),
    (6, 2, 'Soups', 'Light and nutritious soups', TRUE),
    (6, 2, 'Sandwiches', 'Whole grain sandwiches', TRUE),
    (6, 2, 'Drinks', 'Healthy beverage options', TRUE),
    (6, 2, 'Breakfast', 'Healthy morning meals', TRUE);

-- Dessert Den
INSERT INTO menu_categories (tenant_id, restaurant_id, name, description, is_active) VALUES
    (7, 1, 'Cakes', 'Beautifully decorated cakes', TRUE),
    (7, 1, 'Cookies', 'Freshly baked cookies', TRUE),
    (7, 1, 'Ice Cream', 'Premium ice cream', TRUE),
    (7, 1, 'Pastries', 'Delicious pastries', TRUE),
    (7, 1, 'Coffee', 'Specialty coffee drinks', TRUE),
    (7, 1, 'Tea', 'Fine teas', TRUE),
    
    (7, 2, 'Cakes', 'Beautifully decorated cakes', TRUE),
    (7, 2, 'Cookies', 'Freshly baked cookies', TRUE),
    (7, 2, 'Ice Cream', 'Premium ice cream', TRUE),
    (7, 2, 'Pastries', 'Delicious pastries', TRUE),
    (7, 2, 'Drinks', 'Refreshing beverages', TRUE);

-- Breakfast Club
INSERT INTO menu_categories (tenant_id, restaurant_id, name, description, is_active) VALUES
    (8, 1, 'Breakfast', 'Classic breakfast items', TRUE),
    (8, 1, 'Bacon', 'Various bacon dishes', TRUE),
    (8, 1, 'Eggs', 'Egg-based dishes', TRUE),
    (8, 1, 'Pancakes', 'Fluffy pancakes', TRUE),
    (8, 1, 'Drinks', 'Morning beverages', TRUE),
    (8, 1, 'Desserts', 'Sweet breakfast treats', TRUE),
    
    (8, 2, 'Breakfast', 'Classic breakfast items', TRUE),
    (8, 2, 'Bacon', 'Various bacon dishes', TRUE),
    (8, 2, 'Eggs', 'Egg-based dishes', TRUE),
    (8, 2, 'Pancakes', 'Fluffy pancakes', TRUE),
    (8, 2, 'Drinks', 'Morning beverages', TRUE),
    (8, 2, 'Desserts', 'Sweet breakfast treats', TRUE);

-- ============================================
-- Create sample menu items
-- ============================================
-- Burger Haven - Restaurant 1
INSERT INTO menu_items (tenant_id, restaurant_id, category_id, name, description, price, is_active, is_vegetarian, is_vegan, is_gluten_free, is_spicy, image_url, preparation_time_minutes) VALUES
    -- Burgers
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Burgers'), 'Classic Cheeseburger', 'Juicy beef patty with melted cheddar, lettuce, tomato, onion, pickles, and special sauce on a brioche bun', 9.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://burgerhaven.example.com/menu/classic-cheeseburger.jpg', 10),
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Burgers'), 'Spicy Chicken Burger', 'Grilled chicken patty with spicy mayo, pepper jack cheese, bacon, and sriracha aioli', 10.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://burgerhaven.example.com/menu/spicy-chicken-burger.jpg', 12),
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Burgers'), 'Veggie Burger', 'Grilled portobello mushroom patty with avocado, spinach, and vegan cheese', 8.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/veggie-burger.jpg', 10),
    
    -- Sides
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Sides'), 'Crispy Fries', 'Thick-cut fries with sea salt', 3.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/fries.jpg', 8),
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Sides'), 'Sweet Potato Fries', 'Sweet potato fries with garlic parmesan', 4.49, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/sweet-potato-fries.jpg', 9),
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Sides'), 'Onion Rings', 'Crispy onion rings with ranch', 4.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/onion-rings.jpg', 8),
    
    -- Drinks
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Drinks'), 'Classic Coke', 'Original Coca-Cola', 2.49, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/coke.jpg', 2),
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Drinks'), 'Iced Tea', 'Sweet iced tea', 1.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/iced-tea.jpg', 3),
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Drinks'), 'Lemonade', 'Fresh-squeezed lemonade', 2.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/lemonade.jpg', 4),
    
    -- Desserts
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Desserts'), 'Chocolate Chip Cookie', 'Warm chocolate chip cookie', 2.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/chocolate-chip-cookie.jpg', 5),
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Desserts'), 'Brownies', 'Fudgy chocolate brownies', 3.49, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/brownies.jpg', 6),
    
    -- Kids Menu
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Kids Menu'), 'Mini Cheeseburger', 'Smaller version of our classic cheeseburger', 6.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://burgerhaven.example.com/menu/mini-cheeseburger.jpg', 8),
    (1, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 1 AND name = 'Kids Menu'), 'Chicken Tenders', 'Crispy chicken tenders with honey mustard', 5.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://burgerhaven.example.com/menu/chicken-tenders.jpg', 10),
    
    -- Burger Haven - Restaurant 2
    (1, 2, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 2 AND name = 'Burgers'), 'Gourmet Burger', 'Premium beef patty with truffle aioli, caramelized onions, and blue cheese', 14.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://burgerhaven.example.com/menu/gourmet-burger.jpg', 12),
    (1, 2, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 2 AND name = 'Salads'), 'Caesar Salad', 'Romaine lettuce with Caesar dressing, croutons, and parmesan', 7.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://burgerhaven.example.com/menu/caesar-salad.jpg', 8),
    (1, 2, (SELECT category_id FROM menu_categories WHERE tenant_id = 1 AND restaurant_id = 2 AND name = 'Salads'), 'Greek Salad', 'Mixed greens with cucumber, tomato, olives, feta, and Greek dressing', 8.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://burgerhaven.example.com/menu/greek-salad.jpg', 8);

-- Sushi Palace - Restaurant 1
INSERT INTO menu_items (tenant_id, restaurant_id, category_id, name, description, price, is_active, is_vegetarian, is_vegan, is_gluten_free, is_spicy, image_url, preparation_time_minutes) VALUES
    -- Sushi
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Sushi'), 'Dragon Roll', 'Avocado, cucumber, and tempura crunch wrapped in nori with spicy mayo', 12.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://sushipalace.example.com/menu/dragon-roll.jpg', 15),
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Sushi'), 'Rainbow Roll', 'Assorted fish and avocado with eel sauce', 14.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://sushipalace.example.com/menu/rainbow-roll.jpg', 18),
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Sushi'), 'Veggie Roll', 'Avocado, cucumber, carrot, and pickled radish with spicy mayo', 11.99, TRUE, TRUE, TRUE, TRUE, TRUE, 'https://sushipalace.example.com/menu/veggie-roll.jpg', 15),
    
    -- Sashimi
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Sashimi'), 'Tuna Sashimi', 'Fresh sashimi-grade tuna', 18.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://sushipalace.example.com/menu/tuna-sashimi.jpg', 5),
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Sashimi'), 'Salmon Sashimi', 'Fresh sashimi-grade salmon', 16.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://sushipalace.example.com/menu/salmon-sashimi.jpg', 5),
    
    -- Tempura
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Tempura'), 'Tempura Shrimp', 'Lightly battered shrimp with dipping sauce', 14.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://sushipalace.example.com/menu/tempura-shrimp.jpg', 12),
    
    -- Ramen
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Ramen'), 'Spicy Miso Ramen', 'Rich miso broth with chashu pork, soft-boiled egg, and nori', 12.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://sushipalace.example.com/menu/spicy-miso-ramen.jpg', 20),
    
    -- Appetizers
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Appetizers'), 'Edamame', 'Steamed soybeans with sea salt', 6.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://sushipalace.example.com/menu/edamame.jpg', 8),
    
    -- Drinks
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Drinks'), 'Green Tea', 'Freshly brewed green tea', 2.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://sushipalace.example.com/menu/green-tea.jpg', 3),
    (2, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 2 AND restaurant_id = 1 AND name = 'Drinks'), 'Sake', 'Premium Japanese sake', 5.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://sushipalace.example.com/menu/sake.jpg', 5);

-- Veggie Delight - Restaurant 1
INSERT INTO menu_items (tenant_id, restaurant_id, category_id, name, description, price, is_active, is_vegetarian, is_vegan, is_gluten_free, is_spicy, image_url, preparation_time_minutes) VALUES
    -- Vegetarian
    (3, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Vegetarian'), 'Portobello Mushroom Burger', 'Grilled portobello with avocado, spinach, and vegan cheese', 10.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://veggiedelight.example.com/menu/portobello-burger.jpg', 12),
    (3, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Vegetarian'), 'Black Bean Burger', 'Spiced black bean patty with roasted peppers and vegan mayo', 9.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://veggiedelight.example.com/menu/black-bean-burger.jpg', 10),
    
    -- Vegan
    (3, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Vegan'), 'Beyond Burger', 'Plant-based beef patty with all the fixings', 10.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://veggiedelight.example.com/menu/beyond-burger.jpg', 10),
    
    -- Salads
    (3, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Salads'), 'Kale Caesar', 'Kale with Caesar dressing, croutons, and vegan parmesan', 8.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://veggiedelight.example.com/menu/kale-caesar.jpg', 8),
    (3, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Salads'), 'Quinoa Salad', 'Quinoa with roasted veggies and lemon vinaigrette', 9.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://veggiedelight.example.com/menu/quinoa-salad.jpg', 10),
    
    -- Soups
    (3, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Soups'), 'Minestrone', 'Vegetable soup with beans and pasta', 5.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://veggiedelight.example.com/menu/minestrone.jpg', 15),
    
    -- Sandwiches
    (3, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Sandwiches'), 'Veggie Sandwich', 'Whole grain bread with hummus, roasted veggies, and avocado', 8.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://veggiedelight.example.com/menu/veggie-sandwich.jpg', 10),
    
    -- Desserts
    (3, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 3 AND restaurant_id = 1 AND name = 'Desserts'), 'Vegan Chocolate Cake', 'Rich chocolate cake with vegan frosting', 5.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://veggiedelight.example.com/menu/vegan-chocolate-cake.jpg', 10);

-- Italian Eats - Restaurant 1
INSERT INTO menu_items (tenant_id, restaurant_id, category_id, name, description, price, is_active, is_vegetarian, is_vegan, is_gluten_free, is_spicy, image_url, preparation_time_minutes) VALUES
    -- Pasta
    (4, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Pasta'), 'Spaghetti Carbonara', 'Spaghetti with egg, cheese, pancetta, and black pepper', 14.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://italianeats.example.com/menu/spaghetti-carbonara.jpg', 15),
    (4, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Pasta'), 'Penne Arrabbiata', 'Spicy tomato sauce with penne pasta', 12.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://italianeats.example.com/menu/penne-arrabbiata.jpg', 15),
    
    -- Pizza
    (4, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Pizza'), 'Margherita Pizza', 'Classic tomato sauce, mozzarella, and basil', 12.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://italianeats.example.com/menu/margherita-pizza.jpg', 20),
    (4, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Pizza'), 'Veggie Pizza', 'Tomato sauce, mozzarella, bell peppers, onions, olives, and mushrooms', 13.99, TRUE, TRUE, FALSE, FALSE, FALSE, 'https://italianeats.example.com/menu/veggie-pizza.jpg', 20),
    
    -- Antipasti
    (4, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Antipasti'), 'Bruschetta', 'Toasted bread with tomato, garlic, basil, and balsamic glaze', 6.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://italianeats.example.com/menu/bruschetta.jpg', 8),
    
    -- Risotto
    (4, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Risotto'), 'Risotto ai Funghi', 'Creamy risotto with wild mushrooms', 13.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://italianeats.example.com/menu/risotto-funghi.jpg', 20),
    
    -- Desserts
    (4, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Desserts'), 'Tiramisu', 'Classic Italian coffee dessert', 7.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://italianeats.example.com/menu/tiramisu.jpg', 15),
    
    -- Drinks
    (4, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 4 AND restaurant_id = 1 AND name = 'Drinks'), 'Chianti', 'Red wine from Tuscany', 8.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://italianeats.example.com/menu/chianti.jpg', 5);

-- Spicy Wings - Restaurant 1
INSERT INTO menu_items (tenant_id, restaurant_id, category_id, name, description, price, is_active, is_vegetarian, is_vegan, is_gluten_free, is_spicy, image_url, preparation_time_minutes) VALUES
    -- Wings
    (5, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 5 AND restaurant_id = 1 AND name = 'Wings'), 'Buffalo Wings', 'Spicy buffalo sauce with celery and ranch', 10.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://spicywings.example.com/menu/buffalo-wings.jpg', 15),
    (5, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 5 AND restaurant_id = 1 AND name = 'Wings'), 'Honey Sriracha Wings', 'Sweet and spicy honey sriracha glaze', 11.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://spicywings.example.com/menu/honey-sriracha-wings.jpg', 15),
    
    -- Sandwiches
    (5, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 5 AND restaurant_id = 1 AND name = 'Sandwiches'), 'Spicy Chicken Sandwich', 'Grilled chicken with spicy mayo, lettuce, tomato, and bacon', 9.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://spicywings.example.com/menu/spicy-chicken-sandwich.jpg', 12),
    
    -- Burgers
    (5, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 5 AND restaurant_id = 1 AND name = 'Burgers'), 'Spicy BBQ Burger', 'Beef patty with spicy BBQ sauce, bacon, and cheddar', 11.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://spicywings.example.com/menu/spicy-bbq-burger.jpg', 12),
    
    -- Sides
    (5, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 5 AND restaurant_id = 1 AND name = 'Sides'), 'Spicy Fries', 'Fries with spicy mayo', 4.99, TRUE, TRUE, TRUE, TRUE, TRUE, 'https://spicywings.example.com/menu/spicy-fries.jpg', 8),
    
    -- Drinks
    (5, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 5 AND restaurant_id = 1 AND name = 'Drinks'), 'Spicy Margarita', 'Mango margarita with a spicy kick', 7.99, TRUE, TRUE, TRUE, TRUE, TRUE, 'https://spicywings.example.com/menu/spicy-margarita.jpg', 5),
    
    -- Desserts
    (5, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 5 AND restaurant_id = 1 AND name = 'Desserts'), 'Spicy Chocolate Cake', 'Chocolate cake with spicy chocolate frosting', 5.99, TRUE, FALSE, FALSE, FALSE, TRUE, 'https://spicywings.example.com/menu/spicy-chocolate-cake.jpg', 10);

-- Healthy Bites - Restaurant 1
INSERT INTO menu_items (tenant_id, restaurant_id, category_id, name, description, price, is_active, is_vegetarian, is_vegan, is_gluten_free, is_spicy, image_url, preparation_time_minutes) VALUES
    -- Salads
    (6, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 6 AND restaurant_id = 1 AND name = 'Salads'), 'Greek Salad', 'Mixed greens with cucumber, tomato, olives, feta, and Greek dressing', 8.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://healthybites.example.com/menu/greek-salad.jpg', 8),
    (6, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 6 AND restaurant_id = 1 AND name = 'Salads'), 'Caesar Salad', 'Romaine lettuce with Caesar dressing, croutons, and parmesan', 7.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://healthybites.example.com/menu/caesar-salad.jpg', 8),
    
    -- Soups
    (6, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 6 AND restaurant_id = 1 AND name = 'Soups'), 'Tomato Basil Soup', 'Creamy tomato soup with fresh basil', 5.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://healthybites.example.com/menu/tomato-basil-soup.jpg', 15),
    
    -- Sandwiches
    (6, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 6 AND restaurant_id = 1 AND name = 'Sandwiches'), 'Turkey Club', 'Turkey, bacon, lettuce, tomato, and avocado on whole grain', 9.99, TRUE, FALSE, FALSE, TRUE, FALSE, 'https://healthybites.example.com/menu/turkey-club.jpg', 10),
    (6, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 6 AND restaurant_id = 1 AND name = 'Sandwiches'), 'Veggie Club', 'Avocado, spinach, tomato, and vegan mayo on whole grain', 8.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://healthybites.example.com/menu/veggie-club.jpg', 10),
    
    -- Bowls
    (6, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 6 AND restaurant_id = 1 AND name = 'Bowls'), 'Quinoa Bowl', 'Quinoa with roasted veggies, chickpeas, and tahini dressing', 10.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://healthybites.example.com/menu/quinoa-bowl.jpg', 15),
    
    -- Drinks
    (6, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 6 AND restaurant_id = 1 AND name = 'Drinks'), 'Green Smoothie', 'Spinach, banana, and almond milk', 4.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://healthybites.example.com/menu/green-smoothie.jpg', 5),
    
    -- Desserts
    (6, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 6 AND restaurant_id = 1 AND name = 'Desserts'), 'Fruit Salad', 'Fresh mixed fruit with honey-lime dressing', 5.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://healthybites.example.com/menu/fruit-salad.jpg', 5);

-- Dessert Den - Restaurant 1
INSERT INTO menu_items (tenant_id, restaurant_id, category_id, name, description, price, is_active, is_vegetarian, is_vegan, is_gluten_free, is_spicy, image_url, preparation_time_minutes) VALUES
    -- Cakes
    (7, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 7 AND restaurant_id = 1 AND name = 'Cakes'), 'Chocolate Cake', 'Rich chocolate cake with chocolate frosting', 6.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://dessertden.example.com/menu/chocolate-cake.jpg', 20),
    (7, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 7 AND restaurant_id = 1 AND name = 'Cakes'), 'Vanilla Cake', 'Classic vanilla cake with vanilla frosting', 6.49, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://dessertden.example.com/menu/vanilla-cake.jpg', 20),
    
    -- Cookies
    (7, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 7 AND restaurant_id = 1 AND name = 'Cookies'), 'Chocolate Chip Cookie', 'Warm chocolate chip cookie', 2.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://dessertden.example.com/menu/chocolate-chip-cookie.jpg', 5),
    
    -- Ice Cream
    (7, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 7 AND restaurant_id = 1 AND name = 'Ice Cream'), 'Vanilla Ice Cream', 'Classic vanilla ice cream', 3.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://dessertden.example.com/menu/vanilla-ice-cream.jpg', 5),
    (7, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 7 AND restaurant_id = 1 AND name = 'Ice Cream'), 'Chocolate Ice Cream', 'Rich chocolate ice cream', 4.49, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://dessertden.example.com/menu/chocolate-ice-cream.jpg', 5),
    
    -- Pastries
    (7, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 7 AND restaurant_id = 1 AND name = 'Pastries'), 'Croissant', 'Flaky French pastry', 3.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://dessertden.example.com/menu/croissant.jpg', 10),
    
    -- Coffee
    (7, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 7 AND restaurant_id = 1 AND name = 'Coffee'), 'Espresso', 'Strong Italian espresso', 2.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://dessertden.example.com/menu/espresso.jpg', 3),
    
    -- Tea
    (7, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 7 AND restaurant_id = 1 AND name = 'Tea'), 'Earl Grey', 'Black tea with bergamot', 2.49, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://dessertden.example.com/menu/earl-grey.jpg', 3);

-- Breakfast Club - Restaurant 1
INSERT INTO menu_items (tenant_id, restaurant_id, category_id, name, description, price, is_active, is_vegetarian, is_vegan, is_gluten_free, is_spicy, image_url, preparation_time_minutes) VALUES
    -- Breakfast
    (8, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 8 AND restaurant_id = 1 AND name = 'Breakfast'), 'Classic Pancakes', 'Fluffy buttermilk pancakes with maple syrup', 6.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://breakfastclub.example.com/menu/pancakes.jpg', 10),
    (8, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 8 AND restaurant_id = 1 AND name = 'Breakfast'), 'French Toast', 'Thick-cut brioche French toast with berries', 7.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://breakfastclub.example.com/menu/french-toast.jpg', 12),
    
    -- Bacon
    (8, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 8 AND restaurant_id = 1 AND name = 'Bacon'), 'Bacon', 'Crispy bacon', 3.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://breakfastclub.example.com/menu/bacon.jpg', 5),
    
    -- Eggs
    (8, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 8 AND restaurant_id = 1 AND name = 'Eggs'), 'Over Easy', 'Classic over easy egg', 3.99, TRUE, FALSE, FALSE, FALSE, FALSE, 'https://breakfastclub.example.com/menu/over-easy.jpg', 5),
    (8, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 8 AND restaurant_id = 1 AND name = 'Eggs'), 'Veggie Omelet', 'Eggs with spinach, mushrooms, and cheese', 6.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://breakfastclub.example.com/menu/veggie-omelet.jpg', 8),
    
    -- Pancakes
    (8, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 8 AND restaurant_id = 1 AND name = 'Pancakes'), 'Blueberry Pancakes', 'Fluffy pancakes with fresh blueberries', 7.49, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://breakfastclub.example.com/menu/blueberry-pancakes.jpg', 10),
    
    -- Drinks
    (8, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 8 AND restaurant_id = 1 AND name = 'Drinks'), 'Orange Juice', 'Fresh-squeezed orange juice', 3.49, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://breakfastclub.example.com/menu/orange-juice.jpg', 5),
    
    -- Desserts
    (8, 1, (SELECT category_id FROM menu_categories WHERE tenant_id = 8 AND restaurant_id = 1 AND name = 'Desserts'), 'Waffles', 'Belgian-style waffles with maple syrup', 6.99, TRUE, TRUE, TRUE, TRUE, FALSE, 'https://breakfastclub.example.com/menu/waffles.jpg', 10);

-- ============================================
-- Re-enable row-level security
-- ============================================
SET LOCAL row_security = on;