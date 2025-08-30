-- Add version column for optimistic locking
ALTER TABLE products ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

-- Add allocated_quantity column to track reserved stock
ALTER TABLE products ADD COLUMN allocated_quantity INT NOT NULL DEFAULT 0 CHECK (allocated_quantity >= 0);

-- Add constraint to ensure available stock is not negative
ALTER TABLE products ADD CONSTRAINT check_available_stock 
    CHECK (quantity >= allocated_quantity);

-- Create index for better performance on version-based queries
CREATE INDEX IF NOT EXISTS idx_products_version ON products(id, version);