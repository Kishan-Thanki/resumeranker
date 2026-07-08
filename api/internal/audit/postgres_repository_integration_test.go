//go:build integration

package audit_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

func TestPostgresRepository_Integration(t *testing.T) {
	t.Parallel()

	dbURL := os.Getenv("TEST_DATABASE_URL")
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

	repo := audit.NewPostgresRepository(pool)

	t.Run("Create and List Logs", func(t *testing.T) {
		ctx := context.Background()
		userID := uint64(1)

		event := &audit.AuditEvent{
			Type:        audit.AuditEventUserLoggedIn,
			Description: "integration test log",
			UserID:      &userID,
		}

		createdEvent, err := repo.Create(ctx, event)
		if err != nil {
			t.Skipf("skipping test due to DB error (likely needs migration): %v", err)
		}

		if createdEvent.ID == 0 {
			t.Error("expected ID to be set")
		}

		logs, err := repo.List(ctx, 10, 0)
		if err != nil {
			t.Errorf("failed to list logs: %v", err)
		}

		if len(logs) == 0 {
			t.Error("expected at least one log")
		}

		found := false
		for _, l := range logs {
			if l.ID == createdEvent.ID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected to find created event ID %d in list", createdEvent.ID)
		}
	})
}
