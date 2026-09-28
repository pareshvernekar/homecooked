-- +migrate Up
-- REQMENU001, REQMENU005–REQMENU009, REQITEM001, REQITEM003–REQITEM005:
-- Unified menu hierarchy with per-component size options.

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

-- REQMENU009: one active published weekly menu per tenant start_date.
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

-- default_size_option_id set after size options exist (circular FK resolved below).
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

-- Component default must reference a size option (same tenant).
ALTER TABLE menu_item_component
    ADD CONSTRAINT menu_item_component_default_size_fk
        FOREIGN KEY (tenant_id, default_size_option_id)
        REFERENCES menu_item_component_size_option (tenant_id, id)
        ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED;
