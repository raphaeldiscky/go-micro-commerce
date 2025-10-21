CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key UUID NOT NULL UNIQUE,
    customer_id UUID NOT NULL,
    address_id UUID NOT NULL,
    carrier_id TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_gateway VARCHAR(50) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    -- generated from fulfillment-service
    shipping_cost DECIMAL(10, 2) NOT NULL CHECK (shipping_cost >= 0),
    -- SUM(unit_price * quantity) for all items
    subtotal DECIMAL(10, 2) NOT NULL CHECK (subtotal >= 0),
    -- SUM(total_tax) for all items
    total_tax DECIMAL(10, 2) NOT NULL CHECK (total_tax >= 0),
    -- SUM(total_discount) for all items
    total_discount DECIMAL(10, 2) NOT NULL CHECK (total_discount >= 0),
    -- final payable amount 
    -- SUM(unit_price * qty) + SUM(tax) - SUM(discount) + shipping_cost
    total_price DECIMAL(10, 2) NOT NULL CHECK (total_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

ALTER TABLE orders
ADD CONSTRAINT chk_order_status
CHECK (
    status IN (
        'pending',
        'processing',
        'payment_pending',
        'payment_expired',
        'paid',
        'shipped',
        'delivered',
        'completed',
        'failed',
        'canceled'
    )
);

ALTER TABLE orders
ADD CONSTRAINT chk_order_currency
CHECK (currency ~ '^[A-Z]{3}$');

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    -- Price per unit
    unit_price DECIMAL(10, 2) NOT NULL CHECK (unit_price >= 0),
    tax_rate DECIMAL(5, 4) NOT NULL CHECK (tax_rate >= 0),
    -- Tax for this line item
    total_tax DECIMAL(10, 2) NOT NULL CHECK (total_tax >= 0),
    -- Discount for this line item
    total_discount DECIMAL(10, 2) NOT NULL CHECK (total_discount >= 0),
    -- (unit_price * quantity) - total_discount + total_tax
    total_price DECIMAL(10, 2) NOT NULL CHECK (total_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX IF NOT EXISTS idx_order_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_order_created_at ON orders (created_at);
CREATE INDEX IF NOT EXISTS idx_order_customer_id ON orders (customer_id);
CREATE INDEX IF NOT EXISTS idx_order_idempotency_key ON orders (
    idempotency_key
);
CREATE INDEX IF NOT EXISTS idx_fk_order_item_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_fk_order_item_product_id ON order_items (
    product_id
);
