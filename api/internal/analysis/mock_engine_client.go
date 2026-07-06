package analysis

import (
	"context"
	"errors"
	"time"
)

type MockEngineScenario string

const (
	MockScenarioSuccess     MockEngineScenario = "success"
	MockScenarioTimeout     MockEngineScenario = "timeout"
	MockScenarioServerError MockEngineScenario = "server_error"
	MockScenarioRateLimited MockEngineScenario = "rate_limited"
	MockScenarioBadOutput   MockEngineScenario = "bad_output"
)

type MockEngineClient struct {
	URL      string
	Scenario MockEngineScenario
	Delay    time.Duration
}

func NewMockEngineClient(url string) *MockEngineClient {
	return &MockEngineClient{
		URL:      url,
		Scenario: MockScenarioSuccess,
		Delay:    0,
	}
}

func (m *MockEngineClient) Analyze(ctx context.Context, req *EngineRequest) (*EngineResponse, error) {
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	switch m.Scenario {
	case MockScenarioTimeout:
		return nil, context.DeadlineExceeded

	case MockScenarioServerError:
		return nil, errors.New("500 Internal Server Error: external engine crashed")

	case MockScenarioRateLimited:
		return nil, errors.New("429 Too Many Requests: quota exceeded on external provider")

	case MockScenarioBadOutput:
		return &EngineResponse{
			ResultJSON:       `{ "invalid_json": true, missing_quotes }`,
			Model:            "mock-v1-broken",
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		}, nil

	case MockScenarioSuccess:
		fallthrough
	default:
		return &EngineResponse{
			ResultJSON:       `{"score": 95, "feedback": "Excellent candidate, perfectly matches job description."}`,
			Model:            "mock-v1-success",
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		}, nil
	}
}
