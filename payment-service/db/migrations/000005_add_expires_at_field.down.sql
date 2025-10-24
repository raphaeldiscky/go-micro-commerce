BEGIN;

-- Drop index
DROP INDEX IF EXISTS idx_payments_expires_at_status;

-- Remove expires_at column
ALTER TABLE payments
DROP COLUMN IF EXISTS expires_at;

COMMIT;
