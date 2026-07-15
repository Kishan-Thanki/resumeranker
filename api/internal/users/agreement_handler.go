package users

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kishan-thanki/resumeranker/api/internal/ctxkey"
)

type agreementService interface {
	PublishAgreement(ctx context.Context, agType AgreementType, version, content string, notifyUsers bool, userRepo UserRepository) (*Agreement, error)
	GetLatestAgreements(ctx context.Context) ([]*Agreement, error)
	GetPendingAgreements(ctx context.Context, userID uint64) ([]*Agreement, error)
	AcceptAgreements(ctx context.Context, userID uint64, agreementIDs []uint64) error
}

type AgreementHandler struct {
	agreementService agreementService
	userRepo         UserRepository
}

func NewAgreementHandler(agreementService agreementService, userRepo UserRepository) *AgreementHandler {
	return &AgreementHandler{
		agreementService: agreementService,
		userRepo:         userRepo,
	}
}

func (h *AgreementHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler, requireRole func(string) func(http.Handler) http.Handler) {
	r.Get("/agreements/latest", h.GetLatestAgreements)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/agreements/pending", h.GetPendingAgreements)
		r.Post("/agreements/accept", h.AcceptAgreements)

		r.Group(func(r chi.Router) {
			r.Use(requireRole(string(RoleAdmin)))
			r.Post("/agreements", h.PublishAgreement)
		})
	})
}

func (h *AgreementHandler) PublishAgreement(w http.ResponseWriter, r *http.Request) {
	var req PublishAgreementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	agreement, err := h.agreementService.PublishAgreement(r.Context(), req.Type, req.Version, req.Content, true, h.userRepo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, toAgreementResponse(agreement))
}

func (h *AgreementHandler) GetLatestAgreements(w http.ResponseWriter, r *http.Request) {
	agreements, err := h.agreementService.GetLatestAgreements(r.Context())
	if err != nil {
		http.Error(w, "failed to get agreements", http.StatusInternalServerError)
		return
	}

	responses := make([]AgreementResponse, len(agreements))
	for i, a := range agreements {
		responses[i] = toAgreementResponse(a)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agreements": responses,
	})
}

func (h *AgreementHandler) GetPendingAgreements(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	agreements, err := h.agreementService.GetPendingAgreements(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get pending agreements", http.StatusInternalServerError)
		return
	}

	responses := make([]AgreementResponse, len(agreements))
	for i, a := range agreements {
		responses[i] = toAgreementResponse(a)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agreements": responses,
	})
}

func (h *AgreementHandler) AcceptAgreements(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		AgreementIDs []uint64 `json:"agreement_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.agreementService.AcceptAgreements(r.Context(), userID, req.AgreementIDs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
