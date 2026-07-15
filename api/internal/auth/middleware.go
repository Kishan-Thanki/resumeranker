package auth

import (
	"context"
	"net/http"

	"github.com/kishan-thanki/resumeranker/api/internal/ctxkey"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("jwt")
			if err != nil {
				http.Error(w, "missing or invalid authorization cookie", http.StatusUnauthorized)
				return
			}

			claims, err := ValidateToken(cookie.Value, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxkey.UserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxkey.UserRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ctxkey.UserRole).(string)
		if !ok || (role != string(users.RoleAdmin) && role != string(users.RoleOwner)) {
			http.Error(w, "forbidden: requires admin privileges", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func WebAuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("jwt")
			if err != nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			claims, err := ValidateToken(cookie.Value, secret)
			if err != nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), ctxkey.UserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxkey.UserRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
