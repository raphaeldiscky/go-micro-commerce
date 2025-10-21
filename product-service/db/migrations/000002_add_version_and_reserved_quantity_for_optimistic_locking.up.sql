ALTER TABLE products ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

ALTER TABLE products ADD COLUMN reserved_quantity BIGINT NOT NULL DEFAULT 0 CHECK (
    reserved_quantity >= 0
);

CREATE INDEX IF NOT EXISTS idx_products_version ON products (id, version);
