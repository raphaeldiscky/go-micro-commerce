DROP INDEX IF EXISTS idx_products_version;

ALTER TABLE products DROP COLUMN IF EXISTS reserved_quantity;
ALTER TABLE products DROP COLUMN IF EXISTS version;
