package users

import "testing"

func TestRoleIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		role  Role
		valid bool
	}{
		{name: "owner", role: RoleOwner, valid: true},
		{name: "admin", role: RoleAdmin, valid: true},
		{name: "user", role: RoleUser, valid: true},
		{name: "empty", role: "", valid: false},
		{name: "unknown", role: "unknown", valid: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.role.IsValid(); got != tt.valid {
				t.Fatalf("Role.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestAccountStatusIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status AccountStatus
		valid  bool
	}{
		{name: "pending", status: AccountStatusPending, valid: true},
		{name: "active", status: AccountStatusActive, valid: true},
		{name: "suspended", status: AccountStatusSuspended, valid: true},
		{name: "empty", status: "", valid: false},
		{name: "unknown", status: "unknown", valid: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.IsValid(); got != tt.valid {
				t.Fatalf("AccountStatus.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestAgreementTypeIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		kind  AgreementType
		valid bool
	}{
		{
			name:  "terms of service",
			kind:  AgreementTypeTermsOfService,
			valid: true,
		},
		{
			name:  "privacy policy",
			kind:  AgreementTypePrivacyPolicy,
			valid: true,
		},
		{
			name:  "empty",
			kind:  "",
			valid: false,
		},
		{
			name:  "unknown",
			kind:  "unknown",
			valid: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.kind.IsValid(); got != tt.valid {
				t.Fatalf("AgreementType.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}
