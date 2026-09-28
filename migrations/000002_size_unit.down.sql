-- +migrate Down

DROP INDEX IF EXISTS idx_size_unit_tenant;
DROP INDEX IF EXISTS size_unit_tenant_code_uq;
DROP INDEX IF EXISTS size_unit_system_code_uq;
DROP TABLE IF EXISTS size_unit;
