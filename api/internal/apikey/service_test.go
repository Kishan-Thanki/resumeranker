package apikey_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/password"
)

func newAPIKeyService(
	repo *MockRepository,
	auditSvc *MockAuditService,
	emailSvc *MockEmailService,
	rateLimiter *MockRateLimiter,
	wg *sync.WaitGroup,
) *apikey.APIKeyService {
	return apikey.NewAPIKeyService(
		repo,
		auditSvc,
		emailSvc,
		rateLimiter,
		"https://example.com",
		"support@example.com",
		wg,
	)
}

func TestNewAPIKeyServiceNilWaitGroup(t *testing.T) {
	svc := newAPIKeyService(
		&MockRepository{},
		&MockAuditService{},
		&MockEmailService{},
		&MockRateLimiter{},
		nil,
	)

	if svc == nil {
		t.Fatal("expected service")
	}
}

func TestAPIKeyService_GetByID(t *testing.T) {
	t.Parallel()

	expected := &apikey.APIKey{ID: 42}

	repo := &MockRepository{
		GetByIDFunc: func(
			_ context.Context,
			id uint64,
		) (*apikey.APIKey, error) {
			if id != 42 {
				t.Fatalf("expected id 42, got %d", id)
			}
			return expected, nil
		},
	}

	svc := newAPIKeyService(
		repo,
		&MockAuditService{},
		&MockEmailService{},
		&MockRateLimiter{},
		new(sync.WaitGroup),
	)

	got, err := svc.GetByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatal("expected repository result")
	}
}

