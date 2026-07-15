package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

type Service struct {
	rdb *redis.Client
}

func NewService(rdb *redis.Client) *Service {
	return &Service{rdb: rdb}
}

func (s *Service) Close() error {
	if s.rdb != nil {
		return s.rdb.Close()
	}
	return nil
}

func (s *Service) CheckGlobalLimit(ctx context.Context, rpmLimit, rpdLimit int) error {
	now := time.Now()
	minuteKey := fmt.Sprintf("global:rpm:%d", now.Unix()/60)
	dayKey := fmt.Sprintf("global:rpd:%s", now.Format("2006-01-02"))

	return s.checkLimits(ctx, minuteKey, dayKey, rpmLimit, rpdLimit)
}

func (s *Service) CheckKeyLimit(ctx context.Context, apiKeyID uint64, rpmLimit, rpdLimit int) error {
	now := time.Now()
	minuteKey := fmt.Sprintf("key:%d:rpm:%d", apiKeyID, now.Unix()/60)
	dayKey := fmt.Sprintf("key:%d:rpd:%s", apiKeyID, now.Format("2006-01-02"))

	return s.checkLimits(ctx, minuteKey, dayKey, rpmLimit, rpdLimit)
}

func (s *Service) checkLimits(ctx context.Context, minuteKey, dayKey string, rpmLimit, rpdLimit int) error {
	pipe := s.rdb.TxPipeline()

	rpmIncr := pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, time.Minute*2)

	rpdIncr := pipe.Incr(ctx, dayKey)
	pipe.Expire(ctx, dayKey, time.Hour*48)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute redis rate limit pipeline: %w", err)
	}

	if rpmLimit > 0 && int(rpmIncr.Val()) > rpmLimit {
		now := time.Now()
		nextMin := now.Truncate(time.Minute).Add(time.Minute)
		return &RateLimitError{
			RetryAfter: time.Until(nextMin),
			Message:    "per-minute rate limit exceeded",
		}
	}

	if rpdLimit > 0 && int(rpdIncr.Val()) > rpdLimit {
		now := time.Now()
		nextMidnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
		return &RateLimitError{
			RetryAfter: time.Until(nextMidnight),
			Message:    "daily rate limit exceeded",
		}
	}

	return nil
}

func (s *Service) GetKeyUsage(ctx context.Context, apiKeyID uint64) (rpmUsed, rpdUsed int, err error) {
	now := time.Now()
	minuteKey := fmt.Sprintf("key:%d:rpm:%d", apiKeyID, now.Unix()/60)
	dayKey := fmt.Sprintf("key:%d:rpd:%s", apiKeyID, now.Format("2006-01-02"))

	vals, err := s.rdb.MGet(ctx, minuteKey, dayKey).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get rate limit usage from redis: %w", err)
	}

	if len(vals) == 2 {
		if v, ok := vals[0].(string); ok {
			rpmUsed, _ = strconv.Atoi(v)
		}
		if v, ok := vals[1].(string); ok {
			rpdUsed, _ = strconv.Atoi(v)
		}
	}

	return rpmUsed, rpdUsed, nil
}
