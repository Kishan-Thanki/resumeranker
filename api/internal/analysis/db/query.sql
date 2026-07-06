-- name: CreateAnalysisRequest :one
INSERT INTO analysis_requests (
    request_id, user_id, api_key_id, status, error, metadata, started_at, completed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetAnalysisRequestByID :one
SELECT * FROM analysis_requests
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAnalysisRequestsByUserID :many
SELECT * FROM analysis_requests
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateAnalysisRequest :one
UPDATE analysis_requests
SET status = $2, error = $3, metadata = $4, started_at = $5, completed_at = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAnalysisRequest :exec
UPDATE analysis_requests
SET deleted_at = NOW()
WHERE id = $1;


-- name: CreateAnalysisResult :one
INSERT INTO analysis_results (
    analysis_request_id, model, result, prompt_tokens, completion_tokens, total_tokens
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetAnalysisResultByRequestID :one
SELECT * FROM analysis_results
WHERE analysis_request_id = $1 AND deleted_at IS NULL;

-- name: DeleteAnalysisResult :exec
UPDATE analysis_results
SET deleted_at = NOW()
WHERE id = $1;
