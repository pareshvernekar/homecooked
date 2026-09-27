-- +migrate Up
-- Composite tenant keys for food catalog tables.

CREATE TABLE IF NOT EXISTS tenant (
    id          VARCHAR(50) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT
);

-- Recreate catalog tables with composite PKs / FKs.
-- Local/dev: prefer `make db-reset` which re-applies scripts/init.

DROP TABLE IF EXISTS food_item;
DROP TABLE IF EXISTS food_category;

CREATE TABLE food_category (
    id          VARCHAR(50) NOT NULL,
    tenant_id   VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT food_category_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE TABLE food_item (
    id                  VARCHAR(50) NOT NULL,
    tenant_id           VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    name                VARCHAR(100) NOT NULL,
    description         TEXT,
    price               NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    availability_status VARCHAR(50) NOT NULL
        CHECK (availability_status IN ('available', 'low_stock', 'unavailable')),
    category_id         VARCHAR(50),
    image_url           TEXT,
    avoidance           TEXT,
    is_vegetarian       BOOLEAN NOT NULL DEFAULT FALSE,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at          BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    delivered_at        BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT food_item_category_fk
        FOREIGN KEY (tenant_id, category_id)
        REFERENCES food_category (tenant_id, id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_food_category_tenant ON food_category(tenant_id);
CREATE INDEX IF NOT EXISTS idx_food_category_name ON food_category(name);
CREATE INDEX IF NOT EXISTS idx_food_item_tenant ON food_item(tenant_id);
CREATE INDEX IF NOT EXISTS idx_food_item_category ON food_item(tenant_id, category_id);
CREATE INDEX IF NOT EXISTS idx_food_item_availability ON food_item(availability_status);
CREATE INDEX IF NOT EXISTS idx_food_item_is_vegetarian ON food_item(is_vegetarian);

INSERT INTO tenant (id, name, description)
VALUES ('1', 'default', 'Default local development tenant')
ON CONFLICT (id) DO NOTHING;
