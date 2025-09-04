BEGIN;
CREATE EXTENSION IF NOT EXISTS pg_trgm;


CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL, -- Reference to order ID (from order-service)
    amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed, refunded
    payment_method VARCHAR(50) NOT NULL, -- credit_card, bank_transfer, paypal, etc.
    payment_gateway VARCHAR(50), -- stripe, midtrans, xendit, etc.
    gateway_reference_id VARCHAR(255), -- Reference ID from payment gateway
    gateway_response JSONB, -- Raw response from payment gateway
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ
);

CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_gateway_reference ON payments(gateway_reference_id);
CREATE INDEX idx_payments_payment_method ON payments(payment_method);

ALTER TABLE payments 
ADD CONSTRAINT chk_payments_status 
CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'refunded'));

ALTER TABLE payments 
ADD CONSTRAINT chk_payments_amount 
CHECK (amount > 0);

ALTER TABLE payments 
ADD CONSTRAINT chk_payments_currency 
CHECK (currency ~ '^[A-Z]{3}$');


COMMENT ON TABLE payments IS 'Stores payment transactions';
COMMENT ON COLUMN payments.order_id IS 'Reference to the order in order-service';
COMMENT ON COLUMN payments.gateway_reference_id IS 'External payment gateway transaction ID';
COMMENT ON COLUMN payments.gateway_response IS 'Raw response from payment gateway for debugging and reconciliation';
COMMENT ON COLUMN payments.payment_method IS 'Payment method used (credit_card, bank_transfer, e-wallet, etc.)';
COMMENT ON COLUMN payments.payment_gateway IS 'Payment gateway provider (stripe, midtrans, xendit, etc.)';
COMMENT ON COLUMN payments.status IS 'Current payment status (pending, processing, completed, failed, refunded)';
COMMENT ON COLUMN payments.amount IS 'Payment amount, must be greater than 0';
COMMENT ON COLUMN payments.currency IS 'Currency code in ISO 4217 format (3 uppercase letters)';

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