package analysis_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/analysis"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
)

func assertEqual(t *testing.T, got, want any, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

func TestProcessResume(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		plainTextKey   string
		resumeText     string
		jobDescription string
		mockEngine     analysis.MockEngineScenario
		mockAPIKey     func(ctx context.Context, plainTextKey string) (*apikey.APIKey, error)
		wantError      error
		wantTokens     uint32
	}{
		{
			name:           "success path",
			plainTextKey:   "valid-key",
			resumeText:     "go engineer",
			jobDescription: "need go engineer",
			mockEngine:     analysis.MockScenarioSuccess,
			wantError:      nil,
			wantTokens:     150,
		},
		{
			name:           "invalid api key",
			plainTextKey:   "invalid-key",
			resumeText:     "go engineer",
			jobDescription: "need go engineer",
			mockEngine:     analysis.MockScenarioSuccess,
			mockAPIKey: func(ctx context.Context, plainTextKey string) (*apikey.APIKey, error) {
				return nil, errors.New("invalid key")
			},
			wantError:  errors.New("invalid key"),
			wantTokens: 0,
		},
		{
			name:           "insufficient token quota",
			plainTextKey:   "valid-key",
			resumeText:     "go engineer",
			jobDescription: "need go engineer",
			mockEngine:     analysis.MockScenarioSuccess,
			mockAPIKey: func(ctx context.Context, plainTextKey string) (*apikey.APIKey, error) {

				return &apikey.APIKey{ID: 1, UserID: 1, TokenQuota: 100, TokensUsed: 100}, nil
			},
			wantError:  analysis.ErrInsufficientQuota,
			wantTokens: 0,
		},
		{
			name:           "engine server error",
			plainTextKey:   "valid-key",
			resumeText:     "go engineer",
			jobDescription: "need go engineer",
			mockEngine:     analysis.MockScenarioServerError,
			wantError:      analysis.ErrAnalysisFailed,
			wantTokens:     0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &analysis.MockRepository{}
			auditSvc := &analysis.MockAuditService{}

			keySvc := &analysis.MockAPIKeyValidator{}
			if tt.mockAPIKey != nil {
				keySvc.ValidateKeyFunc = tt.mockAPIKey
			}

			engine := analysis.NewMockEngineClient("http://mock")
			engine.Scenario = tt.mockEngine

			svc := analysis.NewAnalysisService(repo, auditSvc, keySvc, engine, 50)

			res, err := svc.ProcessResume(context.Background(), tt.plainTextKey, "resume.pdf", "jd.pdf", []byte(tt.resumeText), []byte(tt.jobDescription))

			if tt.wantError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantError)
				}
				assertEqual(t, err.Error(), tt.wantError.Error(), "error message mismatch")
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if res == nil {
				t.Fatal("expected result, got nil")
			}
			assertEqual(t, res.TotalTokens, tt.wantTokens, "total tokens mismatch")
		})
	}
}

func BenchmarkProcessResume(b *testing.B) {
	b.ReportAllocs()

	repo := &analysis.MockRepository{}
	auditSvc := &analysis.MockAuditService{}
	keySvc := &analysis.MockAPIKeyValidator{}
	engine := analysis.NewMockEngineClient("http://mock")

	svc := analysis.NewAnalysisService(repo, auditSvc, keySvc, engine, 50)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := svc.ProcessResume(ctx, "valid-key", "resume.pdf", "jd.pdf", []byte("resume data"), []byte("job description"))
		_, _ = res, err
	}
}

func FuzzProcessResume(f *testing.F) {
	f.Add("key1", "resume data 1", "job 1")
	f.Add("key2", "", "")
	f.Add("", "very long string", "even longer string")

	repo := &analysis.MockRepository{}
	auditSvc := &analysis.MockAuditService{}
	keySvc := &analysis.MockAPIKeyValidator{}
	engine := analysis.NewMockEngineClient("http://mock")
	svc := analysis.NewAnalysisService(repo, auditSvc, keySvc, engine, 50)

	f.Fuzz(func(t *testing.T, plainTextKey, resumeText, jobDescription string) {

		_, _ = svc.ProcessResume(context.Background(), plainTextKey, "resume.pdf", "jd.pdf", []byte(resumeText), []byte(jobDescription))
	})
}
