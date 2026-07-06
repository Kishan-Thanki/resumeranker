package audit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

func TestAuditService_LogEvent(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo := &MockRepository{
			CreateFunc: func(ctx context.Context, event *audit.AuditEvent) (*audit.AuditEvent, error) {
				event.ID = 1
				return event, nil
			},
		}
		svc := audit.NewAuditService(repo)

		err := svc.LogEvent(context.Background(), &audit.AuditEvent{Type: audit.AuditEventUserLoggedIn})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &MockRepository{
			CreateFunc: func(ctx context.Context, event *audit.AuditEvent) (*audit.AuditEvent, error) {
				return nil, errors.New("db error")
			},
		}
		svc := audit.NewAuditService(repo)

		err := svc.LogEvent(context.Background(), &audit.AuditEvent{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestAuditService_ListLogs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{
			name:       "valid pagination",
			limit:      10,
			offset:     5,
			wantLimit:  10,
			wantOffset: 5,
		},
		{
			name:       "limit too small",
			limit:      0,
			offset:     5,
			wantLimit:  50,
			wantOffset: 5,
		},
		{
			name:       "limit too large",
			limit:      150,
			offset:     5,
			wantLimit:  50,
			wantOffset: 5,
		},
		{
			name:       "negative offset",
			limit:      10,
			offset:     -5,
			wantLimit:  10,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &MockRepository{
				ListFunc: func(ctx context.Context, limit, offset int) ([]*audit.AuditEvent, error) {
					if limit != tt.wantLimit {
						t.Errorf("expected limit %d, got %d", tt.wantLimit, limit)
					}
					if offset != tt.wantOffset {
						t.Errorf("expected offset %d, got %d", tt.wantOffset, offset)
					}
					return []*audit.AuditEvent{}, nil
				},
			}
			svc := audit.NewAuditService(repo)

			_, err := svc.ListLogs(context.Background(), tt.limit, tt.offset)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
