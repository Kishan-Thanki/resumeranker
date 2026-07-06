package config

import (
	"errors"
	"os"
)

type Config struct {
	Environment        string
	DatabaseURL        string
	JWTSecret          string
	Port               string
	AnalysisServiceURL string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	analysisServiceURL := os.Getenv("ANALYSIS_SERVICE_URL")
	if analysisServiceURL == "" {
		analysisServiceURL = "http://localhost:5000"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		Environment:        env,
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		Port:               port,
		AnalysisServiceURL: analysisServiceURL,
	}, nil
}
