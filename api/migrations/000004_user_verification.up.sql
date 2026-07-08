BEGIN;

ALTER TABLE users 
ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN verification_token VARCHAR(255),
ADD COLUMN verification_expires_at TIMESTAMP WITH TIME ZONE;

-- For existing seeded users, mark them as verified so they can log in
UPDATE users SET is_verified = true;

COMMIT;
