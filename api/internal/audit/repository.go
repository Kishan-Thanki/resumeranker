package audit

import "context"

type Repository interface {
	Create(ctx context.Context, event *AuditEvent) (*AuditEvent, error)
}
