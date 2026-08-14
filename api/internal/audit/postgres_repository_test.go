package audit

import (
	"math"
	"net/netip"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kishan-thanki/resumeranker/api/internal/audit/db"
)

func TestValidateID(t *testing.T) {
	t.Run("nil ID is valid", func(t *testing.T) {
		if err := validateID("user_id", nil); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("zero is valid", func(t *testing.T) {
		id := uint64(0)

		if err := validateID("user_id", &id); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("max int64 is valid", func(t *testing.T) {
		id := uint64(math.MaxInt64)

		if err := validateID("user_id", &id); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("above max int64 is rejected", func(t *testing.T) {
		id := uint64(math.MaxInt64) + 1

		err := validateID("user_id", &id)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		expected := "user_id exceeds PostgreSQL BIGINT range"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error uses supplied field name", func(t *testing.T) {
		id := uint64(math.MaxInt64) + 1

		err := validateID("analysis_request_id", &id)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		expected := "analysis_request_id exceeds PostgreSQL BIGINT range"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestEventFromRow(t *testing.T) {
	createdAt := time.Date(
		2026,
		time.August,
		14,
		23,
		0,
		0,
		0,
		time.FixedZone("IST", 5*60*60+30*60),
	)

	userID := int64(101)
	apiKeyID := int64(202)
	analysisRequestID := int64(303)

	ipAddress := netip.MustParseAddr("203.0.113.42")

	got := eventFromRow(db.AuditEvent{
		ID: int64(404),
		UserID: pgtype.Int8{
			Int64: userID,
			Valid: true,
		},
		ApiKeyID: pgtype.Int8{
			Int64: apiKeyID,
			Valid: true,
		},
		AnalysisRequestID: pgtype.Int8{
			Int64: analysisRequestID,
			Valid: true,
		},
		Type:        string(AuditEventAPIKeyUsed),
		Description: "API key used",
		IpAddress:   &ipAddress,
		UserAgent: pgtype.Text{
			String: "Mozilla/5.0",
			Valid:  true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
	})

	if got == nil {
		t.Fatal("expected non-nil event")
	}

	if got.ID != 404 {
		t.Fatalf("expected ID 404, got %d", got.ID)
	}

	if got.UserID == nil || *got.UserID != uint64(userID) {
		t.Fatalf("expected user ID %d, got %v", userID, got.UserID)
	}

	if got.APIKeyID == nil || *got.APIKeyID != uint64(apiKeyID) {
		t.Fatalf("expected API key ID %d, got %v", apiKeyID, got.APIKeyID)
	}

	if got.AnalysisRequestID == nil ||
		*got.AnalysisRequestID != uint64(analysisRequestID) {
		t.Fatalf(
			"expected analysis request ID %d, got %v",
			analysisRequestID,
			got.AnalysisRequestID,
		)
	}

	if got.Type != AuditEventAPIKeyUsed {
		t.Fatalf("expected type %q, got %q", AuditEventAPIKeyUsed, got.Type)
	}

	if got.Description != "API key used" {
		t.Fatalf("expected description %q, got %q", "API key used", got.Description)
	}

	if got.IPAddress == nil || *got.IPAddress != "203.0.113.42" {
		t.Fatalf("expected IP address %q, got %v", "203.0.113.42", got.IPAddress)
	}

	if got.UserAgent == nil || *got.UserAgent != "Mozilla/5.0" {
		t.Fatalf("expected user agent %q, got %v", "Mozilla/5.0", got.UserAgent)
	}

	if !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt %v, got %v", createdAt, got.CreatedAt)
	}
}

func TestEventFromRowWithNullableFields(t *testing.T) {
	got := eventFromRow(db.AuditEvent{
		ID:                1,
		UserID:            pgtype.Int8{Valid: false},
		ApiKeyID:          pgtype.Int8{Valid: false},
		AnalysisRequestID: pgtype.Int8{Valid: false},
		Type:              string(AuditEventUserLoggedIn),
		Description:       "User logged in",
		IpAddress:         nil,
		UserAgent:         pgtype.Text{Valid: false},
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Unix(0, 0).UTC(),
			Valid: true,
		},
	})

	if got == nil {
		t.Fatal("expected non-nil event")
	}

	if got.UserID != nil {
		t.Fatalf("expected nil UserID, got %v", got.UserID)
	}

	if got.APIKeyID != nil {
		t.Fatalf("expected nil APIKeyID, got %v", got.APIKeyID)
	}

	if got.AnalysisRequestID != nil {
		t.Fatalf(
			"expected nil AnalysisRequestID, got %v",
			got.AnalysisRequestID,
		)
	}

	if got.IPAddress != nil {
		t.Fatalf("expected nil IPAddress, got %v", got.IPAddress)
	}

	if got.UserAgent != nil {
		t.Fatalf("expected nil UserAgent, got %v", got.UserAgent)
	}
}
