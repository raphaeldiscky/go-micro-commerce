CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS orders(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key UUID NOT NULL UNIQUE,
    customer_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS order_items(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products (id),
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_order_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_order_created_at ON orders (created_at);
CREATE INDEX IF NOT EXISTS idx_order_customer_id ON orders (customer_id);
CREATE INDEX IF NOT EXISTS idx_order_idempotency_key ON orders (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_fk_order_item_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_fk_order_item_product_id ON order_items (product_id);
