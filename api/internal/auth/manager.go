package auth

import (
	"net/http"
	"time"
)

type Manager struct {
	jwtSecret   string
	environment string
}

func NewManager(jwtSecret, environment string) *Manager {
	return &Manager{
		jwtSecret:   jwtSecret,
		environment: environment,
	}
}

func (m *Manager) IssueSessionCookie(w http.ResponseWriter, userID uint64, role string) error {
	token, err := GenerateToken(userID, role, m.jwtSecret, 24*time.Hour)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   m.environment == "production",
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

func (m *Manager) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   m.environment == "production",
		SameSite: http.SameSiteStrictMode,
	})
}
