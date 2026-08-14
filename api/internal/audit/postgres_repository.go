package audit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/netip"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/audit/db"
	"github.com/kishan-thanki/resumeranker/api/internal/pgutil"
)

type PostgresRepository struct {
	queries *db.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		queries: db.New(pool),
	}
}

func (r *PostgresRepository) Create(ctx context.Context, event *AuditEvent) (*AuditEvent, error) {
	if event == nil {
		return nil, errors.New("audit event cannot be nil")
	}

	if !event.Type.IsValid() {
		return nil, fmt.Errorf("invalid audit event type: %q", event.Type)
	}

	if err := validateID("user_id", event.UserID); err != nil {
		return nil, err
	}

	if err := validateID("api_key_id", event.APIKeyID); err != nil {
		return nil, err
	}

	if err := validateID("analysis_request_id", event.AnalysisRequestID); err != nil {
		return nil, err
	}

	var ipAddr *netip.Addr
	if event.IPAddress != nil && *event.IPAddress != "" {
		addr, err := netip.ParseAddr(*event.IPAddress)
		if err != nil {
			return nil, fmt.Errorf("invalid audit event IP address: %w", err)
		}
		ipAddr = &addr
	}

	e, err := r.queries.CreateAuditEvent(ctx, db.CreateAuditEventParams{
		UserID:            pgutil.ToPgInt8(event.UserID),
		ApiKeyID:          pgutil.ToPgInt8(event.APIKeyID),
		AnalysisRequestID: pgutil.ToPgInt8(event.AnalysisRequestID),
		Type:              string(event.Type),
		Description:       event.Description,
		IpAddress:         ipAddr,
		UserAgent:         pgutil.ToPgText(event.UserAgent),
	})
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to record audit event to database",
			slog.Any("error", err),
			slog.String("event_type", string(event.Type)),
		)
		return nil, err
	}

	event.ID = uint64(e.ID)
	event.CreatedAt = e.CreatedAt.Time

	slog.InfoContext(
		ctx,
		"audit event recorded",
		slog.Any("event", event),
		slog.Any("analysis_request_id", event.AnalysisRequestID),
		slog.String("event_type", string(event.Type)),
		slog.String("description", event.Description),
		slog.Any("ip_address", event.IPAddress),
		slog.Any("user_agent", event.UserAgent),
	)

	slog.DebugContext(
		ctx,
		"audit event recorded",
		slog.Any("event_id", event.ID),
		slog.Any("user_id", event.UserID),
		slog.Any("api_key_id", event.APIKeyID),
		slog.Any("analysis_request_id", event.AnalysisRequestID),
		slog.String("event_type", string(event.Type)),
		slog.String("description", event.Description),
		slog.Any("ip_address", event.IPAddress),
		slog.Any("user_agent", event.UserAgent),
	)

	return event, nil
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]*AuditEvent, error) {
	if offset < 0 {
		offset = 0
	}

	dbEvents, err := r.queries.ListAuditEvents(ctx, db.ListAuditEventsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	events := make([]*AuditEvent, len(dbEvents))
	for i, e := range dbEvents {
		events[i] = eventFromRow(e)
	}

	return events, nil
}

func eventFromRow(e db.AuditEvent) *AuditEvent {
	var ipStr *string
	if e.IpAddress != nil {
		s := e.IpAddress.String()
		ipStr = &s
	}

	return &AuditEvent{
		ID:                uint64(e.ID),
		UserID:            pgutil.FromPgInt8(e.UserID),
		APIKeyID:          pgutil.FromPgInt8(e.ApiKeyID),
		AnalysisRequestID: pgutil.FromPgInt8(e.AnalysisRequestID),
		Type:              AuditEventType(e.Type),
		Description:       e.Description,
		IPAddress:         ipStr,
		UserAgent:         pgutil.FromPgText(e.UserAgent),
		CreatedAt:         e.CreatedAt.Time,
	}
}

func validateID(name string, id *uint64) error {
	if id == nil {
		return nil
	}

	if *id > math.MaxInt64 {
		return fmt.Errorf("%s exceeds PostgreSQL BIGINT range", name)
	}

	return nil
}
