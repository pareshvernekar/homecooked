-- +migrate Up
-- REQSIZE001, REQSIZE002: size_unit table with system standards + tenant customs.

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

-- System unit codes are globally unique.
CREATE UNIQUE INDEX IF NOT EXISTS size_unit_system_code_uq
    ON size_unit (code)
    WHERE is_system = TRUE;

-- Tenant custom codes are unique per tenant.
CREATE UNIQUE INDEX IF NOT EXISTS size_unit_tenant_code_uq
    ON size_unit (tenant_id, code)
    WHERE is_system = FALSE AND tenant_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_size_unit_tenant ON size_unit(tenant_id);

-- Seed standard system units (REQSIZE001).
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
