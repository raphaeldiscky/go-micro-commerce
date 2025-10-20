BEGIN;

-- Drop indexes
DROP INDEX IF EXISTS idx_payments_stripe_customer_id;
DROP INDEX IF EXISTS idx_payments_payment_method_id;

-- Drop columns
ALTER TABLE payments
    DROP COLUMN IF EXISTS stripe_customer_id,
    DROP COLUMN IF EXISTS payment_method_id;

COMMIT;
