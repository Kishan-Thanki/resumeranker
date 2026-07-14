package analysis

import (
	"context"

	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

type MockAPIKeyValidator struct {
	ValidateKeyFunc  func(ctx context.Context, plainTextKey string) (*apikey.APIKey, error)
	DeductTokensFunc func(ctx context.Context, key *apikey.APIKey, tokensUsed uint64) error
}

func (m *MockAPIKeyValidator) ValidateKey(ctx context.Context, plainTextKey string) (*apikey.APIKey, error) {
	if m.ValidateKeyFunc != nil {
		return m.ValidateKeyFunc(ctx, plainTextKey)
	}
	return &apikey.APIKey{ID: 1, UserID: 1, TokenQuota: 1000, TokensUsed: 0}, nil
}

func (m *MockAPIKeyValidator) DeductTokens(ctx context.Context, key *apikey.APIKey, tokensUsed uint64) error {
	if m.DeductTokensFunc != nil {
		return m.DeductTokensFunc(ctx, key, tokensUsed)
	}
	key.TokensUsed += tokensUsed
	return nil
}

type MockAuditService struct {
	LogEventFunc func(ctx context.Context, event *audit.AuditEvent) error
	Events       []*audit.AuditEvent
}

func (m *MockAuditService) LogEvent(ctx context.Context, event *audit.AuditEvent) error {
	m.Events = append(m.Events, event)
	if m.LogEventFunc != nil {
		return m.LogEventFunc(ctx, event)
	}
	return nil
}

type MockRepository struct {
	CreateRequestFunc        func(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error)
	GetRequestByIDFunc       func(ctx context.Context, id uint64) (*AnalysisRequest, error)
	GetRequestByUUIDFunc     func(ctx context.Context, requestID string) (*AnalysisRequest, error)
	ListRequestsByUserIDFunc func(ctx context.Context, userID uint64, limit, offset int) ([]*AnalysisRequest, error)
	UpdateRequestFunc        func(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error)
	DeleteRequestFunc        func(ctx context.Context, id uint64) error

	CreateResultFunc         func(ctx context.Context, result *AnalysisResult) (*AnalysisResult, error)
	GetResultByRequestIDFunc func(ctx context.Context, requestID uint64) (*AnalysisResult, error)
	DeleteResultFunc         func(ctx context.Context, id uint64) error
}

func (m *MockRepository) CreateRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error) {
	if m.CreateRequestFunc != nil {
		return m.CreateRequestFunc(ctx, req)
	}
	req.ID = 1
	return req, nil
}
func (m *MockRepository) GetRequestByID(ctx context.Context, id uint64) (*AnalysisRequest, error) {
	if m.GetRequestByIDFunc != nil {
		return m.GetRequestByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *MockRepository) GetRequestByUUID(ctx context.Context, requestID string) (*AnalysisRequest, error) {
	if m.GetRequestByUUIDFunc != nil {
		return m.GetRequestByUUIDFunc(ctx, requestID)
	}
	return &AnalysisRequest{ID: 1, UserID: 1}, nil
}
func (m *MockRepository) ListRequestsByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*AnalysisRequest, error) {
	if m.ListRequestsByUserIDFunc != nil {
		return m.ListRequestsByUserIDFunc(ctx, userID, limit, offset)
	}
	return nil, nil
}
func (m *MockRepository) UpdateRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error) {
	if m.UpdateRequestFunc != nil {
		return m.UpdateRequestFunc(ctx, req)
	}
	return req, nil
}
func (m *MockRepository) DeleteRequest(ctx context.Context, id uint64) error {
	if m.DeleteRequestFunc != nil {
		return m.DeleteRequestFunc(ctx, id)
	}
	return nil
}
func (m *MockRepository) CreateResult(ctx context.Context, result *AnalysisResult) (*AnalysisResult, error) {
	if m.CreateResultFunc != nil {
		return m.CreateResultFunc(ctx, result)
	}
	result.ID = 1
	return result, nil
}
func (m *MockRepository) GetResultByRequestID(ctx context.Context, requestID uint64) (*AnalysisResult, error) {
	if m.GetResultByRequestIDFunc != nil {
		return m.GetResultByRequestIDFunc(ctx, requestID)
	}
	return nil, nil
}
func (m *MockRepository) DeleteResult(ctx context.Context, id uint64) error {
	if m.DeleteResultFunc != nil {
		return m.DeleteResultFunc(ctx, id)
	}
	return nil
}

type MockRateLimiter struct {
	CheckGlobalLimitFunc func(ctx context.Context, rpmLimit, rpdLimit int) error
	CheckKeyLimitFunc    func(ctx context.Context, apiKeyID uint64, rpmLimit, rpdLimit int) error
}

func (m *MockRateLimiter) CheckGlobalLimit(ctx context.Context, rpmLimit, rpdLimit int) error {
	if m.CheckGlobalLimitFunc != nil {
		return m.CheckGlobalLimitFunc(ctx, rpmLimit, rpdLimit)
	}
	return nil
}

func (m *MockRateLimiter) CheckKeyLimit(ctx context.Context, apiKeyID uint64, rpmLimit, rpdLimit int) error {
	if m.CheckKeyLimitFunc != nil {
		return m.CheckKeyLimitFunc(ctx, apiKeyID, rpmLimit, rpdLimit)
	}
	return nil
}
