BEGIN;

-- Add expires_at column for tracking 24-hour payment window
ALTER TABLE payments
ADD COLUMN expires_at TIMESTAMPTZ;

-- Create index for efficient timeout job queries
-- Allows fast lookup of pending payments that have expired
CREATE INDEX idx_payments_expires_at_status
ON payments (expires_at, status)
WHERE status = 'pending';

-- Add comment for documentation
COMMENT ON COLUMN payments.expires_at IS '24-hour payment window expiry timestamp. Payments automatic timed out';

COMMIT;
