package users

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
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

func (h *AgreementHandler) RegisterRoutes(
	r chi.Router,
	authMiddleware func(http.Handler) http.Handler,
	requireRole func(string) func(http.Handler) http.Handler,
) {
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
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	agreement, err := h.agreementService.PublishAgreement(
		r.Context(),
		req.Type,
		req.Version,
		req.Content,
		true,
		h.userRepo,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAgreementType),
			errors.Is(err, ErrInvalidAgreementVersion),
			errors.Is(err, ErrInvalidAgreementContent):
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			slog.Error("failed to publish agreement", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to publish agreement",
			})
		}
		return
	}

	if agreement == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to publish agreement",
		})
		return
	}

	writeJSON(w, http.StatusCreated, toAgreementResponse(agreement))
}

func (h *AgreementHandler) GetLatestAgreements(w http.ResponseWriter, r *http.Request) {
	agreements, err := h.agreementService.GetLatestAgreements(r.Context())
	if err != nil {
		slog.Error("failed to get latest agreements", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get agreements",
		})
		return
	}

	responses := make([]AgreementResponse, 0, len(agreements))
	for _, agreement := range agreements {
		if agreement == nil {
			continue
		}
		responses = append(responses, toAgreementResponse(agreement))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agreements": responses,
	})
}

func (h *AgreementHandler) GetPendingAgreements(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	agreements, err := h.agreementService.GetPendingAgreements(
		r.Context(),
		userID,
	)
	if err != nil {
		slog.Error(
			"failed to get pending agreements",
			"user_id", userID,
			"error", err,
		)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get pending agreements",
		})
		return
	}

	responses := make([]AgreementResponse, 0, len(agreements))
	for _, agreement := range agreements {
		if agreement == nil {
			continue
		}
		responses = append(responses, toAgreementResponse(agreement))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agreements": responses,
	})
}

func (h *AgreementHandler) AcceptAgreements(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserID).(uint64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	var req AcceptAgreementsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if err := h.agreementService.AcceptAgreements(
		r.Context(),
		userID,
		req.AgreementIDs,
	); err != nil {
		if errors.Is(err, ErrNoAgreementsProvided) ||
			errors.Is(err, ErrInvalidAgreementID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		slog.Error(
			"failed to accept agreements",
			"user_id", userID,
			"error", err,
		)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to accept agreements",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "agreements accepted successfully",
	})
}
