-- Canonical local schema for tenant / food_category / food_item.
-- Aligned with internal/repository SQL and models
-- (string IDs, unix-millis timestamps, soft-delete via is_active).
-- Applied by docker compose via /docker-entrypoint-initdb.d on first volume init.
--
-- Multi-tenancy: every catalog row is scoped by tenant_id.
-- Primary keys are composite (tenant_id, id). Foreign keys to categories
-- are also composite so cross-tenant category references are impossible.
-- REQFOOD001: food_item has no price column.

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
-- Menu hierarchy + size units for local docker init.
-- Keep aligned with migrations/000002_size_unit.up.sql and
-- migrations/000003_menu_hierarchy.up.sql.
-- REQSIZE001–REQSIZE002, REQMENU001, REQMENU005–REQMENU009, REQITEM001, REQITEM003–REQITEM005.

CREATE TABLE IF NOT EXISTS size_unit (
    id           VARCHAR(50) PRIMARY KEY,
    tenant_id    VARCHAR(50) REFERENCES tenant(id) ON DELETE CASCADE,
    code         VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    is_system    BOOLEAN NOT NULL DEFAULT FALSE,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at   BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    CONSTRAINT size_unit_system_tenant_null CHECK (
        (is_system = TRUE AND tenant_id IS NULL)
        OR (is_system = FALSE AND tenant_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS size_unit_system_code_uq
    ON size_unit (code)
    WHERE is_system = TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS size_unit_tenant_code_uq
    ON size_unit (tenant_id, code)
    WHERE is_system = FALSE AND tenant_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_size_unit_tenant ON size_unit(tenant_id);

INSERT INTO size_unit (id, tenant_id, code, display_name, is_system, is_active)
VALUES
    ('su_serving', NULL, 'serving', 'Serving', TRUE, TRUE),
    ('su_tray',    NULL, 'tray',    'Tray',    TRUE, TRUE),
    ('su_piece',   NULL, 'piece',   'Piece',   TRUE, TRUE),
    ('su_dozen',   NULL, 'dozen',   'Dozen',   TRUE, TRUE),
    ('su_kg',      NULL, 'kg',      'Kilogram', TRUE, TRUE),
    ('su_g',       NULL, 'g',       'Gram',    TRUE, TRUE),
    ('su_liter',   NULL, 'liter',   'Liter',   TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS menu (
    id             VARCHAR(50) NOT NULL,
    tenant_id      VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    name           VARCHAR(200) NOT NULL,
    description    TEXT,
    menu_type      VARCHAR(20) NOT NULL
        CHECK (menu_type IN ('daily', 'weekly', 'catering')),
    status         VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published')),
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    menu_date      DATE,
    start_date     DATE,
    end_date       DATE,
    event_date     DATE,
    event_location TEXT,
    created_at     BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at     BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT menu_weekly_dates_order CHECK (
        menu_type <> 'weekly'
        OR end_date IS NULL
        OR start_date IS NULL
        OR end_date >= start_date
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS menu_weekly_published_start_date_uq
    ON menu (tenant_id, start_date)
    WHERE menu_type = 'weekly'
      AND status = 'published'
      AND is_active = TRUE
      AND start_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_menu_tenant ON menu(tenant_id);
CREATE INDEX IF NOT EXISTS idx_menu_tenant_status ON menu(tenant_id, status) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_menu_tenant_type ON menu(tenant_id, menu_type);

CREATE TABLE IF NOT EXISTS menu_category (
    id          VARCHAR(50) NOT NULL,
    tenant_id   VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    menu_id     VARCHAR(50) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    sequence    INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT menu_category_menu_fk
        FOREIGN KEY (tenant_id, menu_id)
        REFERENCES menu (tenant_id, id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_menu_category_menu ON menu_category(tenant_id, menu_id);

CREATE TABLE IF NOT EXISTS menu_item (
    id          VARCHAR(50) NOT NULL,
    tenant_id   VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    menu_id     VARCHAR(50) NOT NULL,
    category_id VARCHAR(50) NOT NULL,
    kind        VARCHAR(20) NOT NULL
        CHECK (kind IN ('simple', 'combo')),
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    sequence    INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at  BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT menu_item_menu_fk
        FOREIGN KEY (tenant_id, menu_id)
        REFERENCES menu (tenant_id, id)
        ON DELETE CASCADE,
    CONSTRAINT menu_item_category_fk
        FOREIGN KEY (tenant_id, category_id)
        REFERENCES menu_category (tenant_id, id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_menu_item_menu ON menu_item(tenant_id, menu_id);
CREATE INDEX IF NOT EXISTS idx_menu_item_category ON menu_item(tenant_id, category_id);

CREATE TABLE IF NOT EXISTS menu_item_component (
    id                     VARCHAR(50) NOT NULL,
    tenant_id              VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    menu_item_id           VARCHAR(50) NOT NULL,
    food_item_id           VARCHAR(50) NOT NULL,
    default_size_option_id VARCHAR(50),
    is_active              BOOLEAN NOT NULL DEFAULT TRUE,
    created_at             BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at             BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT menu_item_component_item_fk
        FOREIGN KEY (tenant_id, menu_item_id)
        REFERENCES menu_item (tenant_id, id)
        ON DELETE CASCADE,
    CONSTRAINT menu_item_component_food_fk
        FOREIGN KEY (tenant_id, food_item_id)
        REFERENCES food_item (tenant_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT menu_item_component_food_unique
        UNIQUE (tenant_id, menu_item_id, food_item_id)
);

CREATE INDEX IF NOT EXISTS idx_menu_item_component_item ON menu_item_component(tenant_id, menu_item_id);

CREATE TABLE IF NOT EXISTS menu_item_component_size_option (
    id           VARCHAR(50) NOT NULL,
    tenant_id    VARCHAR(50) NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    component_id VARCHAR(50) NOT NULL,
    size_unit_id VARCHAR(50) NOT NULL REFERENCES size_unit(id) ON DELETE RESTRICT,
    qty          NUMERIC(12,4) NOT NULL CHECK (qty > 0),
    price        NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    updated_at   BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM TIMEZONE('UTC', NOW())) * 1000)::BIGINT,
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT menu_item_component_size_option_component_fk
        FOREIGN KEY (tenant_id, component_id)
        REFERENCES menu_item_component (tenant_id, id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_menu_item_component_size_option_component
    ON menu_item_component_size_option(tenant_id, component_id);

ALTER TABLE menu_item_component
    DROP CONSTRAINT IF EXISTS menu_item_component_default_size_fk;

ALTER TABLE menu_item_component
    ADD CONSTRAINT menu_item_component_default_size_fk
        FOREIGN KEY (tenant_id, default_size_option_id)
        REFERENCES menu_item_component_size_option (tenant_id, id)
        ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED;
