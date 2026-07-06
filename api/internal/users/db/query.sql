-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, role, status, metadata
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET email = $2, password_hash = $3, role = $4, status = $5, metadata = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1;

-- name: CreateAgreement :one
INSERT INTO agreements (
    type, version, document_url, published_at
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
