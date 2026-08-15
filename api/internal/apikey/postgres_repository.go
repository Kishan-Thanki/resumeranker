package apikey

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey/db"
	"github.com/kishan-thanki/resumeranker/api/internal/pgutil"
)

type PostgresRepository struct {
	queries *db.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		queries: db.New(pool),
	}
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	apiKey *APIKey,
) (*APIKey, error) {
	if apiKey == nil {
		return nil, errors.New("api key cannot be nil")
	}
	if !apiKey.Status.IsValid() {
		return nil, fmt.Errorf("invalid API key status: %q", apiKey.Status)
	}

	userID, err := uint64ToInt64("user_id", apiKey.UserID)
	if err != nil {
		return nil, err
	}

	requestsPerMinute, err := uint64ToInt32(
		"requests_per_minute",
		apiKey.RequestsPerMinute,
	)
	if err != nil {
		return nil, err
	}

	requestsPerDay, err := uint64ToInt32(
		"requests_per_day",
		apiKey.RequestsPerDay,
	)
	if err != nil {
		return nil, err
	}

	tokenQuota, err := uint64ToInt64("token_quota", apiKey.TokenQuota)
	if err != nil {
		return nil, err
	}

	tokensUsed, err := uint64ToInt64("tokens_used", apiKey.TokensUsed)
	if err != nil {
		return nil, err
	}

	k, err := r.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UserID:            userID,
		Name:              apiKey.Name,
		KeyPrefix:         apiKey.KeyPrefix,
		KeySelector:       apiKey.KeySelector,
		KeyHash:           apiKey.KeyHash,
		Status:            string(apiKey.Status),
		RequestsPerMinute: requestsPerMinute,
		RequestsPerDay:    requestsPerDay,
		TokenQuota:        tokenQuota,
		TokensUsed:        tokensUsed,
		ExpiresAt:         pgutil.ToPgTimestamptz(apiKey.ExpiresAt),
	})
	if err != nil {
		return nil, err
	}

	apiKey.ID = uint64(k.ID)
	apiKey.CreatedAt = k.CreatedAt.Time
	apiKey.UpdatedAt = k.UpdatedAt.Time

	return apiKey, nil
}

func (r *PostgresRepository) GetByID(
	ctx context.Context,
	id uint64,
) (*APIKey, error) {
	dbID, err := uint64ToInt64("id", id)
	if err != nil {
		return nil, err
	}
	k, err := r.queries.GetAPIKeyByID(ctx, dbID)
	if err != nil {
		return nil, err
	}

	return mapDBAPIKeyToModel(k), nil
}

func (r *PostgresRepository) GetBySelector(
	ctx context.Context,
	selector string,
) (*APIKey, error) {
	k, err := r.queries.GetAPIKeyBySelector(ctx, selector)
	if err != nil {
		return nil, err
	}
	return mapDBAPIKeyToModel(k), nil
}

func (r *PostgresRepository) Update(
	ctx context.Context,
	apiKey *APIKey,
) (*APIKey, error) {
	if apiKey == nil {
		return nil, errors.New("api key cannot be nil")
	}
	if !apiKey.Status.IsValid() {
		return nil, fmt.Errorf("invalid API key status: %q", apiKey.Status)
	}

	id, err := uint64ToInt64("id", apiKey.ID)
	if err != nil {
		return nil, err
	}

	tokenQuota, err := uint64ToInt64("token_quota", apiKey.TokenQuota)
	if err != nil {
		return nil, err
	}

	tokensUsed, err := uint64ToInt64("tokens_used", apiKey.TokensUsed)
	if err != nil {
		return nil, err
	}

	k, err := r.queries.UpdateAPIKey(ctx, db.UpdateAPIKeyParams{
		ID:         id,
		Status:     string(apiKey.Status),
		TokenQuota: tokenQuota,
		TokensUsed: tokensUsed,
		ExpiresAt:  pgutil.ToPgTimestamptz(apiKey.ExpiresAt),
		LastUsedAt: pgutil.ToPgTimestamptz(apiKey.LastUsedAt),
	})
	if err != nil {
		return nil, err
	}

	apiKey.UpdatedAt = k.UpdatedAt.Time

	return apiKey, nil
}

func (r *PostgresRepository) ListByUserID(
	ctx context.Context,
	userID uint64,
) ([]*APIKey, error) {
	dbUserID, err := uint64ToInt64("user_id", userID)
	if err != nil {
		return nil, err
	}
	dbKeys, err := r.queries.ListAPIKeysByUserID(ctx, dbUserID)
	if err != nil {
		return nil, err
	}

	keys := make([]*APIKey, len(dbKeys))
	for i, k := range dbKeys {
		keys[i] = mapDBAPIKeyToModel(k)
	}

	return keys, nil
}

func (r *PostgresRepository) Delete(
	ctx context.Context,
	id uint64,
) error {
	dbID, err := uint64ToInt64("id", id)
	if err != nil {
		return err
	}
	return r.queries.DeleteAPIKey(ctx, dbID)
}

func (r *PostgresRepository) IsUserActive(
	ctx context.Context,
	userID uint64,
) (bool, error) {
	dbUserID, err := uint64ToInt64("user_id", userID)
	if err != nil {
		return false, err
	}
	status, err := r.queries.IsUserActive(ctx, dbUserID)
	if err != nil {
		return false, err
	}

	return status == "active", nil
}

func (r *PostgresRepository) GetUserEmailByID(
	ctx context.Context,
	userID uint64,
) (string, error) {
	dbUserID, err := uint64ToInt64("user_id", userID)
	if err != nil {
		return "", err
	}
	return r.queries.GetUserEmailByID(ctx, dbUserID)
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

func uint64ToInt64(name string, value uint64) (int64, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("%s exceeds PostgreSQL BIGINT range", name)
	}
	return int64(value), nil
}

func uint64ToInt32(name string, value uint64) (int32, error) {
	if value > math.MaxInt32 {
		return 0, fmt.Errorf("%s exceeds PostgreSQL INTEGER range", name)
	}
	return int32(value), nil
}
