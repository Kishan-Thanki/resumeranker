package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kishan-thanki/logger/v2/slogctx"
	"github.com/kishan-thanki/logger/v2/slogredact"
	"github.com/kishan-thanki/resumeranker/api/internal/analysis"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/auth"
	"github.com/kishan-thanki/resumeranker/api/internal/config"
	"github.com/kishan-thanki/resumeranker/api/internal/database"
	"github.com/kishan-thanki/resumeranker/api/internal/email"
	"github.com/kishan-thanki/resumeranker/api/internal/server"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

func main() {
	baseLogger := slog.NewJSONHandler(os.Stdout, nil)

	safeLogger := slogredact.NewHandler(baseLogger,
		"password", "old_password", "new_password",
		"token", "jwt", "csrf_token", "X-CSRF-Token",
		"plainTextKey", "api_key",
		"Authorization", "Cookie", "Set-Cookie",
	)

	ctxLogger := slogctx.NewHandler(safeLogger)

	slog.SetDefault(slog.New(ctxLogger))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", slog.Any("error", err))
		os.Exit(1)
	}

	userRepo := users.NewPostgresRepository(pool)
	apiKeyRepo := apikey.NewPostgresRepository(pool)
	analysisRepo := analysis.NewPostgresRepository(pool)
	auditRepo := audit.NewPostgresRepository(pool)

	emailService := email.NewResendService(cfg.EmailAPIKey, cfg.EmailFrom)

	auditService := audit.NewAuditService(auditRepo)
	userService := users.NewUserService(userRepo, auditService, emailService)

	if err := userService.SeedFromFixtures(ctx, "fixtures/seeds.json"); err != nil {
		slog.Error("failed to seed from fixtures", "err", err)
	}

	apiKeyService := apikey.NewAPIKeyService(apiKeyRepo, auditService, emailService)

	// TODO: Replace with real gRPC client to external Analysis Service
	engineClient := analysis.NewMockEngineClient(cfg.AnalysisServiceURL)
	analysisService := analysis.NewAnalysisService(analysisRepo, auditService, apiKeyService, engineClient)

	authManager := auth.NewManager(cfg.JWTSecret, cfg.Environment)

	userHandler := users.NewUserHandler(userService, authManager)
	apiKeyHandler := apikey.NewAPIKeyHandler(apiKeyService)
	analysisHandler := analysis.NewAnalysisHandler(analysisService)
	auditHandler := audit.NewAuditHandler(auditService)

	router := server.NewRouter(server.RouterConfig{
		Environment:     cfg.Environment,
		UserHandler:     userHandler,
		APIKeyHandler:   apiKeyHandler,
		AnalysisHandler: analysisHandler,
		AuditHandler:    auditHandler,
		JWTSecret:       cfg.JWTSecret,
		CSRFAuthKey:     cfg.CSRFAuthKey,
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		slog.Info("Starting API server", slog.String("port", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	slog.Info("Initiating graceful shutdown...", slog.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Duration(time.Second))
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown early", slog.Any("error", err))
	} else {
		slog.Info("HTTP server stopped accepting connections")
	}

	pool.Close()
	slog.Info("Database connection pool closed. Shutdown complete.")
}
