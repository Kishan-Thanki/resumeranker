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

	ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	GetMe(ctx context.Context, userID uint64) (*User, error)
	ListUsers(ctx context.Context, limit, offset int32) ([]*User, int64, error)

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
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "email and password are required",
		})
		return
	}

	user, err := h.userService.Register(
		r.Context(),
		req.Email,
		req.Password,
		RoleUser,
		req.AgreedToTerms,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserAlreadyExists):
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, ErrMustAgreeToTerms):
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			slog.Error("failed to register user", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to register user",
			})
		}
		return
	}

	if user == nil {
		slog.Error("user registration returned nil user")
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to register user",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]uint64{
		"user_id": user.ID,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "email and password are required",
		})
		return
	}

	user, err := h.userService.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrAccountSuspended) {
			writeJSON(w, http.StatusForbidden, map[string]string{
				"error": "Your account has been suspended. Please contact support.",
			})
			return
		}

		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "invalid credentials",
		})
		return
	}

	if user == nil {
		slog.Error("authentication returned nil user")
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to authenticate",
		})
		return
	}

	if err := h.authManager.IssueSessionCookie(w, user.ID, string(user.Role)); err != nil {
		slog.Error("failed to issue session cookie", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to issue session",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "logged in successfully",
	})
}

func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "email is required",
		})
		return
	}

	_ = h.userService.ForgotPassword(r.Context(), req.Email)

	// Deliberately do not reveal whether the account exists.
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "if an account exists, a password reset link has been sent",
	})
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "token and new password are required",
		})
		return
	}

	if err := h.userService.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "password reset successfully",
	})
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.Token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "token is required",
		})
		return
	}

	if err := h.userService.VerifyEmail(r.Context(), req.Token); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "email verified successfully",
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authManager.ClearSessionCookie(w)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	user, err := h.userService.GetMe(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get current user", "user_id", userID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get user",
		})
		return
	}

	if user == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get user",
		})
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "old and new password are required",
		})
		return
	}

	err := h.userService.ChangePassword(
		r.Context(),
		userID,
		req.OldPassword,
		req.NewPassword,
	)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		slog.Error("failed to change password", "user_id", userID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to change password",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "password changed successfully",
	})
}

func (h *UserHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || userID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid user id",
		})
		return
	}

	var req ToggleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if !req.Status.IsValid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": ErrInvalidStatus.Error(),
		})
		return
	}

	if err := h.userService.ToggleStatus(r.Context(), userID, req.Status); err != nil {
		if errors.Is(err, ErrInvalidStatus) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		slog.Error("failed to update account status", "user_id", userID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update account status",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "account status updated successfully",
	})
}

func (h *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	if err := h.userService.DeleteAccount(r.Context(), userID); err != nil {
		slog.Error("failed to delete account", "user_id", userID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete account",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "account deleted successfully",
	})
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := h.defaultLimit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid limit",
			})
			return
		}
		limit = parsed
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid offset",
			})
			return
		}
		offset = parsed
	}

	if limit <= 0 || limit > 100 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "limit must be between 1 and 100",
		})
		return
	}

	if offset < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "offset must be zero or greater",
		})
		return
	}

	userList, count, err := h.userService.ListUsers(
		r.Context(),
		int32(limit),
		int32(offset),
	)
	if err != nil {
		slog.Error("failed to list users", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list users",
		})
		return
	}

	resp := make([]UserResponse, 0, len(userList))
	for _, user := range userList {
		if user == nil {
			continue
		}
		resp = append(resp, toUserResponse(user))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": resp,
		"total": count,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode json response", "err", err)
	}
}
