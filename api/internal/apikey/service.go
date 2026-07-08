package apikey

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"strings"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/email"
	"github.com/kishan-thanki/resumeranker/api/internal/password"
)

var (
	ErrInvalidAPIKey = errors.New("invalid or expired api key")
)

type auditService interface {
	LogEvent(ctx context.Context, event *audit.AuditEvent) error
}

type APIKeyService struct {
	repo         Repository
	auditService auditService
	emailService email.Service
}

func NewAPIKeyService(repo Repository, auditService auditService, emailService email.Service) *APIKeyService {
	return &APIKeyService{
		repo:         repo,
		auditService: auditService,
		emailService: emailService,
	}
}

func (s *APIKeyService) GenerateKey(ctx context.Context, userID uint64, name string, quota uint64) (string, *APIKey, error) {
	existingKeys, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	if len(existingKeys) > 0 {
		return "", nil, errors.New("user already has an API key; please revoke the existing key before generating a new one")
	}

	selectorBytes := make([]byte, 15)
	if _, err := rand.Read(selectorBytes); err != nil {
		return "", nil, err
	}
	selector := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(selectorBytes)

	verifierBytes := make([]byte, 30)
	if _, err := rand.Read(verifierBytes); err != nil {
		return "", nil, err
	}
	verifier := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(verifierBytes)

	keyHash := password.HashSHA256(verifier)

	apiKey := &APIKey{
		UserID:      userID,
		Name:        name,
		KeyPrefix:   "rr_" + strings.ToLower(selector) + "_",
		KeySelector: selector,
		KeyHash:     keyHash,
		Status:      APIKeyStatusActive,
		TokenQuota:  quota,
	}

	createdKey, err := s.repo.Create(ctx, apiKey)
	if err != nil {
		return "", nil, err
	}

	plainTextKey := "rr_" + strings.ToLower(selector) + "_" + strings.ToLower(verifier)

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventAPIKeyCreated,
		Description: "api key generated successfully",
		UserID:      &userID,
		APIKeyID:    &createdKey.ID,
	})

	userEmail, err := s.repo.GetUserEmailByID(ctx, userID)
	if err == nil && userEmail != "" {
		go func() {
			_ = s.emailService.SendEmail(context.Background(), &email.SendEmailRequest{
				To:      []string{userEmail},
				Subject: "New API Key Generated",
				Text:    "A new API key was just generated for your account. If you did not authorize this, please log in and revoke it immediately.",
			})
		}()
	}

	return plainTextKey, createdKey, nil
}

func (s *APIKeyService) ValidateKey(ctx context.Context, plainTextKey string) (*APIKey, error) {
	parts := strings.Split(plainTextKey, "_")
	if len(parts) != 3 || parts[0] != "rr" {
		return nil, ErrInvalidAPIKey
	}

	selector := strings.ToUpper(parts[1])
	verifier := strings.ToUpper(parts[2])

	apiKey, err := s.repo.GetBySelector(ctx, selector)
	if err != nil {
		return nil, ErrInvalidAPIKey
	}

	incomingHash := password.HashSHA256(verifier)
	if subtle.ConstantTimeCompare([]byte(apiKey.KeyHash), []byte(incomingHash)) != 1 {
		return nil, ErrInvalidAPIKey
	}

	if apiKey.Status != APIKeyStatusActive {
		return nil, ErrInvalidAPIKey
	}

	isActive, err := s.repo.IsUserActive(ctx, apiKey.UserID)
	if err != nil || !isActive {
		return nil, ErrInvalidAPIKey
	}

	apiKey.LastUsedAt = func(t time.Time) *time.Time { return &t }(time.Now())
	_, _ = s.repo.Update(ctx, apiKey)

	return apiKey, nil
}

func (s *APIKeyService) DeductTokens(ctx context.Context, apiKey *APIKey, tokensUsed uint64) error {
	apiKey.TokensUsed += tokensUsed
	_, err := s.repo.Update(ctx, apiKey)
	return err
}

func (s *APIKeyService) ListKeys(ctx context.Context, userID uint64) ([]*APIKey, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *APIKeyService) ToggleStatus(ctx context.Context, userID, keyID uint64, status APIKeyStatus) error {
	apiKey, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return err
	}

	if apiKey.UserID != userID {
		return errors.New("unauthorized: key does not belong to user")
	}

	apiKey.Status = status
	_, err = s.repo.Update(ctx, apiKey)
	if err != nil {
		return err
	}

	userEmail, err := s.repo.GetUserEmailByID(ctx, userID)
	if err == nil && userEmail != "" {
		go func() {
			_ = s.emailService.SendEmail(context.Background(), &email.SendEmailRequest{
				To:      []string{userEmail},
				Subject: "API Key Status Updated",
				Text:    "The status of your API key has been updated to: " + string(status) + ".",
			})
		}()
	}

	return nil
}

func (s *APIKeyService) RevokeKey(ctx context.Context, userID, keyID uint64) error {
	apiKey, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return err
	}

	if apiKey.UserID != userID {
		return errors.New("unauthorized: key does not belong to user")
	}

	err = s.repo.Delete(ctx, keyID)
	if err != nil {
		return err
	}

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventAPIKeyRevoked,
		Description: "api key revoked",
		UserID:      &userID,
		APIKeyID:    &keyID,
	})

	userEmail, err := s.repo.GetUserEmailByID(ctx, userID)
	if err == nil && userEmail != "" {
		go func() {
			_ = s.emailService.SendEmail(context.Background(), &email.SendEmailRequest{
				To:      []string{userEmail},
				Subject: "API Key Revoked",
				Text:    "An API key associated with your account was successfully revoked and deleted.",
			})
		}()
	}

	return nil
}
