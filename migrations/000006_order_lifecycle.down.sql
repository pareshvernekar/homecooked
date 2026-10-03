-- +migrate Down

UPDATE customer_order SET status = 'COMPLETE' WHERE status = 'READY';
UPDATE customer_order SET status = 'RECEIVED' WHERE status IN ('ACCEPTED', 'DECLINED');

ALTER TABLE customer_order DROP CONSTRAINT IF EXISTS customer_order_status_check;

ALTER TABLE customer_order
    ADD CONSTRAINT customer_order_status_check
        CHECK (status IN ('RECEIVED', 'IN_PROGRESS', 'COMPLETE', 'PICKEDUP'));

ALTER TABLE customer_order DROP COLUMN IF EXISTS refuse_reason;
