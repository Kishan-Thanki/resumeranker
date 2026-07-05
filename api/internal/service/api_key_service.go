package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"strings"

	"github.com/kishan-thanki/resumeranker/api/internal/domain"
	"github.com/kishan-thanki/resumeranker/api/pkg/hashutil"
)

var (
	ErrInvalidAPIKey = errors.New("invalid or expired api key")
)

type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *domain.APIKey) (*domain.APIKey, error)
	GetByID(ctx context.Context, id uint64) (*domain.APIKey, error)
	GetBySelector(ctx context.Context, selector string) (*domain.APIKey, error)
	Update(ctx context.Context, apiKey *domain.APIKey) (*domain.APIKey, error)
}

type APIKeyService struct {
	repo APIKeyRepository
}

func NewAPIKeyService(repo APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		repo: repo,
	}
}

func (s *APIKeyService) GenerateKey(ctx context.Context, userID uint64, name string, quota uint64) (string, *domain.APIKey, error) {
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

	keyHash := hashutil.HashSHA256(verifier)

	apiKey := &domain.APIKey{
		UserID:      userID,
		Name:        name,
		KeySelector: selector,
		KeyHash:     keyHash,
		Status:      domain.APIKeyStatusActive,
		TokenQuota:  quota,
	}

	createdKey, err := s.repo.Create(ctx, apiKey)
	if err != nil {
		return "", nil, err
	}

	plainTextKey := "rr_" + strings.ToLower(selector) + "_" + strings.ToLower(verifier)

	return plainTextKey, createdKey, nil
}

func (s *APIKeyService) ValidateKey(ctx context.Context, plainTextKey string) (*domain.APIKey, error) {
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

	incomingHash := hashutil.HashSHA256(verifier)
	if subtle.ConstantTimeCompare([]byte(apiKey.KeyHash), []byte(incomingHash)) != 1 {
		return nil, ErrInvalidAPIKey
	}

	if apiKey.Status != domain.APIKeyStatusActive {
		return nil, ErrInvalidAPIKey
	}

	return apiKey, nil
}

func (s *APIKeyService) DeductTokens(ctx context.Context, apiKey *domain.APIKey, tokensUsed uint64) error {
	apiKey.TokensUsed += tokensUsed
	_, err := s.repo.Update(ctx, apiKey)
	return err
}
