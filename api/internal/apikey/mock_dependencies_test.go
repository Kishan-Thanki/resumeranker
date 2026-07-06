package apikey_test

import (
	"context"

	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

type MockRepository struct {
	CreateFunc        func(ctx context.Context, apiKey *apikey.APIKey) (*apikey.APIKey, error)
	GetByIDFunc       func(ctx context.Context, id uint64) (*apikey.APIKey, error)
	GetBySelectorFunc func(ctx context.Context, selector string) (*apikey.APIKey, error)
	ListByUserIDFunc  func(ctx context.Context, userID uint64) ([]*apikey.APIKey, error)
	UpdateFunc        func(ctx context.Context, apiKey *apikey.APIKey) (*apikey.APIKey, error)
	DeleteFunc        func(ctx context.Context, id uint64) error
	IsUserActiveFunc  func(ctx context.Context, userID uint64) (bool, error)
}

func (m *MockRepository) Create(ctx context.Context, key *apikey.APIKey) (*apikey.APIKey, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, key)
	}
	return key, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uint64) (*apikey.APIKey, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) GetBySelector(ctx context.Context, selector string) (*apikey.APIKey, error) {
	if m.GetBySelectorFunc != nil {
		return m.GetBySelectorFunc(ctx, selector)
	}
	return nil, nil
}

func (m *MockRepository) ListByUserID(ctx context.Context, userID uint64) ([]*apikey.APIKey, error) {
	if m.ListByUserIDFunc != nil {
		return m.ListByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRepository) Update(ctx context.Context, key *apikey.APIKey) (*apikey.APIKey, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, key)
	}
	return key, nil
}

func (m *MockRepository) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRepository) IsUserActive(ctx context.Context, userID uint64) (bool, error) {
	if m.IsUserActiveFunc != nil {
		return m.IsUserActiveFunc(ctx, userID)
	}
	return true, nil
}

type MockAuditService struct {
	LogEventFunc func(ctx context.Context, event *audit.AuditEvent) error
}

func (m *MockAuditService) LogEvent(ctx context.Context, event *audit.AuditEvent) error {
	if m.LogEventFunc != nil {
		return m.LogEventFunc(ctx, event)
	}
	return nil
}

type MockAPIKeyService struct {
	GenerateKeyFunc  func(ctx context.Context, userID uint64, name string, quota uint64) (string, *apikey.APIKey, error)
	ValidateKeyFunc  func(ctx context.Context, plainTextKey string) (*apikey.APIKey, error)
	DeductTokensFunc func(ctx context.Context, key *apikey.APIKey, tokensUsed uint64) error
	ListKeysFunc     func(ctx context.Context, userID uint64) ([]*apikey.APIKey, error)
	ToggleStatusFunc func(ctx context.Context, userID, keyID uint64, status apikey.APIKeyStatus) error
	RevokeKeyFunc    func(ctx context.Context, userID, keyID uint64) error
}

func (m *MockAPIKeyService) GenerateKey(ctx context.Context, userID uint64, name string, quota uint64) (string, *apikey.APIKey, error) {
	if m.GenerateKeyFunc != nil {
		return m.GenerateKeyFunc(ctx, userID, name, quota)
	}
	return "", nil, nil
}

func (m *MockAPIKeyService) ValidateKey(ctx context.Context, plainTextKey string) (*apikey.APIKey, error) {
	if m.ValidateKeyFunc != nil {
		return m.ValidateKeyFunc(ctx, plainTextKey)
	}
	return nil, nil
}

func (m *MockAPIKeyService) DeductTokens(ctx context.Context, key *apikey.APIKey, tokensUsed uint64) error {
	if m.DeductTokensFunc != nil {
		return m.DeductTokensFunc(ctx, key, tokensUsed)
	}
	return nil
}

func (m *MockAPIKeyService) ListKeys(ctx context.Context, userID uint64) ([]*apikey.APIKey, error) {
	if m.ListKeysFunc != nil {
		return m.ListKeysFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockAPIKeyService) ToggleStatus(ctx context.Context, userID, keyID uint64, status apikey.APIKeyStatus) error {
	if m.ToggleStatusFunc != nil {
		return m.ToggleStatusFunc(ctx, userID, keyID, status)
	}
	return nil
}

func (m *MockAPIKeyService) RevokeKey(ctx context.Context, userID, keyID uint64) error {
	if m.RevokeKeyFunc != nil {
		return m.RevokeKeyFunc(ctx, userID, keyID)
	}
	return nil
}
