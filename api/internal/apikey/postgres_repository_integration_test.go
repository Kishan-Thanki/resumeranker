//go:build integration

package apikey_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
)

func TestPostgresRepository_Integration(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL must be set for integration tests")
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create db pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}

	repo := apikey.NewPostgresRepository(pool)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := "apikey-integration-" + suffix + "@example.com"
	selector := "integration-selector-" + suffix

	var userID uint64
	err = pool.QueryRow(
		ctx,
		`INSERT INTO users (
			email,
			password_hash,
			role,
			status
		) VALUES ($1, $2, $3, $4)
		RETURNING id`,
		email,
		"test-password-hash",
		"user",
		"active",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	})

	key := &apikey.APIKey{
		UserID:            userID,
		Name:              "integration-test-key",
		KeyPrefix:         "rk_test_",
		KeySelector:       selector,
		KeyHash:           "hash_of_verifier_" + suffix,
		Status:            apikey.APIKeyStatusActive,
		RequestsPerMinute: 20,
		RequestsPerDay:    500,
		TokenQuota:        1000,
		TokensUsed:        10,
	}

	t.Run("Create", func(t *testing.T) {
		createdKey, err := repo.Create(ctx, key)
		if err != nil {
			t.Fatalf("failed to create API key: %v", err)
		}

		if createdKey.ID == 0 {
			t.Fatal("expected ID to be set")
		}

		if createdKey.CreatedAt.IsZero() {
			t.Fatal("expected CreatedAt to be set")
		}

		if createdKey.UpdatedAt.IsZero() {
			t.Fatal("expected UpdatedAt to be set")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		fetchedKey, err := repo.GetByID(ctx, key.ID)
		if err != nil {
			t.Fatalf("failed to get API key by ID: %v", err)
		}

		assertAPIKeyEqual(t, key, fetchedKey)
	})

	t.Run("GetBySelector", func(t *testing.T) {
		fetchedKey, err := repo.GetBySelector(ctx, selector)
		if err != nil {
			t.Fatalf("failed to get API key by selector: %v", err)
		}

		assertAPIKeyEqual(t, key, fetchedKey)
	})

	t.Run("ListByUserID", func(t *testing.T) {
		keys, err := repo.ListByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("failed to list API keys: %v", err)
		}

		found := false
		for _, candidate := range keys {
			if candidate.ID == key.ID {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected API key ID %d in user list", key.ID)
		}
	})

	t.Run("Create rejects second active key for same user", func(t *testing.T) {
		secondKey := &apikey.APIKey{
			UserID:            userID,
			Name:              "second-integration-test-key",
			KeyPrefix:         "rk_test_2",
			KeySelector:       selector + "-second",
			KeyHash:           "second-hash_" + suffix,
			Status:            apikey.APIKeyStatusActive,
			RequestsPerMinute: 20,
			RequestsPerDay:    500,
			TokenQuota:        1000,
			TokensUsed:        0,
		}

		_, err := repo.Create(ctx, secondKey)
		if err == nil {
			t.Fatal("expected second active API key creation to fail")
		}

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			t.Fatalf("expected PostgreSQL error, got %T: %v", err, err)
		}

		if pgErr.Code != "23505" {
			t.Fatalf("expected unique_violation (23505), got %s: %s", pgErr.Code, pgErr.Message)
		}

		if pgErr.ConstraintName != "idx_api_keys_one_active_per_user" {
			t.Fatalf(
				"expected constraint %q, got %q",
				"idx_api_keys_one_active_per_user",
				pgErr.ConstraintName,
			)
		}
	})
	t.Run("GetUserEmailByID", func(t *testing.T) {
		gotEmail, err := repo.GetUserEmailByID(ctx, userID)
		if err != nil {
			t.Fatalf("failed to get user email: %v", err)
		}

		if gotEmail != email {
			t.Fatalf("expected email %q, got %q", email, gotEmail)
		}
	})

	t.Run("IsUserActive", func(t *testing.T) {
		active, err := repo.IsUserActive(ctx, userID)
		if err != nil {
			t.Fatalf("failed to check user status: %v", err)
		}

		if !active {
			t.Fatal("expected user to be active")
		}
	})

	t.Run("Update", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		lastUsedAt := time.Now().UTC()

		key.Status = apikey.APIKeyStatusSuspended
		key.TokenQuota = 2000
		key.TokensUsed = 25
		key.ExpiresAt = &expiresAt
		key.LastUsedAt = &lastUsedAt

		updatedKey, err := repo.Update(ctx, key)
		if err != nil {
			t.Fatalf("failed to update API key: %v", err)
		}

		if updatedKey.Status != apikey.APIKeyStatusSuspended {
			t.Fatalf("expected suspended status, got %q", updatedKey.Status)
		}

		if updatedKey.TokenQuota != 2000 {
			t.Fatalf("expected token quota 2000, got %d", updatedKey.TokenQuota)
		}

		if updatedKey.TokensUsed != 25 {
			t.Fatalf("expected tokens used 25, got %d", updatedKey.TokensUsed)
		}

		if updatedKey.ExpiresAt == nil {
			t.Fatal("expected ExpiresAt to be set")
		}

		if updatedKey.LastUsedAt == nil {
			t.Fatal("expected LastUsedAt to be set")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := repo.Delete(ctx, key.ID); err != nil {
			t.Fatalf("failed to delete API key: %v", err)
		}

		_, err := repo.GetByID(ctx, key.ID)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected pgx.ErrNoRows after delete, got %v", err)
		}

		_, err = repo.GetBySelector(ctx, selector)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected pgx.ErrNoRows after delete by selector, got %v", err)
		}
	})
}

