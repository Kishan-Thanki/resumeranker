package auth

import (
	"net/http"
	"time"
)

type Manager struct {
	jwtSecret       string
	environment     string
	sessionDuration time.Duration
}

func NewManager(jwtSecret, environment string, sessionDurationHours int) *Manager {
	return &Manager{
		jwtSecret:       jwtSecret,
		environment:     environment,
		sessionDuration: time.Duration(sessionDurationHours) * time.Hour,
	}
}

func (m *Manager) IssueSessionCookie(w http.ResponseWriter, userID uint64, role string) error {
	token, err := GenerateToken(userID, role, m.jwtSecret, m.sessionDuration)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(m.sessionDuration),
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
