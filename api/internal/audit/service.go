package audit

import (
	"context"
)

type AuditService struct {
	repo Repository
}

func NewAuditService(repo Repository) *AuditService {
	return &AuditService{
		repo: repo,
	}
}

func (s *AuditService) LogEvent(ctx context.Context, event *AuditEvent) error {
	_, err := s.repo.Create(ctx, event)
	return err
}

func (s *AuditService) ListLogs(ctx context.Context, limit, offset int) ([]*AuditEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}
