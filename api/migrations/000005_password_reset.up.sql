ALTER TABLE users
ADD COLUMN password_reset_token VARCHAR(255) UNIQUE,
ADD COLUMN password_reset_expires_at TIMESTAMPTZ;
