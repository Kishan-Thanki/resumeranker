-- name: CreateAPIKey :one
INSERT INTO api_keys (
    user_id, name, key_prefix, key_selector, key_hash, status, token_quota, tokens_used, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetAPIKeyByID :one
SELECT * FROM api_keys
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetAPIKeyBySelector :one
SELECT * FROM api_keys
WHERE key_selector = $1 AND deleted_at IS NULL;

-- name: ListAPIKeysByUserID :many
SELECT * FROM api_keys
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateAPIKey :one
UPDATE api_keys
SET status = $2, token_quota = $3, tokens_used = $4, expires_at = $5, last_used_at = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAPIKey :exec
UPDATE api_keys
SET deleted_at = NOW()
WHERE id = $1;
