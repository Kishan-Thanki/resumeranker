package apikey

import "testing"

func TestAPIKeyStatusIsValid(t *testing.T) {
	validStatuses := []APIKeyStatus{
		APIKeyStatusActive,
		APIKeyStatusInactive,
		APIKeyStatusSuspended,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			if !status.IsValid() {
				t.Errorf("expected %q to be valid", status)
			}
		})
	}
}

func TestAPIKeyStatusIsValidRejectsUnknown(t *testing.T) {
	invalidStatuses := []APIKeyStatus{
		"",
		"unknown",
		"revoked",
		"deleted",
		"expired",
	}

	for _, status := range invalidStatuses {
		t.Run(string(status), func(t *testing.T) {
			if status.IsValid() {
				t.Errorf("expected %q to be invalid", status)
			}
		})
	}
}
