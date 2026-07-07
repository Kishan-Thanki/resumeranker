package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/auth"
)

func TestMiddleware(t *testing.T) {
	t.Parallel()

	secret := "middlewaresecret"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID == nil {
			t.Fatal("expected user_id in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := auth.Middleware(secret)(nextHandler)

	t.Run("valid token", func(t *testing.T) {
		token, _ := auth.GenerateToken(1, "user", secret, time.Hour)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "jwt", Value: token})
		rr := httptest.NewRecorder()

		mw.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("missing header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// Missing cookie!
		rr := httptest.NewRecorder()

		mw.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "jwt", Value: "invalidtoken"})
		rr := httptest.NewRecorder()

		mw.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})
}

func TestAdminMiddleware(t *testing.T) {
	t.Parallel()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := auth.AdminMiddleware(nextHandler)

	t.Run("valid admin role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		dummyMW := auth.Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mw.ServeHTTP(w, r)
		}))

		token, _ := auth.GenerateToken(1, "admin", "secret", time.Hour)
		req.AddCookie(&http.Cookie{Name: "jwt", Value: token})
		rr := httptest.NewRecorder()

		dummyMW.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	})
	
	t.Run("valid owner role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		dummyMW := auth.Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mw.ServeHTTP(w, r)
		}))

		token, _ := auth.GenerateToken(1, "owner", "secret", time.Hour)
		req.AddCookie(&http.Cookie{Name: "jwt", Value: token})
		rr := httptest.NewRecorder()

		dummyMW.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("non admin role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		dummyMW := auth.Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mw.ServeHTTP(w, r)
		}))

		token, _ := auth.GenerateToken(1, "user", "secret", time.Hour)
		req.AddCookie(&http.Cookie{Name: "jwt", Value: token})
		rr := httptest.NewRecorder()

		dummyMW.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rr.Code)
		}
	})
}
