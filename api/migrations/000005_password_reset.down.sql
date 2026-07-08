ALTER TABLE users
DROP COLUMN IF EXISTS password_reset_token,
DROP COLUMN IF EXISTS password_reset_expires_at;
