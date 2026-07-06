package audit

import (
	"context"
	"log/slog"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/audit/db"
)

type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *PostgresRepository) Create(ctx context.Context, event *AuditEvent) (*AuditEvent, error) {
	var ipAddr *netip.Addr
	if event.IPAddress != nil && *event.IPAddress != "" {
		addr, err := netip.ParseAddr(*event.IPAddress)
		if err == nil {
			ipAddr = &addr
		}
	}

	e, err := r.queries.CreateAuditEvent(ctx, db.CreateAuditEventParams{
		UserID:            toPgInt8(event.UserID),
		ApiKeyID:          toPgInt8(event.APIKeyID),
		AnalysisRequestID: toPgInt8(event.AnalysisRequestID),
		Type:              string(event.Type),
		Description:       event.Description,
		IpAddress:         ipAddr,
		UserAgent:         toPgText(event.UserAgent),
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to record audit event to database", slog.Any("error", err), slog.String("event_type", string(event.Type)))
		return nil, err
	}

	event.ID = uint64(e.ID)
	event.CreatedAt = e.CreatedAt.Time

	slog.InfoContext(ctx, "audit event recorded",
		slog.String("event_type", string(event.Type)),
		slog.String("description", event.Description),
		slog.Any("user_id", event.UserID),
		slog.Any("api_key_id", event.APIKeyID),
	)

	return event, nil
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]*AuditEvent, error) {
	dbEvents, err := r.queries.ListAuditEvents(ctx, db.ListAuditEventsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	events := make([]*AuditEvent, len(dbEvents))
	for i, e := range dbEvents {
		var ipStr *string
		if e.IpAddress != nil {
			s := e.IpAddress.String()
			ipStr = &s
		}

		events[i] = &AuditEvent{
			ID:                uint64(e.ID),
			UserID:            fromPgInt8(e.UserID),
			APIKeyID:          fromPgInt8(e.ApiKeyID),
			AnalysisRequestID: fromPgInt8(e.AnalysisRequestID),
			Type:              AuditEventType(e.Type),
			Description:       e.Description,
			IPAddress:         ipStr,
			UserAgent:         fromPgText(e.UserAgent),
			CreatedAt:         e.CreatedAt.Time,
		}
	}
	return events, nil
}

func toPgInt8(id *uint64) pgtype.Int8 {
	if id == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: int64(*id), Valid: true}
}

func fromPgInt8(i pgtype.Int8) *uint64 {
	if !i.Valid {
		return nil
	}
	val := uint64(i.Int64)
	return &val
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func fromPgText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
