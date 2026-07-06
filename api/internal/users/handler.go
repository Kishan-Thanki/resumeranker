package users

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/auth"
)

type userService interface {
	Register(ctx context.Context, email, password string, role Role) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	AcceptTerms(ctx context.Context, userID uint64, version string) error
	HasAcceptedTerms(ctx context.Context, userID uint64, version string) (bool, error)
}

type UserHandler struct {
	userService userService
	jwtSecret   string
}

func NewUserHandler(userService userService, jwtSecret string) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(r.Context(), req.Email, req.Password, RoleUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]uint64{"user_id": user.ID})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, h.jwtSecret, 1*time.Hour)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "logged in successfully",
		"token":      token,
		"expires_in": 3600,
	})
}
