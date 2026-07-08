package analysis

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type analysisService interface {
	ProcessResume(ctx context.Context, plainTextKey string, resumeFilename, jdFilename string, resumePDF, jdPDF []byte) (*AnalysisResult, error)
	ListHistory(ctx context.Context, plainTextKey string, limit, offset int) ([]*AnalysisRequest, error)
	GetResult(ctx context.Context, plainTextKey string, requestID string) (*AnalysisResult, error)
}

type AnalysisHandler struct {
	analysisService analysisService
	defaultLimit    int
}

func NewAnalysisHandler(analysisService analysisService, defaultLimit int) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
		defaultLimit:    defaultLimit,
	}
}

func (h *AnalysisHandler) ProcessResume(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
		return
	}
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	resumeFile, resumeHeader, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, "missing resume file", http.StatusBadRequest)
		return
	}
	if resumeHeader.Header.Get("Content-Type") != "application/pdf" {
		http.Error(w, "resume must be a PDF", http.StatusBadRequest)
		return
	}
	defer resumeFile.Close()

	jdFile, jdHeader, err := r.FormFile("job_description")
	if err != nil {
		http.Error(w, "missing job_description file", http.StatusBadRequest)
		return
	}
	if jdHeader.Header.Get("Content-Type") != "application/pdf" {
		http.Error(w, "job_description must be a PDF", http.StatusBadRequest)
		return
	}
	defer jdFile.Close()

	resumeBytes, err := io.ReadAll(resumeFile)
	if err != nil {
		http.Error(w, "failed to read resume file", http.StatusInternalServerError)
		return
	}

	jdBytes, err := io.ReadAll(jdFile)
	if err != nil {
		http.Error(w, "failed to read job_description file", http.StatusInternalServerError)
		return
	}

	result, err := h.analysisService.ProcessResume(r.Context(), apiKey, resumeHeader.Filename, jdHeader.Filename, resumeBytes, jdBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result.Result,
	})
}

func (h *AnalysisHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
		return
	}
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")

	limit := h.defaultLimit
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil {
			offset = parsedOffset
		}
	}

	history, err := h.analysisService.ListHistory(r.Context(), apiKey, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}

func (h *AnalysisHandler) GetResult(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
		return
	}
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")

	requestID := chi.URLParam(r, "id")

	result, err := h.analysisService.GetResult(r.Context(), apiKey, requestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
