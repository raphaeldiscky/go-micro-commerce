-- Remove constraint first
ALTER TABLE products DROP CONSTRAINT IF EXISTS check_available_stock;

-- Remove index
DROP INDEX IF EXISTS idx_products_version;

-- Remove columns
ALTER TABLE products DROP COLUMN IF EXISTS reserved_quantity;
ALTER TABLE products DROP COLUMN IF EXISTS version;