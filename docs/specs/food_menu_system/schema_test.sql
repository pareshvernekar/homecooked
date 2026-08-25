-- Schema for integration testing with TestContainers
-- This schema defines the structure needed for comprehensive database verification

CREATE TABLE food_category (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_name CHECK (name ~ '^[a-z][a-z0-9_]*$')
);

CREATE TABLE food_item (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    availability_status VARCHAR(50) NOT NULL CHECK (availability_status IN ('available', 'low_stock', 'unavailable')),
    category_id UUID REFERENCES food_category(id) ON DELETE CASCADE,
    image_url TEXT,
    avoidance TEXT,
    is_vegetarian BOOLEAN DEFAULT FALSE,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_category FOREIGN KEY (category_id)
        MATCHES (SELECT id FROM food_category WHERE tenant_id = current_setting('app.current_tenant_id'::text, false)::uuid),
    CONSTRAINT valid_name CHECK (name ~ '^[a-z][a-z0-9_]+[ ]+.*$'),
    CONSTRAINT valid_price CHECK (price >= 0)
);

-- Indexes for query optimization
CREATE INDEX idx_food_category_tenant ON food_category(tenant_id);
CREATE INDEX idx_food_category_name ON food_category(name);
CREATE INDEX idx_food_item_tenant ON food_item(tenant_id);
CREATE INDEX idx_food_item_category ON food_item(category_id);
CREATE INDEX idx_food_item_availability ON food_item(availability_status);
