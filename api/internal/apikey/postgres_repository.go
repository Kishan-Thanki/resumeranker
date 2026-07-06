package apikey

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, apiKey *APIKey) (*APIKey, error) {
	const sql = `
		INSERT INTO api_keys (user_id, name, key_selector, key_hash, status, token_quota, tokens_used, expires_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		apiKey.UserID,
		apiKey.Name,
		apiKey.KeySelector,
		apiKey.KeyHash,
		apiKey.Status,
		apiKey.TokenQuota,
		apiKey.TokensUsed,
		apiKey.ExpiresAt,
		apiKey.LastUsedAt,
	).Scan(
		&apiKey.ID,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint64) (*APIKey, error) {
	const sql = `
		SELECT id, user_id, name, key_selector, key_hash, status, token_quota, tokens_used, expires_at, last_used_at, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE id = $1 AND deleted_at IS NULL
	`
	apiKey := &APIKey{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.KeySelector,
		&apiKey.KeyHash,
		&apiKey.Status,
		&apiKey.TokenQuota,
		&apiKey.TokensUsed,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
		&apiKey.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (r *PostgresRepository) GetBySelector(ctx context.Context, selector string) (*APIKey, error) {
	const sql = `
		SELECT id, user_id, name, key_selector, key_hash, status, token_quota, tokens_used, expires_at, last_used_at, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE key_selector = $1 AND deleted_at IS NULL
	`
	apiKey := &APIKey{}
	err := r.db.QueryRow(ctx, sql, selector).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.KeySelector,
		&apiKey.KeyHash,
		&apiKey.Status,
		&apiKey.TokenQuota,
		&apiKey.TokensUsed,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
		&apiKey.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (r *PostgresRepository) Update(ctx context.Context, apiKey *APIKey) (*APIKey, error) {
	const sql = `
		UPDATE api_keys
		SET name = $1, status = $2, token_quota = $3, tokens_used = $4, expires_at = $5, last_used_at = $6, updated_at = NOW()
		WHERE id = $7 AND deleted_at IS NULL
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		apiKey.Name,
		apiKey.Status,
		apiKey.TokenQuota,
		apiKey.TokensUsed,
		apiKey.ExpiresAt,
		apiKey.LastUsedAt,
		apiKey.ID,
	).Scan(&apiKey.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (r *PostgresRepository) ListByUserID(ctx context.Context, userID uint64) ([]*APIKey, error) {
	const sql = `
		SELECT id, user_id, name, key_selector, key_hash, status, token_quota, tokens_used, expires_at, last_used_at, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		apiKey := &APIKey{}
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Name,
			&apiKey.KeySelector,
			&apiKey.KeyHash,
			&apiKey.Status,
			&apiKey.TokenQuota,
			&apiKey.TokensUsed,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
			&apiKey.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, apiKey)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint64) error {
	const sql = `
		UPDATE api_keys
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}
