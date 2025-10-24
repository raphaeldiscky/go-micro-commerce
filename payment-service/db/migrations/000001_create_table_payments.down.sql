BEGIN;

DROP TABLE IF EXISTS payments CASCADE;
DROP INDEX IF EXISTS idx_payments_expires_at_status;
DROP INDEX IF EXISTS idx_payments_order_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_gateway_txn_id;

COMMIT;
