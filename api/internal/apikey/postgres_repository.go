package apikey

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey/db"
)

type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *PostgresRepository) Create(ctx context.Context, apiKey *APIKey) (*APIKey, error) {

	k, err := r.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UserID:            int64(apiKey.UserID),
		Name:              apiKey.Name,
		KeyPrefix:         apiKey.KeyPrefix,
		KeySelector:       apiKey.KeySelector,
		KeyHash:           apiKey.KeyHash,
		Status:            string(apiKey.Status),
		RequestsPerMinute: int32(apiKey.RequestsPerMinute),
		RequestsPerDay:    int32(apiKey.RequestsPerDay),
		TokenQuota:        int64(apiKey.TokenQuota),
		TokensUsed:        int64(apiKey.TokensUsed),
		ExpiresAt:         toPgTimestamp(apiKey.ExpiresAt),
	})
	if err != nil {
		return nil, err
	}

	apiKey.ID = uint64(k.ID)
	apiKey.CreatedAt = k.CreatedAt.Time
	apiKey.UpdatedAt = k.UpdatedAt.Time

	return apiKey, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint64) (*APIKey, error) {

	k, err := r.queries.GetAPIKeyByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	return mapDBAPIKeyToModel(k), nil
}

func (r *PostgresRepository) GetBySelector(ctx context.Context, selector string) (*APIKey, error) {

	k, err := r.queries.GetAPIKeyBySelector(ctx, selector)
	if err != nil {
		return nil, err
	}

	return mapDBAPIKeyToModel(k), nil
}

func (r *PostgresRepository) Update(ctx context.Context, apiKey *APIKey) (*APIKey, error) {

	k, err := r.queries.UpdateAPIKey(ctx, db.UpdateAPIKeyParams{
		ID:         int64(apiKey.ID),
		Status:     string(apiKey.Status),
		TokenQuota: int64(apiKey.TokenQuota),
		TokensUsed: int64(apiKey.TokensUsed),
		ExpiresAt:  toPgTimestamp(apiKey.ExpiresAt),
		LastUsedAt: toPgTimestamp(apiKey.LastUsedAt),
	})
	if err != nil {
		return nil, err
	}

	apiKey.UpdatedAt = k.UpdatedAt.Time

	return apiKey, nil
}

func (r *PostgresRepository) ListByUserID(ctx context.Context, userID uint64) ([]*APIKey, error) {

	dbKeys, err := r.queries.ListAPIKeysByUserID(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	keys := make([]*APIKey, len(dbKeys))
	for i, k := range dbKeys {
		keys[i] = mapDBAPIKeyToModel(k)
	}

	return keys, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint64) error {
	return r.queries.DeleteAPIKey(ctx, int64(id))
}

func (r *PostgresRepository) IsUserActive(ctx context.Context, userID uint64) (bool, error) {

	var status string
	err := r.pool.QueryRow(ctx, "SELECT status FROM users WHERE id = $1 AND deleted_at IS NULL", int64(userID)).Scan(&status)
	if err != nil {
		return false, err
	}

	return status == "active", nil
}

func (r *PostgresRepository) GetUserEmailByID(ctx context.Context, userID uint64) (string, error) {
	return r.queries.GetUserEmailByID(ctx, int64(userID))
}

func mapDBAPIKeyToModel(k db.ApiKey) *APIKey {

	var deletedAt *time.Time
	if k.DeletedAt.Valid {
		deletedAt = &k.DeletedAt.Time
	}

	var expiresAt *time.Time
	if k.ExpiresAt.Valid {
		expiresAt = &k.ExpiresAt.Time
	}

	var lastUsedAt *time.Time
	if k.LastUsedAt.Valid {
		lastUsedAt = &k.LastUsedAt.Time
	}

	return &APIKey{
		ID:                uint64(k.ID),
		UserID:            uint64(k.UserID),
		Name:              k.Name,
		KeyPrefix:         k.KeyPrefix,
		KeySelector:       k.KeySelector,
		KeyHash:           k.KeyHash,
		Status:            APIKeyStatus(k.Status),
		RequestsPerMinute: uint64(k.RequestsPerMinute),
		RequestsPerDay:    uint64(k.RequestsPerDay),
		TokenQuota:        uint64(k.TokenQuota),
		TokensUsed:        uint64(k.TokensUsed),
		ExpiresAt:         expiresAt,
		LastUsedAt:        lastUsedAt,
		CreatedAt:         k.CreatedAt.Time,
		UpdatedAt:         k.UpdatedAt.Time,
		DeletedAt:         deletedAt,
	}
}

func toPgTimestamp(t *time.Time) pgtype.Timestamptz {

	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}

	return pgtype.Timestamptz{Time: *t, Valid: true}
}
