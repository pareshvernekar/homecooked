-- +migrate Down

ALTER TABLE food_item
    ADD COLUMN IF NOT EXISTS price NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (price >= 0);
