-- +migrate Up
-- REQFOOD001: food items are catalog definitions without price.

ALTER TABLE food_item DROP COLUMN IF EXISTS price;
