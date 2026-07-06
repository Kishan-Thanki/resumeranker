package analysis

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

var (
	ErrInsufficientQuota = errors.New("insufficient token quota")
	ErrAnalysisFailed    = errors.New("analysis failed")
)

type APIKeyValidator interface {
	ValidateKey(ctx context.Context, plainTextKey string) (*apikey.APIKey, error)
	DeductTokens(ctx context.Context, apiKey *apikey.APIKey, tokensUsed uint64) error
}

type auditService interface {
	LogEvent(ctx context.Context, event *audit.AuditEvent) error
}

type EngineRequest struct {
	ResumeText     string `json:"resume_text"`
	JobDescription string `json:"job_description"`
}

type EngineResponse struct {
	ResultJSON       string `json:"result_json"`
	Model            string `json:"model"`
	PromptTokens     uint32 `json:"prompt_tokens"`
	CompletionTokens uint32 `json:"completion_tokens"`
	TotalTokens      uint32 `json:"total_tokens"`
}

type EngineClient interface {
	Analyze(ctx context.Context, req *EngineRequest) (*EngineResponse, error)
}

type AnalysisService struct {
	repo         Repository
	auditService auditService
	keyService   APIKeyValidator
	engineClient EngineClient
}

func NewAnalysisService(
	repo Repository,
	auditService auditService,
	keyService APIKeyValidator,
	engineClient EngineClient,
) *AnalysisService {
	return &AnalysisService{
		repo:         repo,
		auditService: auditService,
		keyService:   keyService,
		engineClient: engineClient,
	}
}

func (s *AnalysisService) ProcessResume(ctx context.Context, plainTextKey, resumeText, jobDescription string) (*AnalysisResult, error) {
	apiKey, err := s.keyService.ValidateKey(ctx, plainTextKey)
	if err != nil {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventAPIKeyUsed,
			Description: "failed to validate api key",
		})
		return nil, err
	}

	if apiKey.TokenQuota <= apiKey.TokensUsed {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventAPIKeyUsed,
			Description: "token quota exceeded",
			UserID:      &apiKey.UserID,
			APIKeyID:    &apiKey.ID,
		})
		return nil, ErrInsufficientQuota
	}

	now := time.Now()
	requestID := uuid.New().String()

	req := &AnalysisRequest{
		RequestID: requestID,
		UserID:    apiKey.UserID,
		APIKeyID:  apiKey.ID,
		Status:    AnalysisRequestStatusProcessing,
		StartedAt: &now,
	}

	req, err = s.repo.CreateRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	engReq := &EngineRequest{
		ResumeText:     resumeText,
		JobDescription: jobDescription,
	}

	engRes, err := s.engineClient.Analyze(ctx, engReq)
	if err != nil {
		errStr := err.Error()
		req.Status = AnalysisRequestStatusFailed
		req.Error = &errStr
		completed := time.Now()
		req.CompletedAt = &completed
		_, _ = s.repo.UpdateRequest(ctx, req)
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventAnalysisFailed,
			Description: "engine analysis failed",
			UserID:      &apiKey.UserID,
			APIKeyID:    &apiKey.ID,
		})
		return nil, ErrAnalysisFailed
	}

	res := &AnalysisResult{
		AnalysisRequestID: req.ID,
		Model:             engRes.Model,
		Result:            engRes.ResultJSON,
		PromptTokens:      engRes.PromptTokens,
		CompletionTokens:  engRes.CompletionTokens,
		TotalTokens:       engRes.TotalTokens,
	}

	res, err = s.repo.CreateResult(ctx, res)
	if err != nil {
		return nil, err
	}

	req.Status = AnalysisRequestStatusCompleted
	completed := time.Now()
	req.CompletedAt = &completed
	_, _ = s.repo.UpdateRequest(ctx, req)

	err = s.keyService.DeductTokens(ctx, apiKey, uint64(engRes.TotalTokens))
	if err != nil {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventAPIKeyUsed,
			Description: "failed to deduct tokens",
			UserID:      &apiKey.UserID,
			APIKeyID:    &apiKey.ID,
		})
	}

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventAnalysisCompleted,
		Description: "successfully processed resume",
		UserID:      &apiKey.UserID,
		APIKeyID:    &apiKey.ID,
	})
	return res, nil
}

func (s *AnalysisService) ListHistory(ctx context.Context, plainTextKey string, limit, offset int) ([]*AnalysisRequest, error) {
	apiKey, err := s.keyService.ValidateKey(ctx, plainTextKey)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListRequestsByUserID(ctx, apiKey.UserID, limit, offset)
}

func (s *AnalysisService) GetResult(ctx context.Context, plainTextKey string, requestID uint64) (*AnalysisResult, error) {
	apiKey, err := s.keyService.ValidateKey(ctx, plainTextKey)
	if err != nil {
		return nil, err
	}

	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.UserID != apiKey.UserID {
		return nil, errors.New("unauthorized: request does not belong to user")
	}

	return s.repo.GetResultByRequestID(ctx, requestID)
}
