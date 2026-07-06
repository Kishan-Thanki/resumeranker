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
}

func NewAuditHandler(auditService auditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

func (h *AuditHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	logs, err := h.auditService.ListLogs(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
