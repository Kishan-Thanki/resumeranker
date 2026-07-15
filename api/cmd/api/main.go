package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/logger/v2/slogctx"
	"github.com/kishan-thanki/logger/v2/slogredact"
	"github.com/kishan-thanki/resumeranker/api/internal/analysis"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/auth"
	"github.com/kishan-thanki/resumeranker/api/internal/config"
	"github.com/kishan-thanki/resumeranker/api/internal/database"
	"github.com/kishan-thanki/resumeranker/api/internal/email"
	"github.com/kishan-thanki/resumeranker/api/internal/ratelimit"
	"github.com/kishan-thanki/resumeranker/api/internal/server"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

func main() {
	if err := run(context.Background(), os.Getenv); err != nil {
		slog.Error("application stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string) error {
	setupLogger(strings.ToLower(getenv("DEBUG")) == "true" || getenv("DEBUG") == "1")

	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	engineClient, err := analysis.NewGrpcEngineClient(cfg.AnalysisServiceURL)
	if err != nil {
		return err
	}
	defer engineClient.Close()

	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr, DB: cfg.RedisDB}
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	var wg sync.WaitGroup

	router, analysisService, err := buildDependencies(ctx, cfg, pool, engineClient, asynqClient, &wg)
	if err != nil {
		return err
	}

	errs := make(chan error, 2)
	httpSrv := startHTTPServer(cfg.Port, router, errs)
	asynqSrv := startWorkerServer(redisOpt, cfg.AsynqConcurrency, analysisService, errs)

	if err := waitForShutdown(httpSrv, asynqSrv, &wg, errs); err != nil {
		return err
	}

	return nil
}

func setupLogger(debugMode bool) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	var baseLogger slog.Handler
	if debugMode {
		opts.Level = slog.LevelDebug
		baseLogger = slog.NewTextHandler(os.Stdout, opts)
	} else {
		baseLogger = slog.NewJSONHandler(os.Stdout, opts)
	}

	safeLogger := slogredact.NewHandler(baseLogger,
		"password", "old_password", "new_password",
		"token", "jwt", "csrf_token", "X-CSRF-Token",
		"plainTextKey", "api_key",
		"Authorization", "Cookie", "Set-Cookie",
	)

	ctxLogger := slogctx.NewHandler(safeLogger)
	slog.SetDefault(slog.New(ctxLogger))
}

func buildDependencies(
	ctx context.Context,
	cfg *config.Config,
	pool *pgxpool.Pool,
	engineClient analysis.EngineClient,
	asynqClient *asynq.Client,
	wg *sync.WaitGroup,
) (http.Handler, *analysis.AnalysisService, error) {

	userRepo := users.NewPostgresRepository(pool)
	apiKeyRepo := apikey.NewPostgresRepository(pool)
	analysisRepo := analysis.NewPostgresRepository(pool)
	auditRepo := audit.NewPostgresRepository(pool)

	emailService := email.NewResendService(cfg.EmailAPIKey, cfg.EmailFrom)
	auditService := audit.NewAuditService(auditRepo, cfg.PaginationDefaultLimit)
	rateLimitService, err := ratelimit.NewService(cfg.RedisAddr)
	if err != nil {
		return nil, nil, err
	}
	authManager := auth.NewManager(cfg.JWTSecret, cfg.Environment, cfg.SessionDurationHours)

	agreementService := users.NewAgreementService(userRepo, emailService, cfg)
	userService := users.NewUserService(userRepo, userRepo, auditService, emailService, cfg, wg)
	apiKeyService := apikey.NewAPIKeyService(apiKeyRepo, auditService, emailService, rateLimitService, cfg.Domain, cfg.EmailContact, wg)
	analysisService := analysis.NewAnalysisService(
		analysisRepo,
		auditService,
		apiKeyService,
		rateLimitService,
		engineClient,
		cfg.PaginationDefaultLimit,
		cfg.GlobalAnalysisRPMLimit,
		cfg.GlobalAnalysisRPDLimit,
		asynqClient,
	)

	if cfg.IsDevelopment() {
		b, err := os.ReadFile(cfg.FixturesPath)
		if err != nil {
			slog.Error("failed to read fixtures file", "err", err)
		} else {
			var f users.Fixtures
			if err := json.Unmarshal(b, &f); err != nil {
				slog.Error("failed to parse fixtures file", "err", err)
			} else {
				if err := users.SeedFromFixtures(ctx, userService, agreementService, cfg, &f); err != nil {
					slog.Error("failed to seed from fixtures", "err", err)
				}
			}
		}
	}

	userHandler := users.NewUserHandler(userService, authManager, cfg.PaginationDefaultLimit)
	agreementHandler := users.NewAgreementHandler(agreementService, userRepo)
	apiKeyHandler := apikey.NewAPIKeyHandler(apiKeyService)
	analysisHandler := analysis.NewAnalysisHandler(analysisService, cfg.PaginationDefaultLimit)
	auditHandler := audit.NewAuditHandler(auditService, cfg.PaginationDefaultLimit)

	router := server.NewRouter(server.RouterConfig{
		Environment:      cfg.Environment,
		UserHandler:      userHandler,
		AgreementHandler: agreementHandler,
		APIKeyHandler:    apiKeyHandler,
		AnalysisHandler:  analysisHandler,
		AuditHandler:     auditHandler,
		JWTSecret:        cfg.JWTSecret,
		CSRFAuthKey:      cfg.CSRFAuthKey,
		AllowedOrigins:   cfg.AllowedOrigins,
	})

	return router, analysisService, nil
}

func startHTTPServer(port string, handler http.Handler, errs chan<- error) *http.Server {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		slog.Info("Starting API server", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", slog.Any("error", err))
			errs <- err
		}
	}()

	return srv
}

func startWorkerServer(redisOpt asynq.RedisClientOpt, concurrency int, analysisService *analysis.AnalysisService, errs chan<- error) *asynq.Server {
	asynqSrv := asynq.NewServer(redisOpt, asynq.Config{Concurrency: concurrency})
	mux := asynq.NewServeMux()
	mux.Handle(analysis.TypeAnalyzeResume, analysis.NewAnalyzeResumeProcessor(analysisService))

	go func() {
		slog.Info("Starting Asynq worker server")
		if err := asynqSrv.Run(mux); err != nil {
			slog.Error("asynq server failed", "err", err)
			errs <- err
		}
	}()

	return asynqSrv
}

func waitForShutdown(srv *http.Server, asynqSrv *asynq.Server, wg *sync.WaitGroup, errs <-chan error) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var fatalErr error

	select {
	case err := <-errs:
		slog.Error("Fatal startup error received, aborting...", slog.Any("error", err))
		fatalErr = err
	case sig := <-quit:
		slog.Info("Initiating graceful shutdown...", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown early", slog.Any("error", err))
	} else {
		slog.Info("HTTP server stopped accepting connections")
	}

	asynqSrv.Shutdown()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("All background tasks finished cleanly")
	case <-shutdownCtx.Done():
		slog.Warn("Timeout exceeded waiting for background tasks, forcing exit")
	}

	return fatalErr
}
