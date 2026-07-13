package config_test

import (
	"os"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/config"
)

func TestLoadConfig(t *testing.T) {

	clearEnv := func() {
		os.Clearenv()
	}

	setValidEnv := func() {
		os.Setenv("FRONTEND_URL", "http://localhost:3000")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
		os.Setenv("DATABASE_URL", "postgres://test")
		os.Setenv("JWT_SECRET", "supersecret")
		os.Setenv("CSRF_AUTH_KEY", "csrf_key_32_bytes_long_123456789")
		os.Setenv("PORT", "9090")
		os.Setenv("ANALYSIS_SERVICE_URL", "http://remote:5000")
		os.Setenv("EMAIL_API_KEY", "test_key")
		os.Setenv("EMAIL_FROM", "test@example.com")
		os.Setenv("EMAIL_CONTACT", "contact@example.com")
		os.Setenv("DOMAIN", "https://example.com")
	}

	t.Run("success with all variables set", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.DatabaseURL != "postgres://test" {
			t.Errorf("expected db url postgres://test, got %s", cfg.DatabaseURL)
		}
		if cfg.JWTSecret != "supersecret" {
			t.Errorf("expected secret supersecret, got %s", cfg.JWTSecret)
		}
		if cfg.Port != "9090" {
			t.Errorf("expected port 9090, got %s", cfg.Port)
		}
		if cfg.AnalysisServiceURL != "http://remote:5000" {
			t.Errorf("expected analysis url http://remote:5000, got %s", cfg.AnalysisServiceURL)
		}
	})

	t.Run("error when ALLOWED_ORIGINS is missing", func(t *testing.T) {
		clearEnv()
		setValidEnv()
		os.Unsetenv("ALLOWED_ORIGINS")

		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "ALLOWED_ORIGINS environment variable is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("error when DATABASE_URL is missing", func(t *testing.T) {
		clearEnv()
		setValidEnv()
		os.Unsetenv("DATABASE_URL")

		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "DATABASE_URL environment variable is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
