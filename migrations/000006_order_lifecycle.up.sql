-- +migrate Up
-- REQLIFE001–REQLIFE005, REQORDER004: lifecycle statuses, refuse_reason, COMPLETE→READY.

UPDATE customer_order SET status = 'READY' WHERE status = 'COMPLETE';

ALTER TABLE customer_order DROP CONSTRAINT IF EXISTS customer_order_status_check;

ALTER TABLE customer_order
    ADD CONSTRAINT customer_order_status_check
        CHECK (status IN ('RECEIVED', 'ACCEPTED', 'DECLINED', 'IN_PROGRESS', 'READY', 'PICKEDUP'));

ALTER TABLE customer_order
    ADD COLUMN IF NOT EXISTS refuse_reason TEXT;
