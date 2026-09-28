-- +migrate Down

ALTER TABLE menu_item_component
    DROP CONSTRAINT IF EXISTS menu_item_component_default_size_fk;

DROP TABLE IF EXISTS menu_item_component_size_option;
DROP TABLE IF EXISTS menu_item_component;
DROP TABLE IF EXISTS menu_item;
DROP TABLE IF EXISTS menu_category;
DROP INDEX IF EXISTS menu_weekly_published_start_date_uq;
DROP TABLE IF EXISTS menu;
