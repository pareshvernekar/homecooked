-- Order intake schema (keep in sync with migrations/000005_order_intake.up.sql and tests/db/schema.sql)

CREATE TABLE IF NOT EXISTS customer_order (
    id                  VARCHAR(50) NOT NULL,
    tenant_id           VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    menu_id             VARCHAR(50) NOT NULL,
    customer_name       VARCHAR(200) NOT NULL,
    customer_phone      VARCHAR(50) NOT NULL,
    received_at         BIGINT NOT NULL,
    expected_at         BIGINT NOT NULL,
    pickedup_at         BIGINT,
    status              VARCHAR(20) NOT NULL DEFAULT 'RECEIVED'
        CHECK (status IN ('RECEIVED', 'IN_PROGRESS', 'COMPLETE', 'PICKEDUP')),
    customization_text  TEXT,
    total_override      NUMERIC(10,2)
        CHECK (total_override IS NULL OR total_override >= 0),
    frozen_total        NUMERIC(10,2),
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at          BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT customer_order_menu_fk
        FOREIGN KEY (tenant_id, menu_id)
        REFERENCES menu (tenant_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_customer_order_tenant ON customer_order(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customer_order_tenant_status
    ON customer_order(tenant_id, status) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_customer_order_menu ON customer_order(tenant_id, menu_id);

CREATE TABLE IF NOT EXISTS order_item (
    id                       VARCHAR(50) NOT NULL,
    tenant_id                VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    order_id                 VARCHAR(50) NOT NULL,
    menu_item_id             VARCHAR(50) NOT NULL,
    quantity                 INT NOT NULL CHECK (quantity > 0),
    customization_text       TEXT,
    unit_price_override      NUMERIC(10,2)
        CHECK (unit_price_override IS NULL OR unit_price_override >= 0),
    frozen_unit_price        NUMERIC(10,2),
    frozen_extended_amount   NUMERIC(10,2),
    is_active                BOOLEAN NOT NULL DEFAULT TRUE,
    created_at               BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at               BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT order_item_order_fk
        FOREIGN KEY (tenant_id, order_id)
        REFERENCES customer_order (tenant_id, id)
        ON DELETE CASCADE,
    CONSTRAINT order_item_menu_item_fk
        FOREIGN KEY (tenant_id, menu_item_id)
        REFERENCES menu_item (tenant_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_order_item_order ON order_item(tenant_id, order_id);

CREATE TABLE IF NOT EXISTS order_item_component_selection (
    id                      VARCHAR(50) NOT NULL,
    tenant_id               VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    order_item_id           VARCHAR(50) NOT NULL,
    menu_item_component_id  VARCHAR(50) NOT NULL,
    size_option_id          VARCHAR(50) NOT NULL,
    created_at              BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at              BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT order_item_component_selection_item_fk
        FOREIGN KEY (tenant_id, order_item_id)
        REFERENCES order_item (tenant_id, id)
        ON DELETE CASCADE,
    CONSTRAINT order_item_component_selection_component_fk
        FOREIGN KEY (tenant_id, menu_item_component_id)
        REFERENCES menu_item_component (tenant_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT order_item_component_selection_option_fk
        FOREIGN KEY (tenant_id, size_option_id)
        REFERENCES menu_item_component_size_option (tenant_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT order_item_component_selection_unique
        UNIQUE (tenant_id, order_item_id, menu_item_component_id)
);

CREATE INDEX IF NOT EXISTS idx_order_item_component_selection_item
    ON order_item_component_selection(tenant_id, order_item_id);

CREATE TABLE IF NOT EXISTS order_payment (
    id              VARCHAR(50) NOT NULL,
    tenant_id       VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    order_id        VARCHAR(50) NOT NULL,
    mode            VARCHAR(20) NOT NULL
        CHECK (mode IN ('cash', 'credit', 'paypal', 'zelle', 'venmo')),
    amount          NUMERIC(10,2) NOT NULL CHECK (amount > 0),
    reference_text  TEXT,
    created_at      BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at      BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT order_payment_order_fk
        FOREIGN KEY (tenant_id, order_id)
        REFERENCES customer_order (tenant_id, id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_order_payment_order ON order_payment(tenant_id, order_id);
