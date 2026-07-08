package users

import (
	"context"
	"encoding/json"
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
	ListUsers(ctx context.Context, limit, offset int32) ([]*User, error)
	AcceptTerms(ctx context.Context, userID uint64, version string) error
	HasAcceptedTerms(ctx context.Context, userID uint64, version string) (bool, error)
	GetPendingAgreements(ctx context.Context, userID uint64) ([]*Agreement, error)
	AcceptAgreements(ctx context.Context, userID uint64, agreementIDs []uint64) error
	ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ToggleStatus(ctx context.Context, userID uint64, status AccountStatus) error
	DeleteAccount(ctx context.Context, userID uint64) error
	GetLatestAgreements(ctx context.Context) ([]*Agreement, error)
	PublishAgreement(ctx context.Context, agType AgreementType, version, content string) (*Agreement, error)
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(r.Context(), req.Email, req.Password, RoleUser, req.AgreedToTerms)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]uint64{"user_id": user.ID})
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.userService.VerifyEmail(r.Context(), req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "email verified successfully"})
}

func (h *UserHandler) GetPendingAgreements(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	agreements, err := h.userService.GetPendingAgreements(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]AgreementResponse, len(agreements))
	for i, a := range agreements {
		responses[i] = AgreementResponse{
			ID:      a.ID,
			Type:    string(a.Type),
			Version: a.Version,
			Content: a.Content,
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

func (h *UserHandler) AcceptAgreements(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req AcceptAgreementsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.userService.AcceptAgreements(r.Context(), userID, req.AgreementIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "agreements accepted successfully"})
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

	if err := h.authManager.IssueSessionCookie(w, user.ID, string(user.Role)); err != nil {
		http.Error(w, "failed to issue session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "logged in successfully",
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authManager.ClearSessionCookie(w)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetMe(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.userService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "password changed successfully"})
}

func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_ = h.userService.ForgotPassword(r.Context(), req.Email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "if an account exists, a password reset link has been sent"})
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.userService.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "password reset successfully"})
}

func (h *UserHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req ToggleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.userService.ToggleStatus(r.Context(), userID, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "account status updated successfully"})
}

func (h *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.userService.DeleteAccount(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "account deleted successfully"})
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

	users, err := h.userService.ListUsers(r.Context(), int32(limit), int32(offset))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetLatestAgreements(w http.ResponseWriter, r *http.Request) {
	agreements, err := h.userService.GetLatestAgreements(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]AgreementResponse, len(agreements))
	for i, a := range agreements {
		response[i] = AgreementResponse{
			ID:      a.ID,
			Type:    string(a.Type),
			Version: a.Version,
			Content: a.Content,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) PublishAgreement(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    AgreementType `json:"type"`
		Version string        `json:"version"`
		Content string        `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	agreement, err := h.userService.PublishAgreement(r.Context(), req.Type, req.Version, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AgreementResponse{
		ID:      agreement.ID,
		Type:    string(agreement.Type),
		Version: agreement.Version,
		Content: agreement.Content,
	})
}
