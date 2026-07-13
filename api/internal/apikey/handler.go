package apikey

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kishan-thanki/resumeranker/api/internal/ctxkey"
)

type apiKeyService interface {
	GenerateKey(ctx context.Context, userID uint64, name string, quota uint64) (string, *APIKey, error)
	ListKeys(ctx context.Context, userID uint64) ([]*APIKey, error)
	ToggleStatus(ctx context.Context, userID, keyID uint64, status APIKeyStatus) error
	RevokeKey(ctx context.Context, userID, keyID uint64) error
}

type APIKeyHandler struct {
	apiKeyService apiKeyService
}

func NewAPIKeyHandler(apiKeyService apiKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

func (h *APIKeyHandler) GenerateKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req GenerateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Quota == 0 {
		req.Quota = 1000000
	}

	plainTextKey, apiKey, err := h.apiKeyService.GenerateKey(r.Context(), userID, req.Name, req.Quota)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(GenerateKeyResponse{
		Message: "Key generated successfully. This is the ONLY time the key will be shown.",
		Key:     plainTextKey,
		KeyID:   apiKey.ID,
	})
}

func (h *APIKeyHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	keys, err := h.apiKeyService.ListKeys(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(keys)
}

func (h *APIKeyHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	keyID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return
	}

	var req UpdateKeyStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != APIKeyStatusActive && req.Status != APIKeyStatusInactive && req.Status != APIKeyStatusSuspended {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	err = h.apiKeyService.ToggleStatus(r.Context(), userID, keyID, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "status updated successfully"})
}

func (h *APIKeyHandler) RevokeKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	keyID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return
	}

	err = h.apiKeyService.RevokeKey(r.Context(), userID, keyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "key revoked successfully"})
}
