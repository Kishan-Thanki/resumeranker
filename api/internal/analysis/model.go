package analysis

import (
	"encoding/json"
	"time"
)

type AnalysisRequestStatus string

const (
	AnalysisRequestStatusQueued     AnalysisRequestStatus = "queued"
	AnalysisRequestStatusProcessing AnalysisRequestStatus = "processing"
	AnalysisRequestStatusCompleted  AnalysisRequestStatus = "completed"
	AnalysisRequestStatusFailed     AnalysisRequestStatus = "failed"
)

type AnalysisRequest struct {
	ID          uint64                `json:"-"`
	RequestID   string                `json:"request_id"`
	UserID      uint64                `json:"-"`
	APIKeyID    uint64                `json:"-"`
	Status      AnalysisRequestStatus `json:"status"`
	Error       *string               `json:"error"`
	Metadata    json.RawMessage       `json:"metadata,omitempty"`
	TotalTokens *uint32               `json:"total_tokens,omitempty"`
	StartedAt   *time.Time            `json:"started_at"`
	CompletedAt *time.Time            `json:"completed_at"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"-"`
	DeletedAt   *time.Time            `json:"-"`
}

type AnalysisResult struct {
	ID                uint64     `json:"-"`
	AnalysisRequestID uint64     `json:"-"`
	Model             string     `json:"-"`
	Result            string     `json:"result"`
	PromptTokens      uint32     `json:"-"`
	CompletionTokens  uint32     `json:"-"`
	TotalTokens       uint32     `json:"total_tokens"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"-"`
	DeletedAt         *time.Time `json:"-"`
}
