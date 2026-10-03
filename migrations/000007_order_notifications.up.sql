-- +migrate Up
-- REQNOTIF001–REQNOTIF005: tenant cook admin phone, transactional notification outbox, local SMS dev sink.
-- Keep in sync with scripts/init/01_food_catalog.sql (tenant column), scripts/init/04_notifications.sql, tests/db/schema.sql.

ALTER TABLE tenant ADD COLUMN IF NOT EXISTS cook_admin_phone VARCHAR(50);

CREATE TABLE IF NOT EXISTS notification_outbox (
    id               VARCHAR(50) NOT NULL,
    tenant_id        VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    order_id         VARCHAR(50) NOT NULL,
    event_type       VARCHAR(50) NOT NULL
        CONSTRAINT notification_outbox_event_type_check CHECK (event_type <> ''),
    channel          VARCHAR(20) NOT NULL DEFAULT 'sms'
        CONSTRAINT notification_outbox_channel_check CHECK (channel IN ('sms')),
    recipient_phone  VARCHAR(50) NOT NULL
        CONSTRAINT notification_outbox_recipient_check CHECK (recipient_phone <> ''),
    body             TEXT NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending'
        CONSTRAINT notification_outbox_status_check
        CHECK (status IN ('pending', 'processing', 'delivered', 'dead')),
    attempts         INT NOT NULL DEFAULT 0
        CONSTRAINT notification_outbox_attempts_check CHECK (attempts >= 0),
    next_attempt_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    last_error       TEXT,
    provider_ref     TEXT,
    created_at       BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at       BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT notification_outbox_order_fk
        FOREIGN KEY (tenant_id, order_id)
        REFERENCES customer_order (tenant_id, id)
        ON DELETE CASCADE,
    -- Idempotent enqueue: one notification per (order, event) in v1.
    CONSTRAINT notification_outbox_order_event_unique
        UNIQUE (tenant_id, order_id, event_type)
);

CREATE INDEX IF NOT EXISTS idx_notification_outbox_order
    ON notification_outbox(tenant_id, order_id);
CREATE INDEX IF NOT EXISTS idx_notification_outbox_claim
    ON notification_outbox(status, next_attempt_at)
    WHERE status IN ('pending', 'processing');

-- Local SmsProvider sink (dev/tests only; no external SMS vendor).
CREATE TABLE IF NOT EXISTS sms_dev_sink (
    id          VARCHAR(50) NOT NULL,
    tenant_id   VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    outbox_id   VARCHAR(50) NOT NULL,
    phone       VARCHAR(50) NOT NULL,
    body        TEXT NOT NULL,
    created_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT sms_dev_sink_outbox_fk
        FOREIGN KEY (tenant_id, outbox_id)
        REFERENCES notification_outbox (tenant_id, id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sms_dev_sink_outbox ON sms_dev_sink(tenant_id, outbox_id);
