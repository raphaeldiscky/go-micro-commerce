CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS checkout_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    cart_id UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    address_id UUID,
    carrier_id TEXT,
    -- pending, order_placed, canceled
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_gateway VARCHAR(50),
    payment_method VARCHAR(50),
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
