package config

import (
	"errors"
	"fmt"
	"net/url"
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
	Domain                   string
	SessionDurationHours     int
	VerifyTokenDurationHours int
	ResetTokenDurationHours  int
	PaginationDefaultLimit   int
	GlobalAnalysisRPMLimit   int
	GlobalAnalysisRPDLimit   int
	WorkerConcurrency        int
	FixturesPath             string
}

func Load(getenv func(string) string) (*Config, error) {
	allowedOriginsStr := getenv("ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		return nil, errors.New("ALLOWED_ORIGINS environment variable is required")
	}
	allowedOrigins := make([]string, 0)

	for _, origin := range strings.Split(allowedOriginsStr, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins = append(allowedOrigins, origin)
		}
	}
	if len(allowedOrigins) == 0 {
		return nil, errors.New("ALLOWED_ORIGINS must contain at least one origin")
	}

	dbURL := getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	jwtSecret := getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	csrfAuthKey := getenv("CSRF_AUTH_KEY")
	if csrfAuthKey == "" {
		return nil, errors.New("CSRF_AUTH_KEY environment variable is required")
	}
	if len([]byte(csrfAuthKey)) != 32 {
		return nil, errors.New("CSRF_AUTH_KEY must be exactly 32 bytes")
	}

	port := getenv("PORT")
	if port == "" {
		return nil, errors.New("PORT environment variable is required")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return nil, errors.New("PORT must be a valid port number between 1 and 65535")
	}

	analysisServiceURL := getenv("ANALYSIS_SERVICE_URL")
	if analysisServiceURL == "" {
		return nil, errors.New("ANALYSIS_SERVICE_URL environment variable is required")
	}

	emailAPIKey := getenv("EMAIL_API_KEY")
	if emailAPIKey == "" {
		return nil, errors.New("EMAIL_API_KEY environment variable is required")
	}

	emailFrom := getenv("EMAIL_FROM")
	if emailFrom == "" {
		return nil, errors.New("EMAIL_FROM environment variable is required")
	}

	emailContact := getenv("EMAIL_CONTACT")
	if emailContact == "" {
		return nil, errors.New("EMAIL_CONTACT environment variable is required")
	}

	env := getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	domain := strings.TrimSpace(getenv("DOMAIN"))
	if domain == "" {
		return nil, errors.New("DOMAIN environment variable is required")
	}

	parsedDomain, err := url.ParseRequestURI(domain)
	if err != nil || parsedDomain.Host == "" {
		return nil, errors.New("DOMAIN must be a valid absolute URL")
	}

	if parsedDomain.Scheme != "http" && parsedDomain.Scheme != "https" {
		return nil, errors.New("DOMAIN must use http or https")
	}

	sessionDurationHours, err := getEnvAsPositiveInt(
		getenv,
		"SESSION_DURATION_HOURS",
		24,
	)
	if err != nil {
		return nil, err
	}

	verifyTokenDurationHours, err := getEnvAsPositiveInt(
		getenv,
		"VERIFY_TOKEN_DURATION_HOURS",
		24,
	)
	if err != nil {
		return nil, err
	}

	resetTokenDurationHours, err := getEnvAsPositiveInt(
		getenv,
		"RESET_TOKEN_DURATION_HOURS",
		1,
	)
	if err != nil {
		return nil, err
	}

	paginationDefaultLimit, err := getEnvAsPositiveInt(
		getenv,
		"PAGINATION_DEFAULT_LIMIT",
		50,
	)
	if err != nil {
		return nil, err
	}

	globalAnalysisRPMLimit, err := getEnvAsPositiveInt(
		getenv,
		"GLOBAL_ANALYSIS_RPM_LIMIT",
		1,
	)
	if err != nil {
		return nil, err
	}

	globalAnalysisRPDLimit, err := getEnvAsPositiveInt(
		getenv,
		"GLOBAL_ANALYSIS_RPD_LIMIT",
		6,
	)
	if err != nil {
		return nil, err
	}

	workerConcurrency, err := getEnvAsPositiveInt(
		getenv,
		"WORKER_CONCURRENCY",
		4,
	)
	if err != nil {
		return nil, err
	}

	fixturesPath := getenv("FIXTURES_PATH")
	if fixturesPath == "" {
		fixturesPath = "fixtures/seeds.json"
	}

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
		Domain:                   domain,
		SessionDurationHours:     sessionDurationHours,
		VerifyTokenDurationHours: verifyTokenDurationHours,
		ResetTokenDurationHours:  resetTokenDurationHours,
		PaginationDefaultLimit:   paginationDefaultLimit,
		GlobalAnalysisRPMLimit:   globalAnalysisRPMLimit,
		GlobalAnalysisRPDLimit:   globalAnalysisRPDLimit,
		WorkerConcurrency:        workerConcurrency,
		FixturesPath:             fixturesPath,
	}, nil
}

func getEnvAsPositiveInt(
	getenv func(string) string,
	name string,
	defaultValue int,
) (int, error) {
	valStr := getenv(name)
	if valStr == "" {
		return defaultValue, nil
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", name)
	}

	if val <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}

	return val, nil
}

func (c *Config) IsDevelopment() bool {
	return strings.EqualFold(c.Environment, "development") ||
		strings.EqualFold(c.Environment, "dev")
}
