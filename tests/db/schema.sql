-- Schema for TestContainers Integration Testing
-- PostgreSQL 14 (Alpine) - Compatible with production database structure

-- ================================
-- food_category table
-- ================================
CREATE TABLE IF NOT EXISTS food_category (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_name CHECK (name ~ '^[a-z][a-z0-9_]*$')
);

-- Indexes for food_category table
CREATE INDEX idx_food_category_tenant_id ON food_category(tenant_id);
CREATE INDEX idx_food_category_name ON food_category(name);

-- ================================
-- food_item table
-- ================================
CREATE TABLE IF NOT EXISTS food_item (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    availability_status VARCHAR(50) NOT NULL
        CHECK (availability_status IN ('available', 'low_stock', 'unavailable')),
    category_id UUID REFERENCES food_category(id) ON DELETE CASCADE,
    image_url TEXT,
    avoidance TEXT,
    is_vegetarian BOOLEAN DEFAULT FALSE,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_name CHECK (name ~ '^[a-z][a-z0-9_]+[ ]+.*$'),
    CONSTRAINT fk_category_exists CHECK (
        category_id IS NOT NULL OR availability_status != 'available'
    ),
    CONSTRAINT valid_price CHECK (price >= 0)
);

-- Indexes for food_item table
CREATE INDEX idx_food_item_tenant_id ON food_item(tenant_id);
CREATE INDEX idx_food_item_category_id ON food_item(category_id);
CREATE INDEX idx_food_item_availability_status ON food_item(availability_status);
CREATE INDEX idx_food_item_is_vegetarian ON food_item(is_vegetarian);

-- ================================
-- Row Level Security (RLS) Policies
-- ================================
-- Enable RLS on food_category table
ALTER TABLE food_category ENABLE ROW LEVEL SECURITY;

-- Create tenant isolation policy (simplified for testing - uses true in all cases)
CREATE OR REPLACE POLICY tenant_isolation ON food_category USING (true);

-- Enable RLS on food_item table
ALTER TABLE food_item ENABLE ROW LEVEL SECURITY;

-- Create tenant isolation policy for food_item
CREATE OR REPLACE POLICY tenant_isolation ON food_item USING (true);
