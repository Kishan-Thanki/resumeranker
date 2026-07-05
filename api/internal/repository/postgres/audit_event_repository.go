package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/domain"
)

type AuditEventRepository struct {
	db *pgxpool.Pool
}

func NewAuditEventRepository(db *pgxpool.Pool) *AuditEventRepository {
	return &AuditEventRepository{db: db}
}

func (r *AuditEventRepository) Create(ctx context.Context, event *domain.AuditEvent) (*domain.AuditEvent, error) {
	const sql = `
		INSERT INTO audit_events (user_id, api_key_id, analysis_request_id, type, description, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, sql,
		event.UserID,
		event.APIKeyID,
		event.AnalysisRequestID,
		event.Type,
		event.Description,
		event.IPAddress,
		event.UserAgent,
	).Scan(
		&event.ID,
		&event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}
