-- Canonical local schema for tenant / food_category / food_item.
-- Aligned with internal/repository SQL and models
-- (string IDs, unix-millis timestamps, soft-delete via is_active).
-- Applied by docker compose via /docker-entrypoint-initdb.d on first volume init.
--
-- Multi-tenancy: every catalog row is scoped by tenant_id.
-- Primary keys are composite (tenant_id, id). Foreign keys to categories
-- are also composite so cross-tenant category references are impossible.

CREATE TABLE IF NOT EXISTS tenant (
    id          VARCHAR(50) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT
);

CREATE TABLE IF NOT EXISTS food_category (
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

CREATE TABLE IF NOT EXISTS food_item (
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

-- Default local/dev tenant used by make local (TENANT_ID defaults to "1").
INSERT INTO tenant (id, name, description)
VALUES ('1', 'default', 'Default local development tenant')
ON CONFLICT (id) DO NOTHING;
