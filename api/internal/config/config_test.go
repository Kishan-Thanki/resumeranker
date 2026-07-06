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

	t.Run("success with all variables set", func(t *testing.T) {
		clearEnv()
		os.Setenv("DATABASE_URL", "postgres://test")
		os.Setenv("JWT_SECRET", "supersecret")
		os.Setenv("PORT", "9090")
		os.Setenv("ANALYSIS_SERVICE_URL", "http://remote:5000")

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

	t.Run("success with default fallbacks", func(t *testing.T) {
		clearEnv()
		os.Setenv("DATABASE_URL", "postgres://test")
		os.Setenv("JWT_SECRET", "supersecret")

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Port != "8080" {
			t.Errorf("expected default port 8080, got %s", cfg.Port)
		}
		if cfg.AnalysisServiceURL != "http://localhost:5000" {
			t.Errorf("expected default analysis url http://localhost:5000, got %s", cfg.AnalysisServiceURL)
		}
	})

	t.Run("error when DATABASE_URL is missing", func(t *testing.T) {
		clearEnv()
		os.Setenv("JWT_SECRET", "supersecret")

		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "DATABASE_URL environment variable is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("error when JWT_SECRET is missing", func(t *testing.T) {
		clearEnv()
		os.Setenv("DATABASE_URL", "postgres://test")

		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "JWT_SECRET environment variable is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
