DROP TRIGGER IF EXISTS trg_update_notifications_updated_at ON notifications;
DROP FUNCTION IF EXISTS update_updated_at_column;
DROP INDEX IF EXISTS idx_notifications_user_created;
DROP INDEX IF EXISTS idx_notifications_user_unread;
DROP INDEX IF EXISTS idx_notifications_user_id;
DROP TABLE IF EXISTS notifications;
