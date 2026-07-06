-- name: CreateAuditEvent :one
INSERT INTO audit_events (
    user_id, api_key_id, analysis_request_id, type, description, ip_address, user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListAuditEvents :many
SELECT * FROM audit_events
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
