package domain

import "time"

type AnalysisResult struct {
	ID                uint64    `json:"id"`
	AnalysisRequestID uint64    `json:"analysis_request_id"`
	Model             string    `json:"model"`
	Result            string    `json:"result"`
	PromptTokens      uint32    `json:"-"`
	CompletionTokens  uint32    `json:"-"`
	TotalTokens       uint32     `json:"-"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
}
