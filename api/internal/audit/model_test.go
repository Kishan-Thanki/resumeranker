package audit

import "testing"

func TestAuditEventTypeIsValid(t *testing.T) {
	validTypes := []AuditEventType{
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
		AuditEventAnalysisFailed,
	}

	for _, eventType := range validTypes {
		t.Run(string(eventType), func(t *testing.T) {
			if !eventType.IsValid() {
				t.Errorf("expected %q to be valid", eventType)
			}
		})
	}
}

func TestAuditEventTypeIsValidRejectsUnknown(t *testing.T) {
	unknownTypes := []AuditEventType{
		"",
		"unknown",
		"user_deleted",
		"analysis_started",
	}

	for _, eventType := range unknownTypes {
		t.Run(string(eventType), func(t *testing.T) {
			if eventType.IsValid() {
				t.Errorf("expected %q to be invalid", eventType)
			}
		})
	}
}
