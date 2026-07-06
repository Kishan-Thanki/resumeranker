package audit

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, event *AuditEvent) (*AuditEvent, error) {
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
		slog.ErrorContext(ctx, "failed to record audit event to database", slog.Any("error", err), slog.String("event_type", string(event.Type)))
		return nil, err
	}
	
	slog.InfoContext(ctx, "audit event recorded", 
		slog.String("event_type", string(event.Type)),
		slog.String("description", event.Description),
		slog.Any("user_id", event.UserID),
		slog.Any("api_key_id", event.APIKeyID),
	)
	
	return event, nil
}
