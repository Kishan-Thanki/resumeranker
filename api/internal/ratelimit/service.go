package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

type Service struct {
	rdb *redis.Client
}

func NewService(redisURL string) (*Service, error) {
	if redisURL == "" {
		return nil, errors.New("redis url is required")
	}

	opt, err := redis.ParseURL("redis://" + redisURL)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Service{rdb: rdb}, nil
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
		return err
	}

	if rpmLimit > 0 && int(rpmIncr.Val()) > rpmLimit {
		return ErrRateLimitExceeded
	}

	if rpdLimit > 0 && int(rpdIncr.Val()) > rpdLimit {
		return ErrRateLimitExceeded
	}

	return nil
}

func (s *Service) GetKeyUsage(ctx context.Context, apiKeyID uint64) (rpmUsed, rpdUsed int, err error) {
	now := time.Now()
	minuteKey := fmt.Sprintf("key:%d:rpm:%d", apiKeyID, now.Unix()/60)
	dayKey := fmt.Sprintf("key:%d:rpd:%s", apiKeyID, now.Format("2006-01-02"))

	vals, err := s.rdb.MGet(ctx, minuteKey, dayKey).Result()
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}

	if len(vals) == 2 {
		if v, ok := vals[0].(string); ok {
			fmt.Sscanf(v, "%d", &rpmUsed)
		}
		if v, ok := vals[1].(string); ok {
			fmt.Sscanf(v, "%d", &rpdUsed)
		}
	}

	return rpmUsed, rpdUsed, nil
}
