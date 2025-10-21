CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    quantity BIGINT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);
CREATE INDEX IF NOT EXISTS idx_product_search ON products USING gist (
    name gist_trgm_ops (siglen = 64)
);
CREATE INDEX IF NOT EXISTS idx_product_price ON products (price);
