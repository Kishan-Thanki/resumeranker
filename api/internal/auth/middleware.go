package auth

import (
	"context"
	"net/http"
)

func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("jwt")
			if err != nil {
				http.Error(w, "missing or invalid authorization cookie", http.StatusUnauthorized)
				return
			}
			tokenStr := cookie.Value

			userID, role, err := ValidateToken(tokenStr, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "user_role", role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("user_role").(string)
		if !ok || (role != "admin" && role != "owner") {
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
			tokenStr := cookie.Value

			userID, role, err := ValidateToken(tokenStr, secret)
			if err != nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "user_role", role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
