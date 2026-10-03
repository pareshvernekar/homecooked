-- +migrate Down

DROP TABLE IF EXISTS order_payment;
DROP TABLE IF EXISTS order_item_component_selection;
DROP TABLE IF EXISTS order_item;
DROP INDEX IF EXISTS idx_customer_order_menu;
DROP INDEX IF EXISTS idx_customer_order_tenant_status;
DROP INDEX IF EXISTS idx_customer_order_tenant;
DROP TABLE IF EXISTS customer_order;
