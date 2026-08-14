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
		t.Parallel()

		expectedEvent := &audit.AuditEvent{
			Type:        audit.AuditEventUserLoggedIn,
			Description: "user logged in",
		}

		var receivedEvent *audit.AuditEvent

		repo := &MockRepository{
			CreateFunc: func(
				_ context.Context,
				event *audit.AuditEvent,
			) (*audit.AuditEvent, error) {
				receivedEvent = event
				event.ID = 1
				return event, nil
			},
		}

		svc := audit.NewAuditService(repo, 50)

		err := svc.LogEvent(context.Background(), expectedEvent)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedEvent != expectedEvent {
			t.Fatal("expected the same event to be passed to repository")
		}

		if expectedEvent.ID != 1 {
			t.Fatalf("expected event ID to be updated, got %d", expectedEvent.ID)
		}
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("db error")

		repo := &MockRepository{
			CreateFunc: func(
				_ context.Context,
				_ *audit.AuditEvent,
			) (*audit.AuditEvent, error) {
				return nil, expectedErr
			},
		}

		svc := audit.NewAuditService(repo, 50)

		err := svc.LogEvent(
			context.Background(),
			&audit.AuditEvent{Type: audit.AuditEventUserLoggedIn},
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected original repository error, got %v", err)
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
			name:       "zero limit uses default",
			limit:      0,
			offset:     5,
			wantLimit:  50,
			wantOffset: 5,
		},
		{
			name:       "negative limit uses default",
			limit:      -1,
			offset:     5,
			wantLimit:  50,
			wantOffset: 5,
		},
		{
			name:       "limit above maximum uses default",
			limit:      150,
			offset:     5,
			wantLimit:  50,
			wantOffset: 5,
		},
		{
			name:       "maximum limit is accepted",
			limit:      100,
			offset:     5,
			wantLimit:  100,
			wantOffset: 5,
		},
		{
			name:       "negative offset becomes zero",
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
				ListFunc: func(
					_ context.Context,
					limit, offset int,
				) ([]*audit.AuditEvent, error) {
					if limit != tt.wantLimit {
						t.Errorf("expected limit %d, got %d", tt.wantLimit, limit)
					}

					if offset != tt.wantOffset {
						t.Errorf("expected offset %d, got %d", tt.wantOffset, offset)
					}

					return []*audit.AuditEvent{}, nil
				},
			}

			svc := audit.NewAuditService(repo, 50)

			_, err := svc.ListLogs(
				context.Background(),
				tt.limit,
				tt.offset,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAuditService_ListLogsPropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("list failed")

	repo := &MockRepository{
		ListFunc: func(
			_ context.Context,
			_, _ int,
		) ([]*audit.AuditEvent, error) {
			return nil, expectedErr
		},
	}

	svc := audit.NewAuditService(repo, 50)

	_, err := svc.ListLogs(context.Background(), 10, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected original repository error, got %v", err)
	}
}
