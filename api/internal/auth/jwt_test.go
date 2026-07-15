package auth_test

import (
	"testing"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/auth"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	t.Parallel()

	const secret = "testsecret123"

	userID := uint64(42)
	role := "admin"

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		token, err := auth.GenerateToken(userID, role, secret, time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		claims, err := auth.ValidateToken(token, secret)
		if err != nil {
			t.Fatalf("unexpected error validating token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("expected user ID %d, got %d", userID, claims.UserID)
		}

		if claims.Role != role {
			t.Errorf("expected role %q, got %q", role, claims.Role)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()

		token, err := auth.GenerateToken(userID, role, secret, -time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		_, err = auth.ValidateToken(token, secret)
		if err != auth.ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		t.Parallel()

		token, err := auth.GenerateToken(userID, role, secret, time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		_, err = auth.ValidateToken(token, "wrongsecret")
		if err != auth.ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		t.Parallel()

		_, err := auth.ValidateToken("not.a.valid.jwt", secret)
		if err != auth.ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}
