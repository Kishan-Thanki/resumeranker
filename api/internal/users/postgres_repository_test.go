package users

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kishan-thanki/resumeranker/api/internal/users/db"
)

func TestUserFromRow(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, time.August, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	deletedAt := updatedAt.Add(time.Hour)

	verificationToken := "verification-token"
	verificationExpiresAt := updatedAt.Add(24 * time.Hour)

	passwordResetToken := "password-reset-token"
	passwordResetExpiresAt := updatedAt.Add(48 * time.Hour)

	row := db.User{
		ID:           42,
		Email:        "user@example.com",
		PasswordHash: "hashed-password",
		Role:         "admin",
		Status:       "active",
		Metadata:     []byte(`{"plan":"pro"}`),
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
		DeletedAt: pgtype.Timestamptz{
			Time:  deletedAt,
			Valid: true,
		},
		IsVerified: true,
		VerificationToken: pgtype.Text{
			String: verificationToken,
			Valid:  true,
		},
		VerificationExpiresAt: pgtype.Timestamptz{
			Time:  verificationExpiresAt,
			Valid: true,
		},
		PasswordResetToken: pgtype.Text{
			String: passwordResetToken,
			Valid:  true,
		},
		PasswordResetExpiresAt: pgtype.Timestamptz{
			Time:  passwordResetExpiresAt,
			Valid: true,
		},
	}

	got := userFromRow(row)

	if got == nil {
		t.Fatal("expected user, got nil")
	}

	if got.ID != 42 {
		t.Fatalf("expected ID 42, got %d", got.ID)
	}
	if got.Email != "user@example.com" {
		t.Fatalf("expected email %q, got %q", "user@example.com", got.Email)
	}
	if got.PasswordHash != "hashed-password" {
		t.Fatalf("expected password hash to be preserved")
	}
	if got.Role != RoleAdmin {
		t.Fatalf("expected role %q, got %q", RoleAdmin, got.Role)
	}
	if got.Status != AccountStatusActive {
		t.Fatalf("expected status %q, got %q", AccountStatusActive, got.Status)
	}
	if string(got.Metadata) != `{"plan":"pro"}` {
		t.Fatalf("expected metadata to be preserved, got %s", got.Metadata)
	}
	if !got.IsVerified {
		t.Fatal("expected user to be verified")
	}
	if got.VerificationToken == nil || *got.VerificationToken != verificationToken {
		t.Fatalf("expected verification token %q, got %v", verificationToken, got.VerificationToken)
	}
	if got.VerificationExpiresAt == nil || !got.VerificationExpiresAt.Equal(verificationExpiresAt) {
		t.Fatalf("expected verification expiry %v, got %v", verificationExpiresAt, got.VerificationExpiresAt)
	}
	if got.PasswordResetToken == nil || *got.PasswordResetToken != passwordResetToken {
		t.Fatalf("expected password reset token %q, got %v", passwordResetToken, got.PasswordResetToken)
	}
	if got.PasswordResetExpiresAt == nil || !got.PasswordResetExpiresAt.Equal(passwordResetExpiresAt) {
		t.Fatalf("expected password reset expiry %v, got %v", passwordResetExpiresAt, got.PasswordResetExpiresAt)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt %v, got %v", createdAt, got.CreatedAt)
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected UpdatedAt %v, got %v", updatedAt, got.UpdatedAt)
	}
	if got.DeletedAt == nil || !got.DeletedAt.Equal(deletedAt) {
		t.Fatalf("expected DeletedAt %v, got %v", deletedAt, got.DeletedAt)
	}
}

func TestUserFromRow_NullableFields(t *testing.T) {
	t.Parallel()
	row := db.User{
		ID:     7,
		Email:  "nullable@example.com",
		Role:   "user",
		Status: "pending",
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Date(2026, time.August, 15, 13, 0, 0, 0, time.UTC),
			Valid: true,
		},
		VerificationToken:      pgtype.Text{Valid: false},
		VerificationExpiresAt:  pgtype.Timestamptz{Valid: false},
		PasswordResetToken:     pgtype.Text{Valid: false},
		PasswordResetExpiresAt: pgtype.Timestamptz{Valid: false},
		DeletedAt:              pgtype.Timestamptz{Valid: false},
	}

	got := userFromRow(row)

	if got == nil {
		t.Fatal("expected user, got nil")
	}

	if got.VerificationToken != nil {
		t.Fatalf("expected nil verification token, got %v", got.VerificationToken)
	}
	if got.VerificationExpiresAt != nil {
		t.Fatalf("expected nil verification expiry, got %v", got.VerificationExpiresAt)
	}
	if got.PasswordResetToken != nil {
		t.Fatalf("expected nil password reset token, got %v", got.PasswordResetToken)
	}
	if got.PasswordResetExpiresAt != nil {
		t.Fatalf("expected nil password reset expiry, got %v", got.PasswordResetExpiresAt)
	}
	if got.DeletedAt != nil {
		t.Fatalf("expected nil DeletedAt, got %v", got.DeletedAt)
	}
}

func TestAgreementFromRow(t *testing.T) {
	t.Parallel()
	publishedAt := time.Date(
		2026,
		time.August,
		15,
		15,
		45,
		0,
		0,
		time.UTC,
	)

	row := db.Agreement{
		ID:      100,
		Type:    "terms_of_service",
		Version: "2.1",
		Content: "Terms content",
		PublishedAt: pgtype.Timestamptz{
			Time:  publishedAt,
			Valid: true,
		},
	}

	got := agreementFromRow(row)

	if got == nil {
		t.Fatal("expected agreement, got nil")
	}

	if got.ID != 100 {
		t.Fatalf("expected ID 100, got %d", got.ID)
	}
	if got.Type != AgreementTypeTermsOfService {
		t.Fatalf("expected type %q, got %q", AgreementTypeTermsOfService, got.Type)
	}
	if got.Version != "2.1" {
		t.Fatalf("expected version %q, got %q", "2.1", got.Version)
	}
	if got.Content != "Terms content" {
		t.Fatalf("expected content %q, got %q", "Terms content", got.Content)
	}
	if !got.PublishedAt.Equal(publishedAt) {
		t.Fatalf("expected PublishedAt %v, got %v", publishedAt, got.PublishedAt)
	}
}

func TestUserFromRow_PreservesMetadataBytes(t *testing.T) {
	t.Parallel()
	metadata := []byte(`{"roles":["admin"],"enabled":true}`)

	row := db.User{
		ID:       9,
		Email:    "metadata@example.com",
		Metadata: metadata,
		Role:     "admin",
		Status:   "active",
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Date(2026, time.August, 15, 16, 0, 0, 0, time.UTC),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Date(2026, time.August, 15, 16, 30, 0, 0, time.UTC),
			Valid: true,
		},
	}

	got := userFromRow(row)

	if string(got.Metadata) != string(metadata) {
		t.Fatalf("expected metadata %q, got %q", metadata, got.Metadata)
	}
}

func TestAgreementTypeConversion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		row  string
		want AgreementType
	}{
		{
			name: "terms",
			row:  "terms_of_service",
			want: AgreementTypeTermsOfService,
		},
		{
			name: "privacy",
			row:  "privacy_policy",
			want: AgreementTypePrivacyPolicy,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := AgreementType(tt.row)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
