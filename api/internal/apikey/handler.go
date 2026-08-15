package apikey

import (
	"context"
	"encoding/json"
	"errors"
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
	GetAPIKeyStats(ctx context.Context, userID, keyID uint64) (*APIKeyUsageResponse, error)
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
	userID, ok := userIDFromRequest(r)
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
		req.Quota = 1_000_000
	}

	plainTextKey, apiKey, err := h.apiKeyService.GenerateKey(
		r.Context(),
		userID,
		req.Name,
		req.Quota,
	)
	if err != nil {
		writeAPIKeyServiceError(w, err)
		return
	}

	if apiKey == nil {
		http.Error(
			w,
			"an internal server error occurred",
			http.StatusInternalServerError,
		)
		return
	}

	writeJSON(w, http.StatusCreated, GenerateKeyResponse{
		Message: "Key generated successfully. This is the ONLY time the key will be shown.",
		Key:     plainTextKey,
		KeyID:   apiKey.ID,
	})
}

func (h *APIKeyHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	keys, err := h.apiKeyService.ListKeys(r.Context(), userID)
	if err != nil {
		writeAPIKeyServiceError(w, err)
		return
	}

	dtos := make([]*APIKeyDTO, 0, len(keys))
	for _, key := range keys {
		if key == nil {
			continue
		}

		dtos = append(dtos, &APIKeyDTO{
			ID:                key.ID,
			Name:              key.Name,
			KeyPrefix:         key.KeyPrefix,
			Status:            key.Status,
			RequestsPerMinute: key.RequestsPerMinute,
			RequestsPerDay:    key.RequestsPerDay,
			TokenQuota:        key.TokenQuota,
			TokensUsed:        key.TokensUsed,
			ExpiresAt:         key.ExpiresAt,
			LastUsedAt:        key.LastUsedAt,
			CreatedAt:         key.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, dtos)
}

func (h *APIKeyHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	keyID, ok := parseKeyID(w, r)
	if !ok {
		return
	}

	var req UpdateKeyStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !req.Status.IsValid() {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	if err := h.apiKeyService.ToggleStatus(
		r.Context(),
		userID,
		keyID,
		req.Status,
	); err != nil {
		writeAPIKeyServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "status updated successfully",
	})
}

func (h *APIKeyHandler) RevokeKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	keyID, ok := parseKeyID(w, r)
	if !ok {
		return
	}

	if err := h.apiKeyService.RevokeKey(r.Context(), userID, keyID); err != nil {
		writeAPIKeyServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "key revoked successfully",
	})
}

func (h *APIKeyHandler) GetAPIKeyStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	keyID, ok := parseKeyID(w, r)
	if !ok {
		return
	}

	stats, err := h.apiKeyService.GetAPIKeyStats(
		r.Context(),
		userID,
		keyID,
	)
	if err != nil {
		writeAPIKeyServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func userIDFromRequest(r *http.Request) (uint64, bool) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	return userID, ok
}

func parseKeyID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	idStr := chi.URLParam(r, "id")

	keyID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return 0, false
	}

	return keyID, true
}

func writeAPIKeyServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrAPIKeyAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ErrUnauthorizedAPIKey):
		http.Error(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, ErrInvalidAPIKey):
		http.Error(w, "invalid api key", http.StatusBadRequest)
	case errors.Is(err, ErrTokenQuotaExceeded):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "an internal server error occurred", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "an internal server error occurred", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, _ = w.Write(body)
}
