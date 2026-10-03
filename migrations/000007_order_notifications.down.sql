-- +migrate Down

DROP INDEX IF EXISTS idx_sms_dev_sink_outbox;
DROP TABLE IF EXISTS sms_dev_sink;
DROP INDEX IF EXISTS idx_notification_outbox_claim;
DROP INDEX IF EXISTS idx_notification_outbox_order;
DROP TABLE IF EXISTS notification_outbox;
ALTER TABLE tenant DROP COLUMN IF EXISTS cook_admin_phone;
