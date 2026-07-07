package analysis_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/analysis"
)

type MockAnalysisService struct {
	ProcessResumeFunc func(ctx context.Context, plainTextKey string, resumeFilename, jdFilename string, resumePDF, jdPDF []byte) (*analysis.AnalysisResult, error)
	ListHistoryFunc   func(ctx context.Context, plainTextKey string, limit, offset int) ([]*analysis.AnalysisRequest, error)
	GetResultFunc     func(ctx context.Context, plainTextKey string, requestID string) (*analysis.AnalysisResult, error)
}

func (m *MockAnalysisService) ProcessResume(ctx context.Context, plainTextKey string, resumeFilename, jdFilename string, resumePDF, jdPDF []byte) (*analysis.AnalysisResult, error) {
	if m.ProcessResumeFunc != nil {
		return m.ProcessResumeFunc(ctx, plainTextKey, resumeFilename, jdFilename, resumePDF, jdPDF)
	}
	return &analysis.AnalysisResult{Result: `{"score":100}`}, nil
}
func (m *MockAnalysisService) ListHistory(ctx context.Context, plainTextKey string, limit, offset int) ([]*analysis.AnalysisRequest, error) {
	if m.ListHistoryFunc != nil {
		return m.ListHistoryFunc(ctx, plainTextKey, limit, offset)
	}
	return nil, nil
}
func (m *MockAnalysisService) GetResult(ctx context.Context, plainTextKey string, requestID string) (*analysis.AnalysisResult, error) {
	if m.GetResultFunc != nil {
		return m.GetResultFunc(ctx, plainTextKey, requestID)
	}
	return nil, nil
}

func TestAnalysisHandler_ProcessResume(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		authHeader string
		setupForm  func() (*bytes.Buffer, string)
		mockSvcErr error
		wantStatus int
	}{
		{
			name:       "missing auth header",
			authHeader: "",
			setupForm: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body, writer.FormDataContentType()
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing files",
			authHeader: "Bearer valid-key",
			setupForm: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body, writer.FormDataContentType()
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "success",
			authHeader: "Bearer valid-key",
			setupForm: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				importTextproto := true
				_ = importTextproto

				h1 := make(map[string][]string)
				h1["Content-Disposition"] = []string{`form-data; name="resume"; filename="resume.pdf"`}
				h1["Content-Type"] = []string{"application/pdf"}
				part1, _ := writer.CreatePart(h1)
				part1.Write([]byte("fake-resume-pdf-data"))

				h2 := make(map[string][]string)
				h2["Content-Disposition"] = []string{`form-data; name="job_description"; filename="jd.pdf"`}
				h2["Content-Type"] = []string{"application/pdf"}
				part2, _ := writer.CreatePart(h2)
				part2.Write([]byte("fake-jd-pdf-data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := &MockAnalysisService{
				ProcessResumeFunc: func(ctx context.Context, key string, resumeFilename, jdFilename string, resume, job []byte) (*analysis.AnalysisResult, error) {
					if tt.mockSvcErr != nil {
						return nil, tt.mockSvcErr
					}
					return &analysis.AnalysisResult{Result: `{"score": 90}`}, nil
				},
			}

			handler := analysis.NewAnalysisHandler(mockSvc)

			body, contentType := tt.setupForm()
			req := httptest.NewRequest(http.MethodPost, "/analyze/resume", body)
			req.Header.Set("Content-Type", contentType)

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
