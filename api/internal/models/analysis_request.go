package models

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
	ID          uint64                `json:"id"`
	RequestID   string                `json:"request_id"`
	UserID      uint64                `json:"user_id"`
	APIKeyID    uint64                `json:"api_key_id"`
	Status      AnalysisRequestStatus `json:"status"`
	Error       *string               `json:"error"`
	Metadata    json.RawMessage       `json:"metadata,omitempty"`
	StartedAt   *time.Time            `json:"started_at"`
	CompletedAt *time.Time            `json:"completed_at"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   *time.Time            `json:"deleted_at"`
}
