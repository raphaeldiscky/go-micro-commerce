CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS orders(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key UUID NOT NULL UNIQUE,
    customer_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    total_tax DECIMAL(10, 2) NOT NULL CHECK (total_tax >= 0), -- Total tax applied
    total_discount DECIMAL(10, 2) NOT NULL CHECK (total_discount >= 0), -- Total discount applied
    total_price DECIMAL(10, 2) NOT NULL CHECK (total_price >= 0), -- Final payable amount SUM(order_items.total_price) - total_discount + total_tax
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    currency VARCHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    unit_price DECIMAL(10, 2) NOT NULL CHECK (unit_price >= 0), -- Price per unit
    total_tax DECIMAL(10, 2) NOT NULL CHECK (total_tax >= 0), -- Tax for this line item
    total_discount DECIMAL(10, 2) NOT NULL CHECK (total_discount >= 0), -- Discount for this line item
    total_price DECIMAL(10, 2) NOT NULL CHECK (total_price >= 0), -- (price * quantity) - total_discount + total_tax
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_order_created_at ON orders (created_at);
CREATE INDEX IF NOT EXISTS idx_order_customer_id ON orders (customer_id);
CREATE INDEX IF NOT EXISTS idx_order_idempotency_key ON orders (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_fk_order_item_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_fk_order_item_product_id ON order_items (product_id);