CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- This model treats the act of starting the checkout as "locking in" the state of the cart. 
-- It creates a snapshot of the items, their quantities, and their prices at that specific moment.
-- Will validate product price if match and stock still available
CREATE TABLE IF NOT EXISTS checkout_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key UUID NOT NULL,
    customer_id UUID NOT NULL,
    cart_id UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    destination JSONB, -- will be added in checkout page
    origin JSONB, -- will be added in checkout page
    courier JSONB, -- will be added in checkout page
    package JSONB, -- will be added in checkout page
    payment_gateway VARCHAR(50),
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    expired_at TIMESTAMPTZ
);

ALTER TABLE checkout_sessions
ADD CONSTRAINT chk_checkout_session_status
CHECK (status IN ('pending', 'order_placed', 'canceled', 'expired'));

CREATE INDEX idx_checkout_sessions_customer_id ON checkout_sessions (customer_id);
CREATE INDEX idx_checkout_sessions_status ON checkout_sessions (status);
CREATE INDEX idx_checkout_sessions_created_at ON checkout_sessions (created_at);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_checkout_sessions_updated_at
BEFORE UPDATE ON checkout_sessions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS checkout_session_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checkout_session_id UUID NOT NULL REFERENCES checkout_sessions (id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_name TEXT NOT NULL,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10, 2) NOT NULL CHECK (unit_price >= 0) -- Price per unit
);

CREATE INDEX IF NOT EXISTS idx_checkout_session_items_session_id ON checkout_session_items (checkout_session_id);
CREATE INDEX IF NOT EXISTS idx_checkout_session_items_product_id ON checkout_session_items (product_id);
