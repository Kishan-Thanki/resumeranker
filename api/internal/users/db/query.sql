-- name: CreateUser :one
INSERT INTO users (
	email,
	password_hash,
	role,
	status,
	metadata,
	is_verified,
	verification_token,
	verification_expires_at,
	password_reset_token,
	password_reset_expires_at
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10
)
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC, id DESC
LIMIT $1
OFFSET $2;

-- name: UpdateUserProfile :one
UPDATE users
SET
	email = $2,
	metadata = $3,
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserRole :one
UPDATE users
SET
	role = $2,
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users
SET
	status = $2,
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET
	password_hash = $2,
	password_reset_token = NULL,
	password_reset_expires_at = NULL,
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: SetVerificationToken :one
UPDATE users
SET
	verification_token = $2,
	verification_expires_at = $3,
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: GetUserByVerificationToken :one
SELECT *
FROM users
WHERE verification_token = $1
  AND verification_expires_at > NOW()
  AND deleted_at IS NULL;

-- name: VerifyUserEmail :one
UPDATE users
SET
	is_verified = TRUE,
	verification_token = NULL,
	verification_expires_at = NULL,
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: SetPasswordResetToken :one
UPDATE users
SET
	password_reset_token = $2,
	password_reset_expires_at = $3,
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: GetUserByPasswordResetToken :one
SELECT *
FROM users
WHERE password_reset_token = $1
  AND password_reset_expires_at > NOW()
  AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET
	email = $2,
	password_hash = $3,
	role = $4,
	status = $5,
	metadata = $6,
	is_verified = $7,
	verification_token = $8,
	verification_expires_at = $9,
	password_reset_token = $10,
	password_reset_expires_at = $11,
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteUser :exec
UPDATE users
SET
	deleted_at = NOW(),
	updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: CreateAgreement :one
INSERT INTO agreements (
	type,
	version,
	content,
	published_at
) VALUES (
	$1,
	$2,
	$3,
	$4
)
RETURNING *;

-- name: GetAgreementByID :one
SELECT *
FROM agreements
WHERE id = $1;

-- name: GetAgreementByTypeAndVersion :one
SELECT *
FROM agreements
WHERE type = $1
  AND version = $2;

-- name: CreateUserAgreement :one
INSERT INTO user_agreements (
	user_id,
	agreement_id
) VALUES (
	$1,
	$2
)
RETURNING *;

-- name: HasUserAcceptedAgreement :one
SELECT EXISTS (
	SELECT 1
	FROM user_agreements
	WHERE user_id = $1
	  AND agreement_id = $2
);

-- name: GetLatestAgreements :many
SELECT DISTINCT ON (type) *
FROM agreements
ORDER BY
	type,
	published_at DESC,
	id DESC;

-- name: GetPendingAgreementsForUser :many
SELECT latest.*
FROM (
	SELECT DISTINCT ON (type)
		id,
		type,
		version,
		published_at,
		created_at,
		updated_at,
		content
	FROM agreements
	ORDER BY
		type,
		published_at DESC,
		id DESC
) AS latest
LEFT JOIN user_agreements AS ua
  ON ua.agreement_id = latest.id
 AND ua.user_id = $1
WHERE ua.id IS NULL
ORDER BY latest.type, latest.id DESC;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users
WHERE deleted_at IS NULL;
