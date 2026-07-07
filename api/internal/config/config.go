package config

import (
	"errors"
	"os"
)

type Config struct {
	Environment        string
	DatabaseURL        string
	JWTSecret          string
	CSRFAuthKey        string
	Port               string
	AnalysisServiceURL string
	EmailAPIKey        string
	EmailFrom          string
	EmailContact       string
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

	csrfAuthKey := os.Getenv("CSRF_AUTH_KEY")
	if csrfAuthKey == "" {
		return nil, errors.New("CSRF_AUTH_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, errors.New("PORT environment variable is required")
	}

	analysisServiceURL := os.Getenv("ANALYSIS_SERVICE_URL")
	if analysisServiceURL == "" {
		return nil, errors.New("ANALYSIS_SERVICE_URL environment variable is required")
	}

	emailAPIKey := os.Getenv("EMAIL_API_KEY")
	if emailAPIKey == "" {
		return nil, errors.New("EMAIL_API_KEY environment variable is required")
	}

	emailFrom := os.Getenv("EMAIL_FROM")
	if emailFrom == "" {
		return nil, errors.New("EMAIL_FROM environment variable is required")
	}

	emailContact := os.Getenv("EMAIL_CONTACT")
	if emailContact == "" {
		return nil, errors.New("EMAIL_CONTACT environment variable is required")
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		Environment:        env,
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		CSRFAuthKey:        csrfAuthKey,
		Port:               port,
		AnalysisServiceURL: analysisServiceURL,
		EmailAPIKey:        emailAPIKey,
		EmailFrom:          emailFrom,
		EmailContact:       emailContact,
	}, nil
}
