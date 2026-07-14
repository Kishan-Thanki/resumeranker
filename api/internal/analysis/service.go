package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
)

var (
	ErrInsufficientQuota = errors.New("insufficient token quota")
	ErrAnalysisFailed    = errors.New("analysis failed")
	ErrRateLimit         = errors.New("the AI service is currently experiencing high load. Please try again later")
	ErrAPIConnection     = errors.New("failed to connect to the AI service provider")
	ErrContextExceeded   = errors.New("the provided documents are too large for the AI to process")
	ErrLLMValidation     = errors.New("the AI service returned an invalid format")
	ErrLLMTimeout        = errors.New("the AI service took too long to respond")
	ErrPDFParse          = errors.New("failed to parse the provided PDF document")
)

type APIKeyValidator interface {
	ValidateKey(ctx context.Context, plainTextKey string) (*apikey.APIKey, error)
	DeductTokens(ctx context.Context, apiKey *apikey.APIKey, tokensUsed uint64) error
}

type auditService interface {
	LogEvent(ctx context.Context, event *audit.AuditEvent) error
}

type RateLimiter interface {
	CheckGlobalLimit(ctx context.Context, rpmLimit, rpdLimit int) error
	CheckKeyLimit(ctx context.Context, apiKeyID uint64, rpmLimit, rpdLimit int) error
}

type EngineRequest struct {
	ResumePDF         []byte `json:"resume_pdf"`
	JobDescriptionPDF []byte `json:"job_description_pdf"`
	RequestID         string `json:"request_id"`
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
	repo           Repository
	auditService   auditService
	keyService     APIKeyValidator
	rateLimiter    RateLimiter
	engineClient   EngineClient
	defaultLimit   int
	globalRPMLimit int
	globalRPDLimit int
}

func NewAnalysisService(
	repo Repository,
	auditService auditService,
	keyService APIKeyValidator,
	rateLimiter RateLimiter,
	engineClient EngineClient,
	defaultLimit int,
	globalRPMLimit int,
	globalRPDLimit int,
) *AnalysisService {
	return &AnalysisService{
		repo:           repo,
		auditService:   auditService,
		keyService:     keyService,
		rateLimiter:    rateLimiter,
		engineClient:   engineClient,
		defaultLimit:   defaultLimit,
		globalRPMLimit: globalRPMLimit,
		globalRPDLimit: globalRPDLimit,
	}
}

func (s *AnalysisService) ProcessResume(ctx context.Context, plainTextKey string, resumeFilename, jdFilename string, resumePDF, jdPDF []byte) (*AnalysisResult, error) {
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

	if err := s.rateLimiter.CheckGlobalLimit(ctx, s.globalRPMLimit, s.globalRPDLimit); err != nil {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventAPIKeyUsed,
			Description: "global rate limit exceeded",
			UserID:      &apiKey.UserID,
			APIKeyID:    &apiKey.ID,
		})
		return nil, ErrRateLimit
	}

	if err := s.rateLimiter.CheckKeyLimit(ctx, apiKey.ID, int(apiKey.RequestsPerMinute), int(apiKey.RequestsPerDay)); err != nil {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventAPIKeyUsed,
			Description: "api key rate limit exceeded",
			UserID:      &apiKey.UserID,
			APIKeyID:    &apiKey.ID,
		})
		return nil, ErrRateLimit
	}

	now := time.Now()
	requestID := uuid.New().String()

	metadataBytes, _ := json.Marshal(map[string]string{
		"resume_filename": resumeFilename,
		"jd_filename":     jdFilename,
	})

	req := &AnalysisRequest{
		RequestID: requestID,
		UserID:    apiKey.UserID,
		APIKeyID:  apiKey.ID,
		Status:    AnalysisRequestStatusProcessing,
		Metadata:  json.RawMessage(metadataBytes),
		StartedAt: &now,
	}

	req, err = s.repo.CreateRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	engReq := &EngineRequest{
		ResumePDF:         resumePDF,
		JobDescriptionPDF: jdPDF,
		RequestID:         fmt.Sprintf("%d", req.ID),
	}

	engRes, err := s.engineClient.Analyze(ctx, engReq)
	if err != nil {
		errStr := err.Error()
		var retErr error = ErrAnalysisFailed

		if strings.Contains(errStr, "ERR_RATE_LIMIT") {
			retErr = ErrRateLimit
			errStr = ErrRateLimit.Error()
		} else if strings.Contains(errStr, "ERR_API_CONNECTION") {
			retErr = ErrAPIConnection
			errStr = ErrAPIConnection.Error()
		} else if strings.Contains(errStr, "ERR_CONTEXT_EXCEEDED") {
			retErr = ErrContextExceeded
			errStr = ErrContextExceeded.Error()
		} else if strings.Contains(errStr, "ERR_LLM_VALIDATION") {
			retErr = ErrLLMValidation
			errStr = ErrLLMValidation.Error()
		} else if strings.Contains(errStr, "ERR_LLM_TIMEOUT") {
			retErr = ErrLLMTimeout
			errStr = ErrLLMTimeout.Error()
		} else if strings.Contains(errStr, "ERR_PDF_PARSE") {
			retErr = ErrPDFParse
			errStr = ErrPDFParse.Error()
		}

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
		return nil, retErr
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
		req.Status = AnalysisRequestStatusFailed
		errStr := "failed to save analysis result to database"
		req.Error = &errStr
		_, _ = s.repo.UpdateRequest(ctx, req)
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventAnalysisFailed,
			Description: "failed to save analysis result",
			UserID:      &apiKey.UserID,
			APIKeyID:    &apiKey.ID,
		})
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
		limit = s.defaultLimit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListRequestsByUserID(ctx, apiKey.UserID, limit, offset)
}

func (s *AnalysisService) GetResult(ctx context.Context, plainTextKey string, requestID string) (*AnalysisResult, error) {
	apiKey, err := s.keyService.ValidateKey(ctx, plainTextKey)
	if err != nil {
		return nil, err
	}

	req, err := s.repo.GetRequestByUUID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.UserID != apiKey.UserID {
		return nil, errors.New("unauthorized: request does not belong to user")
	}

	return s.repo.GetResultByRequestID(ctx, req.ID)
}
