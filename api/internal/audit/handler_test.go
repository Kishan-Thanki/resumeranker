package audit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

func TestAuditHandler_ListLogs(t *testing.T) {
	t.Parallel()

	t.Run("success with query params", func(t *testing.T) {
		svc := &MockAuditService{
			ListLogsFunc: func(ctx context.Context, limit, offset int) ([]*audit.AuditEvent, error) {
				if limit != 10 {
					t.Errorf("expected limit 10, got %d", limit)
				}
				if offset != 5 {
					t.Errorf("expected offset 5, got %d", offset)
				}
				return []*audit.AuditEvent{
					{ID: 1, Type: audit.AuditEventUserLoggedIn},
				}, nil
			},
		}
		h := audit.NewAuditHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/logs?limit=10&offset=5", nil)
		rr := httptest.NewRecorder()

		h.ListLogs(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("success with default params", func(t *testing.T) {
		svc := &MockAuditService{
			ListLogsFunc: func(ctx context.Context, limit, offset int) ([]*audit.AuditEvent, error) {
				if limit != 0 {
					t.Errorf("expected limit 0, got %d", limit)
				}
				if offset != 0 {
					t.Errorf("expected offset 0, got %d", offset)
				}
				return []*audit.AuditEvent{}, nil
			},
		}
		h := audit.NewAuditHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/logs", nil)
		rr := httptest.NewRecorder()

		h.ListLogs(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	})
}
