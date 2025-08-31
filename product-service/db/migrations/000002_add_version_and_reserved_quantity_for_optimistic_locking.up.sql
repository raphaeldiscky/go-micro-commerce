-- Add version column for optimistic locking
ALTER TABLE products ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

-- Add reserved_quantity column to track reserved stock
ALTER TABLE products ADD COLUMN reserved_quantity BIGINT NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0);

-- Add constraint to ensure available stock is not negative
ALTER TABLE products ADD CONSTRAINT check_available_stock 
    CHECK (quantity >= reserved_quantity);

-- Create index for better performance on version-based queries
CREATE INDEX IF NOT EXISTS idx_products_version ON products(id, version);