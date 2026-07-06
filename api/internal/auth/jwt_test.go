package auth_test

import (
	"testing"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/auth"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	t.Parallel()

	secret := "testsecret123"
	userID := uint64(42)
	role := "admin"

	t.Run("success", func(t *testing.T) {
		token, err := auth.GenerateToken(userID, role, secret, time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		gotID, gotRole, err := auth.ValidateToken(token, secret)
		if err != nil {
			t.Fatalf("unexpected error validating token: %v", err)
		}

		if gotID != userID {
			t.Errorf("expected user ID %d, got %d", userID, gotID)
		}
		if gotRole != role {
			t.Errorf("expected role %s, got %s", role, gotRole)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token, err := auth.GenerateToken(userID, role, secret, -time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		_, _, err = auth.ValidateToken(token, secret)
		if err != auth.ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		token, err := auth.GenerateToken(userID, role, secret, time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		_, _, err = auth.ValidateToken(token, "wrongsecret")
		if err != auth.ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		_, _, err := auth.ValidateToken("not.a.valid.jwt", secret)
		if err != auth.ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}
