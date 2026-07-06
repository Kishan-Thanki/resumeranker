package analysis_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/analysis"
)

// MockAnalysisService implements the internal analysisService interface used by the handler
type MockAnalysisService struct {
	ProcessResumeFunc func(ctx context.Context, plainTextKey, resumeText, jobDescription string) (*analysis.AnalysisResult, error)
	ListHistoryFunc   func(ctx context.Context, plainTextKey string, limit, offset int) ([]*analysis.AnalysisRequest, error)
	GetResultFunc     func(ctx context.Context, plainTextKey string, requestID uint64) (*analysis.AnalysisResult, error)
}

func (m *MockAnalysisService) ProcessResume(ctx context.Context, plainTextKey, resumeText, jobDescription string) (*analysis.AnalysisResult, error) {
	if m.ProcessResumeFunc != nil {
		return m.ProcessResumeFunc(ctx, plainTextKey, resumeText, jobDescription)
	}
	return &analysis.AnalysisResult{Result: `{"score":100}`}, nil
}
func (m *MockAnalysisService) ListHistory(ctx context.Context, plainTextKey string, limit, offset int) ([]*analysis.AnalysisRequest, error) {
	if m.ListHistoryFunc != nil {
		return m.ListHistoryFunc(ctx, plainTextKey, limit, offset)
	}
	return nil, nil
}
func (m *MockAnalysisService) GetResult(ctx context.Context, plainTextKey string, requestID uint64) (*analysis.AnalysisResult, error) {
	if m.GetResultFunc != nil {
		return m.GetResultFunc(ctx, plainTextKey, requestID)
	}
	return nil, nil
}

func TestAnalysisHandler_ProcessResume(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		authHeader  string
		requestBody map[string]string
		mockSvcErr  error
		wantStatus  int
	}{
		{
			name:        "missing auth header",
			authHeader:  "",
			requestBody: map[string]string{"resume_text": "text", "job_description": "job"},
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "invalid body",
			authHeader:  "Bearer valid-key",
			requestBody: nil,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "service error",
			authHeader:  "Bearer valid-key",
			requestBody: map[string]string{"resume_text": "text", "job_description": "job"},
			mockSvcErr:  analysis.ErrInsufficientQuota,
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name:        "success",
			authHeader:  "Bearer valid-key",
			requestBody: map[string]string{"resume_text": "text", "job_description": "job"},
			wantStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := &MockAnalysisService{
				ProcessResumeFunc: func(ctx context.Context, key, resume, job string) (*analysis.AnalysisResult, error) {
					if tt.mockSvcErr != nil {
						return nil, tt.mockSvcErr
					}
					return &analysis.AnalysisResult{Result: `{"score": 90}`}, nil
				},
			}

			handler := analysis.NewAnalysisHandler(mockSvc)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte(`{bad-json`)
			}

			req := httptest.NewRequest(http.MethodPost, "/analyze/resume", bytes.NewReader(body))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			handler.ProcessResume(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("ProcessResume() status = %v, want %v", rec.Code, tt.wantStatus)
			}
		})
	}
}