func TestAPIKeyService_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			ListByUserIDFunc: func(
				_ context.Context,
				_ uint64,
			) ([]*apikey.APIKey, error) {
				return nil, nil
			},
			CreateFunc: func(
				_ context.Context,
				key *apikey.APIKey,
			) (*apikey.APIKey, error) {
				key.ID = 1
				return key, nil
			},
		}

		auditSvc := &MockAuditService{}
		emailSvc := &MockEmailService{}
		wg := &sync.WaitGroup{}

		svc := newAPIKeyService(
			repo,
			auditSvc,
			emailSvc,
			&MockRateLimiter{},
			wg,
		)

		plainKey, created, err := svc.GenerateKey(
			context.Background(),
			1,
			"test-key",
			100,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if created == nil || created.ID != 1 {
			t.Fatalf("expected created key with ID 1, got %#v", created)
		}

		parts := strings.Split(plainKey, "_")
		if len(parts) != 3 {
			t.Fatalf("expected three key parts, got %d", len(parts))
		}

		if parts[0] != "rr" {
			t.Fatalf("expected rr prefix, got %q", parts[0])
		}

		if len(parts[1]) != 24 {
			t.Fatalf("expected 24-character selector, got %d", len(parts[1]))
		}

		if len(parts[2]) != 48 {
			t.Fatalf("expected 48-character verifier, got %d", len(parts[2]))
		}

		if created.KeyPrefix == "" {
			t.Fatal("expected key prefix")
		}

		wg.Wait()
	})

	t.Run("repository list error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("list failed")

		repo := &MockRepository{
			ListByUserIDFunc: func(
				_ context.Context,
				_ uint64,
			) ([]*apikey.APIKey, error) {
				return nil, expectedErr
			},
		}

		svc := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		)

		_, _, err := svc.GenerateKey(context.Background(), 1, "key", 100)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})

	t.Run("user already has key", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			ListByUserIDFunc: func(
				_ context.Context,
				_ uint64,
			) ([]*apikey.APIKey, error) {
				return []*apikey.APIKey{{ID: 1}}, nil
			},
		}

		svc := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		)

		_, _, err := svc.GenerateKey(
			context.Background(),
			1,
			"test-key",
			100,
		)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("repository create error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("create failed")

		repo := &MockRepository{
			CreateFunc: func(
				_ context.Context,
				_ *apikey.APIKey,
			) (*apikey.APIKey, error) {
				return nil, expectedErr
			},
		}

		svc := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		)

		_, _, err := svc.GenerateKey(
			context.Background(),
			1,
			"test-key",
			100,
		)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func TestAPIKeyService_ValidateKey(t *testing.T) {
	t.Parallel()

	selector := "AAAAAAAAAAAAAAAAAAAAAAAA"
	verifier := "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"
	plainTextKey := "rr_" + selector + "_" + verifier
	validHash := password.HashSHA256(verifier)

	newService := func(repo *MockRepository) *apikey.APIKeyService {
		return newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		)
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				sel string,
			) (*apikey.APIKey, error) {
				if sel != selector {
					t.Errorf("expected selector %s, got %s", selector, sel)
				}
				return &apikey.APIKey{
					ID:          1,
					UserID:      10,
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatusActive,
					ExpiresAt:   func() *time.Time { t := now.Add(time.Hour); return &t }(),
				}, nil
			},
			IsUserActiveFunc: func(
				_ context.Context,
				userID uint64,
			) (bool, error) {
				if userID != 10 {
					t.Errorf("expected user ID 10, got %d", userID)
				}
				return true, nil
			},
			UpdateFunc: func(
				_ context.Context,
				key *apikey.APIKey,
			) (*apikey.APIKey, error) {
				if key.LastUsedAt == nil {
					t.Fatal("expected LastUsedAt")
				}
				return key, nil
			},
		}

		key, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key == nil {
			t.Fatal("expected key")
		}
		if key.LastUsedAt == nil {
			t.Fatal("expected LastUsedAt")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()

		svc := newService(&MockRepository{})

		for _, value := range []string{
			"invalid-format",
			"rr_only_two",
			"rr_ABC_",
			"rr_ABC_123",
			"xx_" + selector + "_" + verifier,
			"rr_" + selector + "_" + verifier + "_extra",
		} {
			_, err := svc.ValidateKey(context.Background(), value)
			if !errors.Is(err, apikey.ErrInvalidAPIKey) {
				t.Errorf("expected ErrInvalidAPIKey for %q, got %v", value, err)
			}
		}
	})

	t.Run("repository lookup error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("lookup failed")
		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return nil, expectedErr
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, apikey.ErrInvalidAPIKey) {
			t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("nil repository result", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return nil, nil
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, apikey.ErrInvalidAPIKey) {
			t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("wrong verifier hash", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     password.HashSHA256("WRONGVERIFIER"),
					Status:      apikey.APIKeyStatusActive,
				}, nil
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, apikey.ErrInvalidAPIKey) {
			t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("inactive key", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatusInactive,
				}, nil
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, apikey.ErrAPIKeyInactive) {
			t.Fatalf("expected ErrAPIKeyInactive, got %v", err)
		}
	})

	t.Run("suspended key", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatusSuspended,
				}, nil
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, apikey.ErrAPIKeySuspended) {
			t.Fatalf("expected ErrAPIKeySuspended, got %v", err)
		}
	})

	t.Run("unknown status", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatus("unknown"),
				}, nil
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, apikey.ErrInvalidAPIKey) {
			t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("expired key", func(t *testing.T) {
		t.Parallel()

		expired := time.Now().Add(-time.Minute)
		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatusActive,
					ExpiresAt:   &expired,
				}, nil
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, apikey.ErrInvalidAPIKey) {
			t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("inactive user", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatusActive,
				}, nil
			},
			IsUserActiveFunc: func(
				_ context.Context,
				_ uint64,
			) (bool, error) {
				return false, nil
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, apikey.ErrInvalidAPIKey) {
			t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("update failed")
		repo := &MockRepository{
			GetBySelectorFunc: func(
				_ context.Context,
				_ string,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatusActive,
				}, nil
			},
			IsUserActiveFunc: func(
				_ context.Context,
				_ uint64,
			) (bool, error) {
				return true, nil
			},
			UpdateFunc: func(
				_ context.Context,
				_ *apikey.APIKey,
			) (*apikey.APIKey, error) {
				return nil, expectedErr
			},
		}

		_, err := newService(repo).ValidateKey(
			context.Background(),
			plainTextKey,
		)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func TestAPIKeyService_DeductTokens(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		key := &apikey.APIKey{
			TokensUsed: 10,
			TokenQuota: 100,
		}

		repo := &MockRepository{
			UpdateFunc: func(
				_ context.Context,
				got *apikey.APIKey,
			) (*apikey.APIKey, error) {
				if got.TokensUsed != 30 {
					t.Fatalf("expected TokensUsed 30, got %d", got.TokensUsed)
				}
				return got, nil
			},
		}

		err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).DeductTokens(context.Background(), key, 20)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if key.TokensUsed != 30 {
			t.Fatalf("expected key usage 30, got %d", key.TokensUsed)
		}
	})

	t.Run("nil key", func(t *testing.T) {
		t.Parallel()

		err := newAPIKeyService(
			&MockRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).DeductTokens(context.Background(), nil, 1)

		if !errors.Is(err, apikey.ErrInvalidAPIKey) {
			t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("quota exceeded", func(t *testing.T) {
		t.Parallel()

		key := &apikey.APIKey{
			TokensUsed: 90,
			TokenQuota: 100,
		}

		err := newAPIKeyService(
			&MockRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).DeductTokens(context.Background(), key, 11)

		if !errors.Is(err, apikey.ErrTokenQuotaExceeded) {
			t.Fatalf("expected ErrTokenQuotaExceeded, got %v", err)
		}
	})

	t.Run("usage already exceeds quota", func(t *testing.T) {
		t.Parallel()

		key := &apikey.APIKey{
			TokensUsed: 101,
			TokenQuota: 100,
		}

		err := newAPIKeyService(
			&MockRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).DeductTokens(context.Background(), key, 1)

		if !errors.Is(err, apikey.ErrTokenQuotaExceeded) {
			t.Fatalf("expected ErrTokenQuotaExceeded, got %v", err)
		}
	})

	t.Run("repository update error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("update failed")
		repo := &MockRepository{
			UpdateFunc: func(
				_ context.Context,
				_ *apikey.APIKey,
			) (*apikey.APIKey, error) {
				return nil, expectedErr
			},
		}

		key := &apikey.APIKey{
			TokensUsed: 10,
			TokenQuota: 100,
		}

		err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).DeductTokens(context.Background(), key, 20)

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func TestAPIKeyService_ListKeys(t *testing.T) {
	t.Parallel()

	expected := []*apikey.APIKey{{ID: 1}, {ID: 2}}

	repo := &MockRepository{
		ListByUserIDFunc: func(
			_ context.Context,
			userID uint64,
		) ([]*apikey.APIKey, error) {
			if userID != 42 {
				t.Fatalf("expected user ID 42, got %d", userID)
			}
			return expected, nil
		},
	}

	got, err := newAPIKeyService(
		repo,
		&MockAuditService{},
		&MockEmailService{},
		&MockRateLimiter{},
		new(sync.WaitGroup),
	).ListKeys(context.Background(), 42)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(got))
	}
}

func TestAPIKeyService_ToggleStatus(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		key := &apikey.APIKey{
			ID:     10,
			UserID: 20,
			Status: apikey.APIKeyStatusActive,
		}

		wg := &sync.WaitGroup{}
		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return key, nil
			},
			UpdateFunc: func(
				_ context.Context,
				got *apikey.APIKey,
			) (*apikey.APIKey, error) {
				if got.Status != apikey.APIKeyStatusSuspended {
					t.Fatalf("expected suspended status, got %q", got.Status)
				}
				return got, nil
			},
			GetUserEmailByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (string, error) {
				return "", nil
			},
		}

		err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			wg,
		).ToggleStatus(
			context.Background(),
			20,
			10,
			apikey.APIKeyStatusSuspended,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{}
		err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).ToggleStatus(
			context.Background(),
			20,
			10,
			apikey.APIKeyStatus("invalid"),
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{ID: 10, UserID: 99}, nil
			},
		}

		err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).ToggleStatus(context.Background(), 20, 10, apikey.APIKeyStatusActive)

		if err == nil || !strings.Contains(err.Error(), "unauthorized") {
			t.Fatalf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("repository update error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("update failed")
		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{ID: 10, UserID: 20}, nil
			},
			UpdateFunc: func(
				_ context.Context,
				_ *apikey.APIKey,
			) (*apikey.APIKey, error) {
				return nil, expectedErr
			},
		}

		err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).ToggleStatus(context.Background(), 20, 10, apikey.APIKeyStatusActive)

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func TestAPIKeyService_RevokeKey(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var auditCalled bool

		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{ID: 10, UserID: 20}, nil
			},
			DeleteFunc: func(
				_ context.Context,
				id uint64,
			) error {
				if id != 10 {
					t.Fatalf("expected key ID 10, got %d", id)
				}
				return nil
			},
			GetUserEmailByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (string, error) {
				return "", nil
			},
		}

		auditSvc := &MockAuditService{
			LogEventFunc: func(
				_ context.Context,
				event *audit.AuditEvent,
			) error {
				auditCalled = true
				if event.Type != audit.AuditEventAPIKeyRevoked {
					t.Fatalf("unexpected audit event type %q", event.Type)
				}
				return nil
			},
		}

		err := newAPIKeyService(
			repo,
			auditSvc,
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).RevokeKey(context.Background(), 20, 10)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !auditCalled {
			t.Fatal("expected audit event")
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{ID: 10, UserID: 99}, nil
			},
		}

		err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).RevokeKey(context.Background(), 20, 10)

		if err == nil || !strings.Contains(err.Error(), "unauthorized") {
			t.Fatalf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("delete failed")
		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{ID: 10, UserID: 20}, nil
			},
			DeleteFunc: func(
				_ context.Context,
				_ uint64,
			) error {
				return expectedErr
			},
		}

		err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).RevokeKey(context.Background(), 20, 10)

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func TestAPIKeyService_GetAPIKeyStats(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					ID:                10,
					UserID:            20,
					RequestsPerMinute: 20,
					RequestsPerDay:    500,
					TokensUsed:        75,
					TokenQuota:        1000,
				}, nil
			},
		}

		rateLimiter := &MockRateLimiter{
			GetKeyUsageFunc: func(
				_ context.Context,
				keyID uint64,
			) (int, int, error) {
				if keyID != 10 {
					t.Fatalf("expected key ID 10, got %d", keyID)
				}
				return 5, 40, nil
			},
		}

		stats, err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			rateLimiter,
			new(sync.WaitGroup),
		).GetAPIKeyStats(context.Background(), 20, 10)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stats.RPMUsed != 5 || stats.RPMLimit != 20 {
			t.Fatalf("unexpected RPM values: %+v", stats)
		}

		if stats.RPDUsed != 40 || stats.RPDLimit != 500 {
			t.Fatalf("unexpected RPD values: %+v", stats)
		}

		if stats.TokensUsed != 75 || stats.TokenQuota != 1000 {
			t.Fatalf("unexpected token values: %+v", stats)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()

		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{ID: 10, UserID: 99}, nil
			},
		}

		_, err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			&MockRateLimiter{},
			new(sync.WaitGroup),
		).GetAPIKeyStats(context.Background(), 20, 10)

		if err == nil || !strings.Contains(err.Error(), "unauthorized") {
			t.Fatalf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("rate limiter error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("rate limiter failed")
		repo := &MockRepository{
			GetByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*apikey.APIKey, error) {
				return &apikey.APIKey{ID: 10, UserID: 20}, nil
			},
		}

		rateLimiter := &MockRateLimiter{
			GetKeyUsageFunc: func(
				_ context.Context,
				_ uint64,
			) (int, int, error) {
				return 0, 0, expectedErr
			},
		}

		_, err := newAPIKeyService(
			repo,
			&MockAuditService{},
			&MockEmailService{},
			rateLimiter,
			new(sync.WaitGroup),
		).GetAPIKeyStats(context.Background(), 20, 10)

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}
