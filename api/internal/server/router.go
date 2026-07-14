package server

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/gorilla/csrf"
	"github.com/kishan-thanki/logger/v2/httptelemetry"
	"github.com/kishan-thanki/resumeranker/api/docs"
	"github.com/kishan-thanki/resumeranker/api/internal/analysis"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/auth"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

type RouterConfig struct {
	Environment     string
	UserHandler     *users.UserHandler
	APIKeyHandler   *apikey.APIKeyHandler
	AnalysisHandler *analysis.AnalysisHandler
	AuditHandler    *audit.AuditHandler
	JWTSecret       string
	CSRFAuthKey     string
	AllowedOrigins  []string
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(httptelemetry.Middleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(180 * time.Second))

	corsOptions := cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	r.Use(cors.Handler(corsOptions))

	r.Group(func(r chi.Router) {
		r.Get("/docs/public", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write(docs.PublicHTML)
		})

		r.Get("/docs/public/yaml", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			w.Write(docs.PublicAPIYAML)
		})
	})

	if cfg.Environment == "development" {
		r.Group(func(r chi.Router) {
			r.Get("/docs/internal", func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.Write(docs.InternalHTML)
			})

			r.Get("/docs/internal/yaml", func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/yaml")
				w.Write(docs.InternalAPIYAML)
			})
		})
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.ClientIPFromXFF())

		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitBy(5, 1*time.Minute, func(req *http.Request) (string, error) {
				return httprate.CanonicalizeIP(middleware.GetClientIP(req.Context())), nil
			}))
			r.Post("/users/register", cfg.UserHandler.Register)
			r.Post("/users/login", cfg.UserHandler.Login)
			r.Post("/users/password/forgot", cfg.UserHandler.ForgotPassword)
			r.Post("/users/password/reset", cfg.UserHandler.ResetPassword)
		})

		r.Post("/users/verify", cfg.UserHandler.VerifyEmail)
		r.Get("/agreements/latest", cfg.UserHandler.GetLatestAgreements)

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))

			csrfMiddleware := csrf.Protect(
				[]byte(cfg.CSRFAuthKey),
				csrf.Secure(cfg.Environment == "production"),
				csrf.Path("/"),
				csrf.TrustedOrigins([]string{"localhost", "localhost:8080", "localhost:80", "localhost:8443", "localhost:9080"}),
			)
			r.Use(csrfMiddleware)

			r.Get("/csrf-token", func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-CSRF-Token", csrf.Token(req))
				w.WriteHeader(http.StatusOK)
			})

			r.Get("/users/me", cfg.UserHandler.GetMe)
			r.Post("/users/logout", cfg.UserHandler.Logout)
			r.Put("/users/me/password", cfg.UserHandler.ChangePassword)
			r.Delete("/users/me", cfg.UserHandler.DeleteAccount)

			r.Get("/auth/agreements/pending", cfg.UserHandler.GetPendingAgreements)
			r.Post("/auth/agreements/accept", cfg.UserHandler.AcceptAgreements)

			r.Group(func(r chi.Router) {
				r.Use(httprate.LimitBy(10, 1*time.Minute, func(req *http.Request) (string, error) {
					return httprate.CanonicalizeIP(middleware.GetClientIP(req.Context())), nil
				}))
				r.Post("/keys/generate", cfg.APIKeyHandler.GenerateKey)
			})

			r.Get("/keys", cfg.APIKeyHandler.ListKeys)
			r.Get("/keys/{id}/stats", cfg.APIKeyHandler.GetAPIKeyStats)
			r.Put("/keys/{id}/status", cfg.APIKeyHandler.ToggleStatus)
			r.Delete("/keys/{id}", cfg.APIKeyHandler.RevokeKey)

			r.Group(func(r chi.Router) {
				r.Use(auth.AdminMiddleware)
				r.Get("/admin/users", cfg.UserHandler.ListUsers)
				r.Get("/admin/audit-logs", cfg.AuditHandler.ListLogs)
				r.Put("/admin/users/{id}/status", cfg.UserHandler.ToggleStatus)
				r.Post("/admin/agreements", cfg.UserHandler.PublishAgreement)

				if cfg.Environment == "development" {
					r.HandleFunc("/admin/debug/pprof/*", pprof.Index)
					r.HandleFunc("/admin/debug/pprof/cmdline", pprof.Cmdline)
					r.HandleFunc("/admin/debug/pprof/profile", pprof.Profile)
					r.HandleFunc("/admin/debug/pprof/symbol", pprof.Symbol)
					r.HandleFunc("/admin/debug/pprof/trace", pprof.Trace)
				}
			})
		})

		r.Post("/analyze/resume", cfg.AnalysisHandler.ProcessResume)
		r.Get("/analyze/history", cfg.AnalysisHandler.ListHistory)
		r.Get("/analyze/{id}/result", cfg.AnalysisHandler.GetResult)
	})

	return r
}
