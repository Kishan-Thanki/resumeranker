package audit_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

func TestAuditHandler_ListLogs(t *testing.T) {
	t.Parallel()

	t.Run("success with query params", func(t *testing.T) {
		t.Parallel()

		var gotLimit, gotOffset int

		svc := &MockAuditService{
			ListLogsFunc: func(
				_ context.Context,
				limit, offset int,
			) ([]*audit.AuditEvent, error) {
				gotLimit = limit
				gotOffset = offset

				return []*audit.AuditEvent{
					{
						ID:   1,
						Type: audit.AuditEventUserLoggedIn,
					},
				}, nil
			},
		}

		h := audit.NewAuditHandler(svc, 50)

		req := httptest.NewRequest(
			http.MethodGet,
			"/logs?limit=10&offset=5",
			nil,
		)
		rr := httptest.NewRecorder()

		h.ListLogs(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		if gotLimit != 10 {
			t.Fatalf("expected limit 10, got %d", gotLimit)
		}

		if gotOffset != 5 {
			t.Fatalf("expected offset 5, got %d", gotOffset)
		}

		if got := rr.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", got)
		}

		var logs []*audit.AuditEvent
		if err := json.Unmarshal(rr.Body.Bytes(), &logs); err != nil {
			t.Fatalf("expected valid JSON response: %v", err)
		}

		if len(logs) != 1 {
			t.Fatalf("expected 1 log, got %d", len(logs))
		}

		if logs[0].ID != 1 {
			t.Fatalf("expected log ID 1, got %d", logs[0].ID)
		}

		if logs[0].Type != audit.AuditEventUserLoggedIn {
			t.Fatalf(
				"expected event type %q, got %q",
				audit.AuditEventUserLoggedIn,
				logs[0].Type,
			)
		}
	})

	t.Run("success with default params", func(t *testing.T) {
		t.Parallel()

		var gotLimit, gotOffset int

		svc := &MockAuditService{
			ListLogsFunc: func(
				_ context.Context,
				limit, offset int,
			) ([]*audit.AuditEvent, error) {
				gotLimit = limit
				gotOffset = offset

				return []*audit.AuditEvent{}, nil
			},
		}

		h := audit.NewAuditHandler(svc, 50)

		req := httptest.NewRequest(http.MethodGet, "/logs", nil)
		rr := httptest.NewRecorder()

		h.ListLogs(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		if gotLimit != 50 {
			t.Fatalf("expected limit 50, got %d", gotLimit)
		}

		if gotOffset != 0 {
			t.Fatalf("expected offset 0, got %d", gotOffset)
		}

		if got := rr.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", got)
		}
	})

	t.Run("invalid limit returns bad request", func(t *testing.T) {
		t.Parallel()

		called := false

		svc := &MockAuditService{
			ListLogsFunc: func(
				_ context.Context,
				_, _ int,
			) ([]*audit.AuditEvent, error) {
				called = true
				return nil, nil
			},
		}

		h := audit.NewAuditHandler(svc, 50)

		req := httptest.NewRequest(http.MethodGet, "/logs?limit=abc", nil)
		rr := httptest.NewRecorder()

		h.ListLogs(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}

		if called {
			t.Fatal("expected service not to be called")
		}

		if !strings.Contains(rr.Body.String(), "invalid limit") {
			t.Fatalf("expected invalid limit message, got %q", rr.Body.String())
		}
	})

	t.Run("invalid offset returns bad request", func(t *testing.T) {
		t.Parallel()

		called := false

		svc := &MockAuditService{
			ListLogsFunc: func(
				_ context.Context,
				_, _ int,
			) ([]*audit.AuditEvent, error) {
				called = true
				return nil, nil
			},
		}

		h := audit.NewAuditHandler(svc, 50)

		req := httptest.NewRequest(http.MethodGet, "/logs?offset=abc", nil)
		rr := httptest.NewRecorder()

		h.ListLogs(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}

		if called {
			t.Fatal("expected service not to be called")
		}

		if !strings.Contains(rr.Body.String(), "invalid offset") {
			t.Fatalf("expected invalid offset message, got %q", rr.Body.String())
		}
	})

	t.Run("service error returns internal server error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("service failed")

		svc := &MockAuditService{
			ListLogsFunc: func(
				_ context.Context,
				_, _ int,
			) ([]*audit.AuditEvent, error) {
				return nil, expectedErr
			},
		}

		h := audit.NewAuditHandler(svc, 50)

		req := httptest.NewRequest(http.MethodGet, "/logs", nil)
		rr := httptest.NewRecorder()

		h.ListLogs(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rr.Code)
		}

		if !strings.Contains(
			rr.Body.String(),
			"an internal server error occurred",
		) {
			t.Fatalf("unexpected response body: %q", rr.Body.String())
		}
	})

	t.Run("negative values are passed to service for service-level validation", func(t *testing.T) {
		t.Parallel()

		var gotLimit, gotOffset int

		svc := &MockAuditService{
			ListLogsFunc: func(
				_ context.Context,
				limit, offset int,
			) ([]*audit.AuditEvent, error) {
				gotLimit = limit
				gotOffset = offset
				return []*audit.AuditEvent{}, nil
			},
		}

		h := audit.NewAuditHandler(svc, 50)

		req := httptest.NewRequest(http.MethodGet, "/logs?limit=-1&offset=-5", nil)
		rr := httptest.NewRecorder()

		h.ListLogs(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		if gotLimit != -1 {
			t.Fatalf("expected limit -1, got %d", gotLimit)
		}

		if gotOffset != -5 {
			t.Fatalf("expected offset -5, got %d", gotOffset)
		}
	})
}
