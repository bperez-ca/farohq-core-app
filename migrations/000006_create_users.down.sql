-- Rollback migration for users table

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_users_updated_at();

DROP TABLE IF EXISTS users;
