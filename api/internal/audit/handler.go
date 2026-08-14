package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type auditService interface {
	ListLogs(ctx context.Context, limit, offset int) ([]*AuditEvent, error)
}

type AuditHandler struct {
	auditService auditService
	defaultLimit int
}

func NewAuditHandler(auditService auditService, defaultLimit int) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		defaultLimit: defaultLimit,
	}
}

func (h *AuditHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := h.defaultLimit
	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsedOffset
	}

	logs, err := h.auditService.ListLogs(r.Context(), limit, offset)
	if err != nil {
		http.Error(
			w,
			"an internal server error occurred",
			http.StatusInternalServerError,
		)
		return
	}

	body, err := json.Marshal(logs)
	if err != nil {
		http.Error(
			w,
			"an internal server error occurred",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		return
	}
}
