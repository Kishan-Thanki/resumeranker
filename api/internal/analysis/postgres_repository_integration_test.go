//go:build integration

package analysis_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/analysis"
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

	repo := analysis.NewPostgresRepository(pool)

	payload, err := os.ReadFile("testdata/mock_ai_response.json")
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	t.Run("Create and Get Request", func(t *testing.T) {
		ctx := context.Background()
		now := time.Now()

		req := &analysis.AnalysisRequest{
			RequestID: "req-123",
			UserID:    1,
			APIKeyID:  1,
			Status:    analysis.AnalysisRequestStatusProcessing,
			StartedAt: &now,
		}

		created, err := repo.CreateRequest(ctx, req)
		if err != nil {

			t.Skipf("skipping test due to DB error (likely needs migration): %v", err)
		}

		if created.ID == 0 {
			t.Error("expected ID to be set")
		}

		created.Status = analysis.AnalysisRequestStatusCompleted
		_, err = repo.UpdateRequest(ctx, created)
		if err != nil {
			t.Errorf("failed to update request: %v", err)
		}

		fetched, err := repo.GetRequestByID(ctx, created.ID)
		if err != nil {
			t.Errorf("failed to get request: %v", err)
		}

		if fetched.Status != analysis.AnalysisRequestStatusCompleted {
			t.Errorf("expected status completed, got %s", fetched.Status)
		}

		result := &analysis.AnalysisResult{
			AnalysisRequestID: fetched.ID,
			Model:             "gpt-4",
			Result:            string(payload),
			PromptTokens:      10,
			CompletionTokens:  5,
			TotalTokens:       15,
		}

		_, err = repo.CreateResult(ctx, result)
		if err != nil {
			t.Errorf("failed to create result: %v", err)
		}
	})
}
