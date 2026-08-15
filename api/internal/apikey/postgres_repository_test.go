package apikey

import (
	"math"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey/db"
)

func TestUint64ToInt64(t *testing.T) {
	t.Run("zero is valid", func(t *testing.T) {
		got, err := uint64ToInt64("id", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
	})

	t.Run("max int64 is valid", func(t *testing.T) {
		got, err := uint64ToInt64("id", uint64(math.MaxInt64))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != math.MaxInt64 {
			t.Fatalf("expected %d, got %d", math.MaxInt64, got)
		}
	})

	t.Run("above max int64 is rejected", func(t *testing.T) {
		_, err := uint64ToInt64(
			"token_quota",
			uint64(math.MaxInt64)+1,
		)
		if err == nil {
			t.Fatal("expected error")
		}

		expected := "token_quota exceeds PostgreSQL BIGINT range"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestUint64ToInt32(t *testing.T) {
	t.Run("zero is valid", func(t *testing.T) {
		got, err := uint64ToInt32("requests_per_day", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
	})

	t.Run("max int32 is valid", func(t *testing.T) {
		got, err := uint64ToInt32(
			"requests_per_minute",
			uint64(math.MaxInt32),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != math.MaxInt32 {
			t.Fatalf("expected %d, got %d", math.MaxInt32, got)
		}
	})

	t.Run("above max int32 is rejected", func(t *testing.T) {
		_, err := uint64ToInt32(
			"requests_per_day",
			uint64(math.MaxInt32)+1,
		)
		if err == nil {
			t.Fatal("expected error")
		}

		expected := "requests_per_day exceeds PostgreSQL INTEGER range"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestMapDBAPIKeyToModel(t *testing.T) {
	expiresAt := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	lastUsedAt := time.Date(2026, 8, 20, 11, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2026, 8, 20, 13, 0, 0, 0, time.UTC)

	got := mapDBAPIKeyToModel(db.ApiKey{
		ID:                101,
		UserID:            202,
		Name:              "Test Key",
		KeySelector:       "selector",
		KeyHash:           "hash",
		Status:            string(APIKeyStatusSuspended),
		TokenQuota:        5000,
		TokensUsed:        1234,
		ExpiresAt:         pgtype.Timestamptz{Time: expiresAt, Valid: true},
		LastUsedAt:        pgtype.Timestamptz{Time: lastUsedAt, Valid: true},
		CreatedAt:         pgtype.Timestamptz{Time: createdAt, Valid: true},
		UpdatedAt:         pgtype.Timestamptz{Time: updatedAt, Valid: true},
		DeletedAt:         pgtype.Timestamptz{Time: deletedAt, Valid: true},
		KeyPrefix:         "rk_live_",
		RequestsPerMinute: 20,
		RequestsPerDay:    500,
	})

	if got.ID != 101 {
		t.Fatalf("expected ID 101, got %d", got.ID)
	}
	if got.UserID != 202 {
		t.Fatalf("expected UserID 202, got %d", got.UserID)
	}
	if got.Name != "Test Key" {
		t.Fatalf("expected name %q, got %q", "Test Key", got.Name)
	}
	if got.KeyPrefix != "rk_live_" {
		t.Fatalf("expected key prefix %q, got %q", "rk_live_", got.KeyPrefix)
	}
	if got.KeySelector != "selector" {
		t.Fatalf("expected key selector %q, got %q", "selector", got.KeySelector)
	}
	if got.KeyHash != "hash" {
		t.Fatalf("expected key hash %q, got %q", "hash", got.KeyHash)
	}
	if got.Status != APIKeyStatusSuspended {
		t.Fatalf("expected status %q, got %q", APIKeyStatusSuspended, got.Status)
	}
	if got.RequestsPerMinute != 20 {
		t.Fatalf("expected RPM 20, got %d", got.RequestsPerMinute)
	}
	if got.RequestsPerDay != 500 {
		t.Fatalf("expected RPD 500, got %d", got.RequestsPerDay)
	}
	if got.TokenQuota != 5000 {
		t.Fatalf("expected token quota 5000, got %d", got.TokenQuota)
	}
	if got.TokensUsed != 1234 {
		t.Fatalf("expected tokens used 1234, got %d", got.TokensUsed)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected ExpiresAt: %v", got.ExpiresAt)
	}
	if got.LastUsedAt == nil || !got.LastUsedAt.Equal(lastUsedAt) {
		t.Fatalf("unexpected LastUsedAt: %v", got.LastUsedAt)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt %v, got %v", createdAt, got.CreatedAt)
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected UpdatedAt %v, got %v", updatedAt, got.UpdatedAt)
	}
	if got.DeletedAt == nil || !got.DeletedAt.Equal(deletedAt) {
		t.Fatalf("unexpected DeletedAt: %v", got.DeletedAt)
	}
}

func TestMapDBAPIKeyToModelNullableFields(t *testing.T) {
	got := mapDBAPIKeyToModel(db.ApiKey{
		ID:          1,
		UserID:      2,
		Name:        "Test Key",
		KeySelector: "selector",
		KeyHash:     "hash",
		Status:      string(APIKeyStatusActive),
		ExpiresAt:   pgtype.Timestamptz{Valid: false},
		LastUsedAt:  pgtype.Timestamptz{Valid: false},
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Unix(0, 0).UTC(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Unix(0, 0).UTC(),
			Valid: true,
		},
		DeletedAt: pgtype.Timestamptz{Valid: false},
	})

	if got.ExpiresAt != nil {
		t.Fatalf("expected nil ExpiresAt, got %v", got.ExpiresAt)
	}
	if got.LastUsedAt != nil {
		t.Fatalf("expected nil LastUsedAt, got %v", got.LastUsedAt)
	}
	if got.DeletedAt != nil {
		t.Fatalf("expected nil DeletedAt, got %v", got.DeletedAt)
	}
}

func TestPostgresRepositoryCreateValidation(t *testing.T) {
	repo := &PostgresRepository{}

	t.Run("nil API key", func(t *testing.T) {
		_, err := repo.Create(nil, nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		_, err := repo.Create(nil, &APIKey{
			UserID: 1,
			Status: "invalid",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
