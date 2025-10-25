BEGIN;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- payment transactions table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL, -- Reference to order ID (from order-service)
    amount DECIMAL(12, 2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, completed, etc
    payment_gateway VARCHAR(50) NOT NULL, -- stripe, midtrans, xendit, etc.
    gateway_transaction_id TEXT UNIQUE, -- transaction id from gateways (stripe Transaction `ipi_xxx`)
    gateway_metadata JSONB DEFAULT '{}'::JSONB, -- payment_method_id (pm_xxx), stripe_customer_id (cust_xxx), client_secret, etc
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL, -- payment expires at, set at payment tx initialization
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,

    CONSTRAINT chk_payments_status CHECK (
        status IN ('pending', 'processing', 'timeout', 'completed', 'failed', 'canceled', 'refunded')
    ),
    CONSTRAINT unique_gateway_txn UNIQUE (payment_gateway, gateway_transaction_id)
);

CREATE INDEX idx_payments_order_id ON payments (order_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_created_at ON payments (created_at);
CREATE INDEX idx_payments_expires_at_status
ON payments (expires_at, status)
WHERE status = 'pending';
CREATE INDEX idx_payments_gateway_txn_id ON payments (gateway_transaction_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER trigger_payments_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;
