
-- Drop triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;

-- Drop indexes for sessions table
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_refresh_token;
DROP INDEX IF EXISTS idx_sessions_is_active;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_created_at;
DROP INDEX IF EXISTS idx_sessions_last_used_at;

-- Drop indexes for users table
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email_verification_token;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_is_email_verified;
DROP INDEX IF EXISTS idx_users_created_at;

-- Drop tables (sessions first due to foreign key dependency)
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

