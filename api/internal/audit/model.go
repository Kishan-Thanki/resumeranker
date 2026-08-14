package audit

import "time"

type AuditEventType string

const (
	AuditEventUserRegistered      AuditEventType = "user_registered"
	AuditEventUserVerified        AuditEventType = "user_verified"
	AuditEventUserLoggedIn        AuditEventType = "user_logged_in"
	AuditEventUserPasswordChanged AuditEventType = "user_password_changed"
	AuditEventAPIKeyCreated       AuditEventType = "api_key_created"
	AuditEventAPIKeyRevoked       AuditEventType = "api_key_revoked"
	AuditEventAPIKeyUsed          AuditEventType = "api_key_used"

	AuditEventAgreementAccepted AuditEventType = "agreement_accepted"

	AuditEventAnalysisRequested AuditEventType = "analysis_requested"
	AuditEventAnalysisCompleted AuditEventType = "analysis_completed"
	AuditEventAnalysisFailed    AuditEventType = "analysis_failed"
)

func (t AuditEventType) IsValid() bool {
	switch t {
	case
		AuditEventUserRegistered,
		AuditEventUserVerified,
		AuditEventUserLoggedIn,
		AuditEventUserPasswordChanged,
		AuditEventAPIKeyCreated,
		AuditEventAPIKeyRevoked,
		AuditEventAPIKeyUsed,
		AuditEventAgreementAccepted,
		AuditEventAnalysisRequested,
		AuditEventAnalysisCompleted,
		AuditEventAnalysisFailed:
		return true
	}
	return false
}

type AuditEvent struct {
	ID                uint64  `json:"id"`
	UserID            *uint64 `json:"user_id"`
	APIKeyID          *uint64 `json:"api_key_id"`
	AnalysisRequestID *uint64 `json:"analysis_request_id"`

	Type        AuditEventType `json:"type"`
	Description string         `json:"description"`

	IPAddress *string `json:"ip_address"`
	UserAgent *string `json:"user_agent"`

	CreatedAt time.Time `json:"created_at"`
}
