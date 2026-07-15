package users

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kishan-thanki/resumeranker/api/internal/ctxkey"
)

type userService interface {
	Register(ctx context.Context, email, password string, role Role, agreedToTerms bool) (*User, error)
	VerifyEmail(ctx context.Context, token string) error
	Authenticate(ctx context.Context, email, password string) (*User, error)
	GetMe(ctx context.Context, userID uint64) (*User, error)
	ListUsers(ctx context.Context, limit, offset int32) ([]*User, int64, error)
	ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ToggleStatus(ctx context.Context, userID uint64, status AccountStatus) error
	DeleteAccount(ctx context.Context, userID uint64) error
}

type authManager interface {
	IssueSessionCookie(w http.ResponseWriter, userID uint64, role string) error
	ClearSessionCookie(w http.ResponseWriter)
}

type UserHandler struct {
	userService  userService
	authManager  authManager
	defaultLimit int
}

func NewUserHandler(userService userService, authManager authManager, defaultLimit int) *UserHandler {
	return &UserHandler{
		userService:  userService,
		authManager:  authManager,
		defaultLimit: defaultLimit,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	user, err := h.userService.Register(r.Context(), req.Email, req.Password, RoleUser, req.AgreedToTerms)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]uint64{"user_id": user.ID})
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token is required"})
		return
	}

	err := h.userService.VerifyEmail(r.Context(), req.Token)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	user, err := h.userService.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrAccountSuspended) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Your account has been suspended. Please contact support."})
			return
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if err := h.authManager.IssueSessionCookie(w, user.ID, string(user.Role)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue session"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "logged in successfully",
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authManager.ClearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.userService.GetMe(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "old and new password are required"})
		return
	}

	err := h.userService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}

func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	_ = h.userService.ForgotPassword(r.Context(), req.Email)

	writeJSON(w, http.StatusOK, map[string]string{"message": "if an account exists, a password reset link has been sent"})
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token and new password are required"})
		return
	}

	err := h.userService.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

func (h *UserHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var req ToggleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err = h.userService.ToggleStatus(r.Context(), userID, req.Status)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "account status updated successfully"})
}

func (h *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	err := h.userService.DeleteAccount(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "account deleted successfully"})
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := h.defaultLimit
	if l, err := strconv.Atoi(limitStr); err == nil {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil {
		offset = o
	}

	users, count, err := h.userService.ListUsers(r.Context(), int32(limit), int32(offset))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(u)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": resp,
		"total": count,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error("failed to encode json response", "err", err)
		}
	}
}
