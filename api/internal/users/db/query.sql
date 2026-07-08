-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, role, status, metadata, is_verified, verification_token, verification_expires_at, password_reset_token, password_reset_expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET email = $2, password_hash = $3, role = $4, status = $5, metadata = $6, is_verified = $7, verification_token = $8, verification_expires_at = $9, password_reset_token = $10, password_reset_expires_at = $11, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserByVerificationToken :one
SELECT * FROM users
WHERE verification_token = $1 AND deleted_at IS NULL;

-- name: GetUserByPasswordResetToken :one
SELECT * FROM users
WHERE password_reset_token = $1 AND deleted_at IS NULL;

-- name: VerifyUserEmail :one
UPDATE users
SET is_verified = true, verification_token = NULL, verification_expires_at = NULL, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CreateAgreement :one
INSERT INTO agreements (
    type, version, content, published_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetAgreementByID :one
SELECT * FROM agreements
WHERE id = $1;

-- name: GetAgreementByTypeAndVersion :one
SELECT * FROM agreements
WHERE type = $1 AND version = $2;

-- name: CreateUserAgreement :one
INSERT INTO user_agreements (
    user_id, agreement_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: HasUserAcceptedAgreement :one
SELECT EXISTS (
    SELECT 1 FROM user_agreements
    WHERE user_id = $1 AND agreement_id = $2
);

-- name: GetLatestAgreements :many
SELECT DISTINCT ON (type) * FROM agreements
ORDER BY type, published_at DESC;

-- name: GetPendingAgreementsForUser :many
SELECT a.* FROM agreements a
INNER JOIN (
    SELECT type, MAX(published_at) as max_published_at
    FROM agreements
    GROUP BY type
) latest_a ON a.type = latest_a.type AND a.published_at = latest_a.max_published_at
LEFT JOIN user_agreements ua ON a.id = ua.agreement_id AND ua.user_id = $1
WHERE ua.id IS NULL;