func assertAPIKeyEqual(t *testing.T, expected, actual *apikey.APIKey) {
	t.Helper()
	if actual == nil {
		t.Fatal("expected API key, got nil")
	}

	if actual.ID != expected.ID {
		t.Fatalf("expected ID %d, got %d", expected.ID, actual.ID)
	}

	if actual.UserID != expected.UserID {
		t.Fatalf("expected UserID %d, got %d", expected.UserID, actual.UserID)
	}

	if actual.Name != expected.Name {
		t.Fatalf("expected Name %q, got %q", expected.Name, actual.Name)
	}

	if actual.KeyPrefix != expected.KeyPrefix {
		t.Fatalf("expected KeyPrefix %q, got %q", expected.KeyPrefix, actual.KeyPrefix)
	}

	if actual.KeySelector != expected.KeySelector {
		t.Fatalf("expected KeySelector %q, got %q", expected.KeySelector, actual.KeySelector)
	}

	if actual.KeyHash != expected.KeyHash {
		t.Fatalf("expected KeyHash %q, got %q", expected.KeyHash, actual.KeyHash)
	}

	if actual.Status != expected.Status {
		t.Fatalf("expected Status %q, got %q", expected.Status, actual.Status)
	}

	if actual.RequestsPerMinute != expected.RequestsPerMinute {
		t.Fatalf(
			"expected RequestsPerMinute %d, got %d",
			expected.RequestsPerMinute,
			actual.RequestsPerMinute,
		)
	}

	if actual.RequestsPerDay != expected.RequestsPerDay {
		t.Fatalf(
			"expected RequestsPerDay %d, got %d",
			expected.RequestsPerDay,
			actual.RequestsPerDay,
		)
	}

	if actual.TokenQuota != expected.TokenQuota {
		t.Fatalf(
			"expected TokenQuota %d, got %d",
			expected.TokenQuota,
			actual.TokenQuota,
		)
	}

	if actual.TokensUsed != expected.TokensUsed {
		t.Fatalf(
			"expected TokensUsed %d, got %d",
			expected.TokensUsed,
			actual.TokensUsed,
		)
	}
}
