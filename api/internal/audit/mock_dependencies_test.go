package audit_test

import (
	"context"

	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

type MockRepository struct {
	CreateFunc func(ctx context.Context, event *audit.AuditEvent) (*audit.AuditEvent, error)
	ListFunc   func(ctx context.Context, limit, offset int) ([]*audit.AuditEvent, error)
}

func (m *MockRepository) Create(ctx context.Context, event *audit.AuditEvent) (*audit.AuditEvent, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, event)
	}
	return event, nil
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*audit.AuditEvent, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return nil, nil
}

type MockAuditService struct {
	ListLogsFunc func(ctx context.Context, limit, offset int) ([]*audit.AuditEvent, error)
}

func (m *MockAuditService) ListLogs(ctx context.Context, limit, offset int) ([]*audit.AuditEvent, error) {
	if m.ListLogsFunc != nil {
		return m.ListLogsFunc(ctx, limit, offset)
	}
	return nil, nil
}
