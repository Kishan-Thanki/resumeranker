package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/ratelimit/db"
)

type fakeRateLimitStore struct {
	upsertCounts []int32
	upsertErrs   []error
	usageRows    []db.GetRateLimitCountersRow
	usageErr     error
	cleanupErr   error

	upsertCalls []db.UpsertRateLimitCounterParams
	usageKeys   []string
	cleanupCall bool
}

func (f *fakeRateLimitStore) UpsertRateLimitCounter(
	_ context.Context,
	arg db.UpsertRateLimitCounterParams,
) (int32, error) {
	f.upsertCalls = append(f.upsertCalls, arg)

	var err error
	if len(f.upsertErrs) > 0 {
		err = f.upsertErrs[0]
		f.upsertErrs = f.upsertErrs[1:]
	}
	if err != nil {
		return 0, err
	}

	if len(f.upsertCounts) == 0 {
		return 0, nil
	}

	count := f.upsertCounts[0]
	f.upsertCounts = f.upsertCounts[1:]

	return count, nil
}

func (f *fakeRateLimitStore) GetRateLimitCounters(
	_ context.Context,
	keys []string,
) ([]db.GetRateLimitCountersRow, error) {
	f.usageKeys = append([]string(nil), keys...)

	if f.usageErr != nil {
		return nil, f.usageErr
	}

	return f.usageRows, nil
}

func (f *fakeRateLimitStore) DeleteExpiredRateLimitCounters(
	_ context.Context,
) error {
	f.cleanupCall = true
	return f.cleanupErr
}

func newTestService(store *fakeRateLimitStore) *Service {
	return &Service{queries: store}
}

func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{
		RetryAfter: 30 * time.Second,
		Message:    "rate limit exceeded",
	}

	if err.Error() != "rate limit exceeded" {
		t.Fatalf("expected error message %q, got %q", "rate limit exceeded", err.Error())
	}

	if err.RetryAfter != 30*time.Second {
		t.Fatalf("expected RetryAfter %v, got %v", 30*time.Second, err.RetryAfter)
	}
}

