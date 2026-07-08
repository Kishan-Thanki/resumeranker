package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment              string
	DatabaseURL              string
	JWTSecret                string
	CSRFAuthKey              string
	Port                     string
	AnalysisServiceURL       string
	EmailAPIKey              string
	EmailFrom                string
	EmailContact             string
	AllowedOrigins           []string
	FrontendURL              string
	SessionDurationHours     int
	VerifyTokenDurationHours int
	ResetTokenDurationHours  int
	PaginationDefaultLimit   int
	BulkEmailBatchSize       int
}

func Load() (*Config, error) {

	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		return nil, errors.New("ALLOWED_ORIGINS environment variable is required")
	}
	allowedOrigins := strings.Split(allowedOriginsStr, ",")

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

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		return nil, errors.New("FRONTEND_URL environment variable is required")
	}

	sessionDurationHours := getEnvAsInt("SESSION_DURATION_HOURS", 24)

	verifyTokenDurationHours := getEnvAsInt("VERIFY_TOKEN_DURATION_HOURS", 24)

	resetTokenDurationHours := getEnvAsInt("RESET_TOKEN_DURATION_HOURS", 1)

	paginationDefaultLimit := getEnvAsInt("PAGINATION_DEFAULT_LIMIT", 50)

	bulkEmailBatchSize := getEnvAsInt("BULK_EMAIL_BATCH_SIZE", 100)

	return &Config{
		Environment:              env,
		DatabaseURL:              dbURL,
		JWTSecret:                jwtSecret,
		CSRFAuthKey:              csrfAuthKey,
		Port:                     port,
		AnalysisServiceURL:       analysisServiceURL,
		EmailAPIKey:              emailAPIKey,
		EmailFrom:                emailFrom,
		EmailContact:             emailContact,
		AllowedOrigins:           allowedOrigins,
		FrontendURL:              frontendURL,
		SessionDurationHours:     sessionDurationHours,
		VerifyTokenDurationHours: verifyTokenDurationHours,
		ResetTokenDurationHours:  resetTokenDurationHours,
		PaginationDefaultLimit:   paginationDefaultLimit,
		BulkEmailBatchSize:       bulkEmailBatchSize,
	}, nil
}

func getEnvAsInt(name string, defaultValue int) int {

	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}

	return val
}
