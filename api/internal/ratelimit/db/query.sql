-- name: UpsertRateLimitCounter :one
-- Atomically increments counter, creating it if it doesn't exist.
-- Returns the new count after increment.
INSERT INTO rate_limit_counters (key, count, expires_at)
VALUES ($1, 1, $2)
ON CONFLICT (key) DO UPDATE
    SET count = rate_limit_counters.count + 1,
        -- Refresh expiry only if the incoming expiry is later (prevents shrinking TTL)
        expires_at = GREATEST(rate_limit_counters.expires_at, EXCLUDED.expires_at)
RETURNING count;

-- name: GetRateLimitCounters :many
-- Fetches two counter values in one round-trip (minute + day key).
SELECT key, count FROM rate_limit_counters
WHERE key = ANY($1::text[]) AND expires_at > NOW();

-- name: DeleteExpiredRateLimitCounters :exec
DELETE FROM rate_limit_counters WHERE expires_at < NOW();