func TestCheckGlobalLimit(t *testing.T) {
	t.Run("allows request at limit", func(t *testing.T) {
		store := &fakeRateLimitStore{
			upsertCounts: []int32{20, 100},
		}
		service := newTestService(store)

		err := service.CheckGlobalLimit(context.Background(), 20, 100)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if len(store.upsertCalls) != 2 {
			t.Fatalf("expected 2 upsert calls, got %d", len(store.upsertCalls))
		}

		if !strings.HasPrefix(store.upsertCalls[0].Key, "global:rpm:") {
			t.Fatalf("unexpected RPM key: %q", store.upsertCalls[0].Key)
		}

		if !strings.HasPrefix(store.upsertCalls[1].Key, "global:rpd:") {
			t.Fatalf("unexpected RPD key: %q", store.upsertCalls[1].Key)
		}

		if !store.upsertCalls[0].ExpiresAt.Valid {
			t.Fatal("expected RPM expiry to be valid")
		}

		expectedMinuteExpiry := time.Now().Truncate(time.Minute).Add(2 * time.Minute)
		if store.upsertCalls[0].ExpiresAt.Time.Before(expectedMinuteExpiry.Add(-2*time.Second)) ||
			store.upsertCalls[0].ExpiresAt.Time.After(expectedMinuteExpiry.Add(2*time.Second)) {
			t.Fatalf(
				"unexpected RPM expiry: got %v, expected around %v",
				store.upsertCalls[0].ExpiresAt.Time,
				expectedMinuteExpiry,
			)
		}

		if !store.upsertCalls[1].ExpiresAt.Valid {
			t.Fatal("expected RPD expiry to be valid")
		}

		now := time.Now()
		today := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0, 0, 0, 0,
			now.Location(),
		)
		expectedDayExpiry := today.AddDate(0, 0, 1)

		if !store.upsertCalls[1].ExpiresAt.Time.Equal(expectedDayExpiry) {
			t.Fatalf(
				"unexpected RPD expiry: got %v, expected %v",
				store.upsertCalls[1].ExpiresAt.Time,
				expectedDayExpiry,
			)
		}
	})

	t.Run("rejects request above rpm limit", func(t *testing.T) {
		store := &fakeRateLimitStore{
			upsertCounts: []int32{21, 100},
		}
		service := newTestService(store)

		err := service.CheckGlobalLimit(context.Background(), 20, 100)
		if err == nil {
			t.Fatal("expected rate limit error")
		}

		var rateErr *RateLimitError
		if !errors.As(err, &rateErr) {
			t.Fatalf("expected RateLimitError, got %T: %v", err, err)
		}

		if rateErr.Message != "per-minute rate limit exceeded" {
			t.Fatalf("unexpected message: %q", rateErr.Message)
		}

		if rateErr.RetryAfter <= 0 {
			t.Fatalf("expected positive RetryAfter, got %v", rateErr.RetryAfter)
		}
	})

	t.Run("rejects request above rpd limit", func(t *testing.T) {
		store := &fakeRateLimitStore{
			upsertCounts: []int32{20, 101},
		}
		service := newTestService(store)

		err := service.CheckGlobalLimit(context.Background(), 20, 100)
		if err == nil {
			t.Fatal("expected rate limit error")
		}

		var rateErr *RateLimitError
		if !errors.As(err, &rateErr) {
			t.Fatalf("expected RateLimitError, got %T: %v", err, err)
		}

		if rateErr.Message != "daily rate limit exceeded" {
			t.Fatalf("unexpected message: %q", rateErr.Message)
		}

		if rateErr.RetryAfter <= 0 {
			t.Fatalf("expected positive RetryAfter, got %v", rateErr.RetryAfter)
		}
	})

	t.Run("zero limits disable enforcement", func(t *testing.T) {
		store := &fakeRateLimitStore{
			upsertCounts: []int32{1000, 1000},
		}
		service := newTestService(store)

		err := service.CheckGlobalLimit(context.Background(), 0, 0)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("rpm database error is wrapped", func(t *testing.T) {
		expectedErr := errors.New("database unavailable")
		store := &fakeRateLimitStore{
			upsertErrs: []error{expectedErr},
		}
		service := newTestService(store)

		err := service.CheckGlobalLimit(context.Background(), 20, 100)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected wrapped database error, got %v", err)
		}

		if !strings.Contains(err.Error(), "failed to upsert rpm counter") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("rpd database error is wrapped", func(t *testing.T) {
		expectedErr := errors.New("database unavailable")
		store := &fakeRateLimitStore{
			upsertCounts: []int32{1},
			upsertErrs:   []error{nil, expectedErr},
		}
		service := newTestService(store)

		err := service.CheckGlobalLimit(context.Background(), 20, 100)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected wrapped database error, got %v", err)
		}

		if !strings.Contains(err.Error(), "failed to upsert rpd counter") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestCheckKeyLimit(t *testing.T) {
	store := &fakeRateLimitStore{
		upsertCounts: []int32{1, 1},
	}
	service := newTestService(store)

	err := service.CheckKeyLimit(context.Background(), 42, 20, 100)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(store.upsertCalls) != 2 {
		t.Fatalf("expected 2 upsert calls, got %d", len(store.upsertCalls))
	}

	if !strings.HasPrefix(store.upsertCalls[0].Key, "key:42:rpm:") {
		t.Fatalf("unexpected RPM key: %q", store.upsertCalls[0].Key)
	}

	if !strings.HasPrefix(store.upsertCalls[1].Key, "key:42:rpd:") {
		t.Fatalf("unexpected RPD key: %q", store.upsertCalls[1].Key)
	}
}

func TestGetKeyUsage(t *testing.T) {
	t.Run("returns usage counters", func(t *testing.T) {
		store := &fakeRateLimitStore{}
		now := time.Now()

		minuteKey := "key:42:rpm:" + formatMinuteKey(now)
		dayKey := "key:42:rpd:" + now.Format("2006-01-02")

		store.usageRows = []db.GetRateLimitCountersRow{
			{
				Key:   minuteKey,
				Count: 7,
			},
			{
				Key:   dayKey,
				Count: 35,
			},
		}

		service := newTestService(store)

		rpmUsed, rpdUsed, err := service.GetKeyUsage(context.Background(), 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rpmUsed != 7 {
			t.Fatalf("expected rpm usage 7, got %d", rpmUsed)
		}

		if rpdUsed != 35 {
			t.Fatalf("expected rpd usage 35, got %d", rpdUsed)
		}

		if len(store.usageKeys) != 2 {
			t.Fatalf("expected 2 usage keys, got %d", len(store.usageKeys))
		}
	})

	t.Run("returns zero when counters are absent", func(t *testing.T) {
		store := &fakeRateLimitStore{}
		service := newTestService(store)

		rpmUsed, rpdUsed, err := service.GetKeyUsage(context.Background(), 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rpmUsed != 0 || rpdUsed != 0 {
			t.Fatalf("expected 0,0, got %d,%d", rpmUsed, rpdUsed)
		}
	})

	t.Run("wraps database error", func(t *testing.T) {
		expectedErr := errors.New("database unavailable")
		store := &fakeRateLimitStore{
			usageErr: expectedErr,
		}
		service := newTestService(store)

		_, _, err := service.GetKeyUsage(context.Background(), 42)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected wrapped database error, got %v", err)
		}

		if !strings.Contains(err.Error(), "failed to get key usage") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestCleanupExpired(t *testing.T) {
	expectedErr := errors.New("cleanup failed")
	store := &fakeRateLimitStore{
		cleanupErr: expectedErr,
	}
	service := newTestService(store)

	err := service.CleanupExpired(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected cleanup error, got %v", err)
	}

	if !store.cleanupCall {
		t.Fatal("expected cleanup query to be called")
	}
}

func formatMinuteKey(t time.Time) string {
	return fmt.Sprintf("%d", t.Unix()/60)
}
