-- Schema for TestContainers Integration Testing
-- Must stay aligned with scripts/init/01_food_catalog.sql and repository SQL.

CREATE TABLE IF NOT EXISTS food_category (
    id          VARCHAR(50) PRIMARY KEY,
    tenant_id   VARCHAR(50) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    CONSTRAINT food_category_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_food_category_tenant ON food_category(tenant_id);
CREATE INDEX IF NOT EXISTS idx_food_category_name ON food_category(name);

CREATE TABLE IF NOT EXISTS food_item (
    id                  VARCHAR(50) PRIMARY KEY,
    tenant_id           VARCHAR(50) NOT NULL,
    name                VARCHAR(100) NOT NULL,
    description         TEXT,
    price               NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    availability_status VARCHAR(50) NOT NULL
        CHECK (availability_status IN ('available', 'low_stock', 'unavailable')),
    category_id         VARCHAR(50) REFERENCES food_category(id) ON DELETE CASCADE,
    image_url           TEXT,
    avoidance           TEXT,
    is_vegetarian       BOOLEAN NOT NULL DEFAULT FALSE,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at          BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    delivered_at        BIGINT
);

CREATE INDEX IF NOT EXISTS idx_food_item_tenant ON food_item(tenant_id);
CREATE INDEX IF NOT EXISTS idx_food_item_category ON food_item(category_id);
CREATE INDEX IF NOT EXISTS idx_food_item_availability ON food_item(availability_status);
CREATE INDEX IF NOT EXISTS idx_food_item_is_vegetarian ON food_item(is_vegetarian);
