BEGIN;

-- Add columns for storing Stripe payment method and customer IDs
ALTER TABLE payments
    ADD COLUMN payment_method_id TEXT,
    ADD COLUMN stripe_customer_id TEXT;

-- Create indexes for faster lookups
CREATE INDEX idx_payments_payment_method_id ON payments(payment_method_id);
CREATE INDEX idx_payments_stripe_customer_id ON payments(stripe_customer_id);

-- Add comments for documentation
COMMENT ON COLUMN payments.payment_method_id IS 'Stripe PaymentMethod ID (pm_xxx) for off-session charging';
COMMENT ON COLUMN payments.stripe_customer_id IS 'Stripe Customer ID (cus_xxx) for payment method attachment';

COMMIT;
