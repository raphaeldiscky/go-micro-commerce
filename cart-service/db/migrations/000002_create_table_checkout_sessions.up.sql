CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS checkout_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key UUID NOT NULL,
    customer_id UUID NOT NULL,
    address_id UUID,
    carrier_id TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_gateway VARCHAR(50),
    payment_method VARCHAR(50),
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

ALTER TABLE checkout_sessions
ADD CONSTRAINT chk_checkout_session_status
CHECK (status IN ('pending', 'order_placed', 'canceled'));

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
    quantity BIGINT NOT NULL CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_checkout_session_items_session_id ON checkout_session_items (checkout_session_id);
CREATE INDEX IF NOT EXISTS idx_checkout_session_items_product_id ON checkout_session_items (product_id);
