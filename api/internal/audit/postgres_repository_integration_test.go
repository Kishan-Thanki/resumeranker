//go:build integration

package audit_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
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

	repo := audit.NewPostgresRepository(pool)

	t.Run("create and list audit events", func(t *testing.T) {
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

		var userID uint64
		err := pool.QueryRow(
			ctx,
			`INSERT INTO users (
				email,
				password_hash,
				role,
				status
			) VALUES ($1, $2, $3, $4)
			RETURNING id`,
			"audit-integration-"+timestamp+"@example.com",
			"test-password-hash",
			"user",
			"active",
		).Scan(&userID)
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		var apiKeyID uint64
		err = pool.QueryRow(
			ctx,
			`INSERT INTO api_keys (
				user_id,
				name,
				key_selector,
				key_hash,
				status
			) VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			userID,
			"audit integration test key",
			"audit-test-selector-"+timestamp,
			"audit-test-hash-"+timestamp,
			"active",
		).Scan(&apiKeyID)
		if err != nil {
			t.Fatalf("failed to create test API key: %v", err)
		}

		var analysisRequestID uint64
		err = pool.QueryRow(
			ctx,
			`INSERT INTO analysis_requests (
				request_id,
				user_id,
				api_key_id,
				status
			) VALUES ($1, $2, $3, $4)
			RETURNING id`,
			"audit-test-request-"+timestamp,
			userID,
			apiKeyID,
			"completed",
		).Scan(&analysisRequestID)
		if err != nil {
			t.Fatalf("failed to create test analysis request: %v", err)
		}

		ipAddress := "203.0.113.42"
		userAgent := "audit-integration-test"

		firstEvent := &audit.AuditEvent{
			UserID:            &userID,
			APIKeyID:          &apiKeyID,
			AnalysisRequestID: &analysisRequestID,
			Type:              audit.AuditEventAPIKeyUsed,
			Description:       "integration test audit event first",
			IPAddress:         &ipAddress,
			UserAgent:         &userAgent,
		}

		createdFirst, err := repo.Create(ctx, firstEvent)
		if err != nil {
			t.Fatalf("failed to create first audit event: %v", err)
		}

		if createdFirst.ID == 0 {
			t.Fatal("expected first event ID to be set")
		}

		if createdFirst.CreatedAt.IsZero() {
			t.Fatal("expected first event CreatedAt to be set")
		}

		if createdFirst.IPAddress == nil || *createdFirst.IPAddress != ipAddress {
			t.Fatalf(
				"expected IP address %q, got %v",
				ipAddress,
				createdFirst.IPAddress,
			)
		}

		if createdFirst.UserAgent == nil || *createdFirst.UserAgent != userAgent {
			t.Fatalf(
				"expected user agent %q, got %v",
				userAgent,
				createdFirst.UserAgent,
			)
		}

		if createdFirst.UserID == nil || *createdFirst.UserID != userID {
			t.Fatalf(
				"expected user ID %d, got %v",
				userID,
				createdFirst.UserID,
			)
		}

		if createdFirst.APIKeyID == nil || *createdFirst.APIKeyID != apiKeyID {
			t.Fatalf(
				"expected API key ID %d, got %v",
				apiKeyID,
				createdFirst.APIKeyID,
			)
		}

		if createdFirst.AnalysisRequestID == nil ||
			*createdFirst.AnalysisRequestID != analysisRequestID {
			t.Fatalf(
				"expected analysis request ID %d, got %v",
				analysisRequestID,
				createdFirst.AnalysisRequestID,
			)
		}

		secondEvent := &audit.AuditEvent{
			Type:        audit.AuditEventUserLoggedIn,
			Description: "integration test audit event second",
		}

		createdSecond, err := repo.Create(ctx, secondEvent)
		if err != nil {
			t.Fatalf("failed to create second audit event: %v", err)
		}

		if createdSecond.ID == 0 {
			t.Fatal("expected second event ID to be set")
		}

		if createdSecond.CreatedAt.IsZero() {
			t.Fatal("expected second event CreatedAt to be set")
		}

		events, err := repo.List(ctx, 100, 0)
		if err != nil {
			t.Fatalf("failed to list audit events: %v", err)
		}

		firstIndex := -1
		secondIndex := -1

		for i, event := range events {
			switch event.ID {
			case createdFirst.ID:
				firstIndex = i
			case createdSecond.ID:
				secondIndex = i
			}
		}

		if firstIndex == -1 {
			t.Fatalf("expected to find first audit event ID %d", createdFirst.ID)
		}

		if secondIndex == -1 {
			t.Fatalf("expected to find second audit event ID %d", createdSecond.ID)
		}

		if secondIndex >= firstIndex {
			t.Fatalf(
				"expected newer second event at an earlier list position: firstIndex=%d secondIndex=%d",
				firstIndex,
				secondIndex,
			)
		}

		if events[firstIndex].Type != firstEvent.Type {
			t.Fatalf(
				"expected first event type %q, got %q",
				firstEvent.Type,
				events[firstIndex].Type,
			)
		}

		if events[firstIndex].Description != firstEvent.Description {
			t.Fatalf(
				"expected first description %q, got %q",
				firstEvent.Description,
				events[firstIndex].Description,
			)
		}

		if events[secondIndex].Type != secondEvent.Type {
			t.Fatalf(
				"expected second event type %q, got %q",
				secondEvent.Type,
				events[secondIndex].Type,
			)
		}

		if events[secondIndex].Description != secondEvent.Description {
			t.Fatalf(
				"expected second description %q, got %q",
				secondEvent.Description,
				events[secondIndex].Description,
			)
		}
	})
}
