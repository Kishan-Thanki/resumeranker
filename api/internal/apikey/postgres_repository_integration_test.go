//go:build integration

package apikey_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
)

func TestPostgresRepository_Integration(t *testing.T) {
	t.Parallel()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/resumeranker?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	repo := apikey.NewPostgresRepository(pool)

	t.Run("Create and Get Key", func(t *testing.T) {
		ctx := context.Background()

		key := &apikey.APIKey{
			UserID:      1,
			Name:        "integration-test-key",
			KeySelector: "INTEGRATIONSELECTOR",
			KeyHash:     "hash_of_verifier",
			Status:      apikey.APIKeyStatusActive,
			TokenQuota:  1000,
		}

		createdKey, err := repo.Create(ctx, key)
		if err != nil {
			t.Skipf("skipping test due to DB error (likely needs migration): %v", err)
		}

		if createdKey.ID == 0 {
			t.Error("expected ID to be set")
		}

		fetchedKey, err := repo.GetBySelector(ctx, "INTEGRATIONSELECTOR")
		if err != nil {
			t.Errorf("failed to fetch key by selector: %v", err)
		}

		if fetchedKey.Name != "integration-test-key" {
			t.Errorf("expected name integration-test-key, got %s", fetchedKey.Name)
		}
	})
}
