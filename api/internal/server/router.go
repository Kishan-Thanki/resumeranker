package server

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kishan-thanki/logger/v2/httptelemetry"
	"github.com/kishan-thanki/resumeranker/api/internal/analysis"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/auth"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

type RouterConfig struct {
	UserHandler     *users.UserHandler
	APIKeyHandler   *apikey.APIKeyHandler
	AnalysisHandler *analysis.AnalysisHandler
	AuditHandler    *audit.AuditHandler
	JWTSecret       string
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(httptelemetry.Middleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/users/register", cfg.UserHandler.Register)
		r.Post("/users/login", cfg.UserHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))
			r.Post("/keys/generate", cfg.APIKeyHandler.GenerateKey)
			r.Get("/keys", cfg.APIKeyHandler.ListKeys)
			r.Put("/keys/{id}/status", cfg.APIKeyHandler.ToggleStatus)
			r.Delete("/keys/{id}", cfg.APIKeyHandler.RevokeKey)

			r.Group(func(r chi.Router) {
				r.Use(auth.AdminMiddleware)
				r.Get("/admin/audit-logs", cfg.AuditHandler.ListLogs)

				r.HandleFunc("/admin/debug/pprof/*", pprof.Index)
				r.HandleFunc("/admin/debug/pprof/cmdline", pprof.Cmdline)
				r.HandleFunc("/admin/debug/pprof/profile", pprof.Profile)
				r.HandleFunc("/admin/debug/pprof/symbol", pprof.Symbol)
				r.HandleFunc("/admin/debug/pprof/trace", pprof.Trace)
			})
		})

		r.Post("/analyze/resume", cfg.AnalysisHandler.ProcessResume)
		r.Get("/analyze/history", cfg.AnalysisHandler.ListHistory)
		r.Get("/analyze/{id}/result", cfg.AnalysisHandler.GetResult)
	})

	return r
}
