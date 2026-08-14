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

		cfg, err := config.Load(os.Getenv)
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

		_, err := config.Load(os.Getenv)
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

		_, err := config.Load(os.Getenv)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "DATABASE_URL environment variable is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("error when integer configuration is invalid", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		os.Setenv("SESSION_DURATION_HOURS", "not-a-number")

		_, err := config.Load(os.Getenv)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "SESSION_DURATION_HOURS must be a valid integer" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("error when integer configuration is not positive", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		os.Setenv("WORKER_CONCURRENCY", "0")

		_, err := config.Load(os.Getenv)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "WORKER_CONCURRENCY must be greater than zero" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("trims and ignores empty allowed origins", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		os.Setenv(
			"ALLOWED_ORIGINS",
			" http://localhost:3000, https://example.com, ",
		)

		cfg, err := config.Load(os.Getenv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.AllowedOrigins) != 2 {
			t.Fatalf("expected 2 origins, got %d", len(cfg.AllowedOrigins))
		}

		if cfg.AllowedOrigins[0] != "http://localhost:3000" {
			t.Errorf("unexpected first origin: %q", cfg.AllowedOrigins[0])
		}

		if cfg.AllowedOrigins[1] != "https://example.com" {
			t.Errorf("unexpected second origin: %q", cfg.AllowedOrigins[1])
		}
	})

	t.Run("error when CSRF_AUTH_KEY is not exactly 32 bytes", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		os.Setenv("CSRF_AUTH_KEY", "too-short")

		_, err := config.Load(os.Getenv)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "CSRF_AUTH_KEY must be exactly 32 bytes" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("error when PORT is invalid", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		os.Setenv("PORT", "99999")

		_, err := config.Load(os.Getenv)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "PORT must be a valid port number between 1 and 65535" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("error when ALLOWED_ORIGINS contains no valid origins", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		os.Setenv("ALLOWED_ORIGINS", " , , ")

		_, err := config.Load(os.Getenv)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "ALLOWED_ORIGINS must contain at least one origin" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("error when DOMAIN is not a valid URL", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		os.Setenv("DOMAIN", "not-a-url")

		_, err := config.Load(os.Getenv)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "DOMAIN must be a valid absolute URL" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("error when DOMAIN uses unsupported scheme", func(t *testing.T) {
		clearEnv()
		setValidEnv()

		os.Setenv("DOMAIN", "ftp://example.com")

		_, err := config.Load(os.Getenv)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "DOMAIN must use http or https" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
