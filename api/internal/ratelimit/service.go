package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/ratelimit/db"
)

// RateLimitError is returned when a rate limit is exceeded.
// RetryAfter tells the caller how long to wait before retrying.
type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// Service implements rate limiting using Postgres INSERT ON CONFLICT DO UPDATE.
// This replaces the former Redis INCR/EXPIRE implementation with zero new
// infrastructure — the existing Postgres pool is reused.
type Service struct {
	queries rateLimitStore
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{queries: db.New(pool)}
}

// CheckGlobalLimit checks the global (system-wide) RPM and RPD limits.
func (s *Service) CheckGlobalLimit(ctx context.Context, rpmLimit, rpdLimit int) error {
	now := time.Now()
	minuteKey := fmt.Sprintf("global:rpm:%d", now.Unix()/60)
	dayKey := fmt.Sprintf("global:rpd:%s", now.Format("2006-01-02"))

	return s.checkLimits(ctx, minuteKey, dayKey, rpmLimit, rpdLimit, now)
}

// CheckKeyLimit checks RPM and RPD limits for a specific API key.
func (s *Service) CheckKeyLimit(ctx context.Context, apiKeyID uint64, rpmLimit, rpdLimit int) error {
	now := time.Now()
	minuteKey := fmt.Sprintf("key:%d:rpm:%d", apiKeyID, now.Unix()/60)
	dayKey := fmt.Sprintf("key:%d:rpd:%s", apiKeyID, now.Format("2006-01-02"))

	return s.checkLimits(ctx, minuteKey, dayKey, rpmLimit, rpdLimit, now)
}

// checkLimits increments both counters independently and checks their
// thresholds. Each individual counter update is atomic at the PostgreSQL
// row level via INSERT ... ON CONFLICT DO UPDATE.
func (s *Service) checkLimits(
	ctx context.Context,
	minuteKey,
	dayKey string,
	rpmLimit,
	rpdLimit int,
	now time.Time,
) error {
	minuteExpiry := pgtype.Timestamptz{
		Time:  now.Truncate(time.Minute).Add(2 * time.Minute),
		Valid: true,
	}

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		now.Location(),
	)
	tomorrow := today.AddDate(0, 0, 1)

	dayExpiry := pgtype.Timestamptz{
		Time:  tomorrow,
		Valid: true,
	}

	rpmCount, err := s.queries.UpsertRateLimitCounter(
		ctx,
		db.UpsertRateLimitCounterParams{
			Key:       minuteKey,
			ExpiresAt: minuteExpiry,
		},
	)
	if err != nil {
		return fmt.Errorf("ratelimit: failed to upsert rpm counter: %w", err)
	}

	rpdCount, err := s.queries.UpsertRateLimitCounter(
		ctx,
		db.UpsertRateLimitCounterParams{
			Key:       dayKey,
			ExpiresAt: dayExpiry,
		},
	)
	if err != nil {
		return fmt.Errorf("ratelimit: failed to upsert rpd counter: %w", err)
	}

	if rpmLimit > 0 && int(rpmCount) > rpmLimit {
		nextMinute := now.Truncate(time.Minute).Add(time.Minute)

		return &RateLimitError{
			RetryAfter: time.Until(nextMinute),
			Message:    "per-minute rate limit exceeded",
		}
	}

	if rpdLimit > 0 && int(rpdCount) > rpdLimit {
		return &RateLimitError{
			RetryAfter: time.Until(tomorrow),
			Message:    "daily rate limit exceeded",
		}
	}

	return nil
}

// GetKeyUsage returns the current rpm and rpd usage for a given API key.
// Used by the API key stats endpoint. Returns 0,0 if counters don't exist yet.
func (s *Service) GetKeyUsage(
	ctx context.Context,
	apiKeyID uint64,
) (rpmUsed, rpdUsed int, err error) {
	now := time.Now()
	minuteKey := fmt.Sprintf("key:%d:rpm:%d", apiKeyID, now.Unix()/60)
	dayKey := fmt.Sprintf("key:%d:rpd:%s", apiKeyID, now.Format("2006-01-02"))

	rows, err := s.queries.GetRateLimitCounters(
		ctx,
		[]string{minuteKey, dayKey},
	)
	if err != nil {
		return 0, 0, fmt.Errorf("ratelimit: failed to get key usage: %w", err)
	}

	for _, row := range rows {
		switch row.Key {
		case minuteKey:
			rpmUsed = int(row.Count)
		case dayKey:
			rpdUsed = int(row.Count)
		}
	}

	return rpmUsed, rpdUsed, nil
}

// CleanupExpired removes expired counter rows. Call this periodically
// (e.g. once per minute from a goroutine) to keep the table small.
func (s *Service) CleanupExpired(ctx context.Context) error {
	return s.queries.DeleteExpiredRateLimitCounters(ctx)
}

type rateLimitStore interface {
	UpsertRateLimitCounter(
		ctx context.Context,
		arg db.UpsertRateLimitCounterParams,
	) (int32, error)

	GetRateLimitCounters(
		ctx context.Context,
		keys []string,
	) ([]db.GetRateLimitCountersRow, error)

	DeleteExpiredRateLimitCounters(ctx context.Context) error
}
