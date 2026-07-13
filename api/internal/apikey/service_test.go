package apikey_test

import (
	"context"
	"strings"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/password"
)

func TestAPIKeyService_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo := &MockRepository{
			ListByUserIDFunc: func(ctx context.Context, userID uint64) ([]*apikey.APIKey, error) {
				return nil, nil
			},
			CreateFunc: func(ctx context.Context, key *apikey.APIKey) (*apikey.APIKey, error) {
				key.ID = 1
				return key, nil
			},
		}
		auditSvc := &MockAuditService{}
		emailSvc := &MockEmailService{}
		svc := apikey.NewAPIKeyService(repo, auditSvc, emailSvc, "", "")

		plainKey, created, err := svc.GenerateKey(context.Background(), 1, "test-key", 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if plainKey == "" {
			t.Error("expected plaintext key to be returned")
		}
		if created == nil || created.ID == 0 {
			t.Error("expected valid created key")
		}
		if !strings.HasPrefix(plainKey, "rr_") {
			t.Error("expected key to have rr_ prefix")
		}
	})

	t.Run("user already has key", func(t *testing.T) {
		repo := &MockRepository{
			ListByUserIDFunc: func(ctx context.Context, userID uint64) ([]*apikey.APIKey, error) {
				return []*apikey.APIKey{{ID: 1}}, nil
			},
		}
		auditSvc := &MockAuditService{}
		emailSvc := &MockEmailService{}
		svc := apikey.NewAPIKeyService(repo, auditSvc, emailSvc, "", "")

		_, _, err := svc.GenerateKey(context.Background(), 1, "test-key", 100)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestAPIKeyService_ValidateKey(t *testing.T) {
	t.Parallel()

	selector := "AAAAAAAAAAAAAAAAAAAAAAAA"
	verifier := "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"
	plainTextKey := "rr_" + selector + "_" + verifier

	validHash := password.HashSHA256(verifier)

	t.Run("success", func(t *testing.T) {
		repo := &MockRepository{
			GetBySelectorFunc: func(ctx context.Context, sel string) (*apikey.APIKey, error) {
				if sel != selector {
					t.Errorf("expected selector %s, got %s", selector, sel)
				}
				return &apikey.APIKey{
					ID:          1,
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatusActive,
				}, nil
			},
		}
		svc := apikey.NewAPIKeyService(repo, &MockAuditService{}, &MockEmailService{}, "", "")

		key, err := svc.ValidateKey(context.Background(), plainTextKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key == nil {
			t.Error("expected key to be returned")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		svc := apikey.NewAPIKeyService(&MockRepository{}, &MockAuditService{}, &MockEmailService{}, "", "")

		_, err := svc.ValidateKey(context.Background(), "invalid-format")
		if err != apikey.ErrInvalidAPIKey {
			t.Errorf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("wrong verifier hash", func(t *testing.T) {
		repo := &MockRepository{
			GetBySelectorFunc: func(ctx context.Context, sel string) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     password.HashSHA256("WRONGVERIFIER"),
					Status:      apikey.APIKeyStatusActive,
				}, nil
			},
		}
		svc := apikey.NewAPIKeyService(repo, &MockAuditService{}, &MockEmailService{}, "", "")

		_, err := svc.ValidateKey(context.Background(), plainTextKey)
		if err != apikey.ErrInvalidAPIKey {
			t.Errorf("expected ErrInvalidAPIKey, got %v", err)
		}
	})

	t.Run("suspended key", func(t *testing.T) {
		repo := &MockRepository{
			GetBySelectorFunc: func(ctx context.Context, sel string) (*apikey.APIKey, error) {
				return &apikey.APIKey{
					KeySelector: selector,
					KeyHash:     validHash,
					Status:      apikey.APIKeyStatusSuspended,
				}, nil
			},
		}
		svc := apikey.NewAPIKeyService(repo, &MockAuditService{}, &MockEmailService{}, "", "")

		_, err := svc.ValidateKey(context.Background(), plainTextKey)
		if err != apikey.ErrAPIKeySuspended {
			t.Errorf("expected ErrAPIKeySuspended, got %v", err)
		}
	})
}
