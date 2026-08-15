package apikey_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/ctxkey"
)

func requestWithUser(
	method string,
	target string,
	body string,
	userID uint64,
) *http.Request {
	req := httptest.NewRequest(
		method,
		target,
		bytes.NewBufferString(body),
	)

	return req.WithContext(
		context.WithValue(req.Context(), ctxkey.UserID, userID),
	)
}

func requestWithRouteParam(
	method string,
	target string,
	body string,
	userID uint64,
	keyID string,
) *http.Request {
	req := requestWithUser(method, target, body, userID)

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", keyID)

	return req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, routeContext),
	)
}

func TestAPIKeyHandler_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("success forwards quota and returns JSON", func(t *testing.T) {
		t.Parallel()

		var gotUserID uint64
		var gotName string
		var gotQuota uint64

		svc := &MockAPIKeyService{
			GenerateKeyFunc: func(
				_ context.Context,
				userID uint64,
				name string,
				quota uint64,
			) (string, *apikey.APIKey, error) {
				gotUserID = userID
				gotName = name
				gotQuota = quota

				return "rr_plain_key", &apikey.APIKey{
					ID:         1,
					Name:       name,
					TokenQuota: quota,
				}, nil
			},
		}

		h := apikey.NewAPIKeyHandler(svc)

		req := requestWithUser(
			http.MethodPost,
			"/keys",
			`{"name":"test-key","quota":1000}`,
			42,
		)
		rr := httptest.NewRecorder()

		h.GenerateKey(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rr.Code)
		}
		if gotUserID != 42 {
			t.Fatalf("expected user ID 42, got %d", gotUserID)
		}
		if gotName != "test-key" {
			t.Fatalf("expected name test-key, got %q", gotName)
		}
		if gotQuota != 1000 {
			t.Fatalf("expected quota 1000, got %d", gotQuota)
		}
		if got := rr.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected application/json, got %q", got)
		}

		var resp apikey.GenerateKeyResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}
		if resp.Key != "rr_plain_key" {
			t.Fatalf("expected key rr_plain_key, got %q", resp.Key)
		}
		if resp.KeyID != 1 {
			t.Fatalf("expected key ID 1, got %d", resp.KeyID)
		}
	})

	t.Run("zero quota uses default", func(t *testing.T) {
		t.Parallel()

		var gotQuota uint64

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			GenerateKeyFunc: func(
				_ context.Context,
				_ uint64,
				_ string,
				quota uint64,
			) (string, *apikey.APIKey, error) {
				gotQuota = quota
				return "rr_key", &apikey.APIKey{ID: 1}, nil
			},
		})

		req := requestWithUser(
			http.MethodPost,
			"/keys",
			`{"name":"test-key","quota":0}`,
			1,
		)
		rr := httptest.NewRecorder()

		h.GenerateKey(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rr.Code)
		}
		if gotQuota != 1_000_000 {
			t.Fatalf("expected default quota 1000000, got %d", gotQuota)
		}
	})

	t.Run("missing user context", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{})

		req := httptest.NewRequest(
			http.MethodPost,
			"/keys",
			strings.NewReader(`{"name":"test-key","quota":1000}`),
		)
		rr := httptest.NewRecorder()

		h.GenerateKey(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{})

		req := requestWithUser(
			http.MethodPost,
			"/keys",
			`{"name":`,
			1,
		)
		rr := httptest.NewRecorder()

		h.GenerateKey(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("already has key returns conflict", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			GenerateKeyFunc: func(
				_ context.Context,
				_ uint64,
				_ string,
				_ uint64,
			) (string, *apikey.APIKey, error) {
				return "", nil, apikey.ErrAPIKeyAlreadyExists
			},
		})

		req := requestWithUser(
			http.MethodPost,
			"/keys",
			`{"name":"test-key","quota":1000}`,
			1,
		)
		rr := httptest.NewRecorder()

		h.GenerateKey(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", rr.Code)
		}
	})
}

func TestAPIKeyHandler_ListKeys(t *testing.T) {
	t.Parallel()

	t.Run("success hides secret fields", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			ListKeysFunc: func(
				_ context.Context,
				userID uint64,
			) ([]*apikey.APIKey, error) {
				if userID != 42 {
					t.Fatalf("expected user ID 42, got %d", userID)
				}

				return []*apikey.APIKey{
					{
						ID:                1,
						UserID:            42,
						Name:              "test",
						KeyPrefix:         "rr_abcd_",
						KeySelector:       "SELECTOR",
						KeyHash:           "SECRET_HASH",
						Status:            apikey.APIKeyStatusActive,
						RequestsPerMinute: 1,
						RequestsPerDay:    6,
						TokenQuota:        1000,
						TokensUsed:        25,
					},
				}, nil
			},
		})

		req := requestWithUser(http.MethodGet, "/keys", "", 42)
		rr := httptest.NewRecorder()

		h.ListKeys(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
		if got := rr.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected application/json, got %q", got)
		}

		body := rr.Body.String()
		if strings.Contains(body, "SECRET_HASH") {
			t.Fatal("response exposed KeyHash")
		}
		if strings.Contains(body, "SELECTOR") {
			t.Fatal("response exposed KeySelector")
		}

		var got []*apikey.APIKeyDTO
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected one API key, got %d", len(got))
		}
		if got[0].ID != 1 || got[0].TokenQuota != 1000 {
			t.Fatalf("unexpected DTO: %+v", got[0])
		}
	})

	t.Run("missing user context", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{})

		req := httptest.NewRequest(http.MethodGet, "/keys", nil)
		rr := httptest.NewRecorder()

		h.ListKeys(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			ListKeysFunc: func(
				_ context.Context,
				_ uint64,
			) ([]*apikey.APIKey, error) {
				return nil, errors.New("database failure")
			},
		})

		req := requestWithUser(http.MethodGet, "/keys", "", 1)
		rr := httptest.NewRecorder()

		h.ListKeys(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rr.Code)
		}
	})
}

func TestAPIKeyHandler_ToggleStatus(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var gotUserID uint64
		var gotKeyID uint64
		var gotStatus apikey.APIKeyStatus

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			ToggleStatusFunc: func(
				_ context.Context,
				userID, keyID uint64,
				status apikey.APIKeyStatus,
			) error {
				gotUserID = userID
				gotKeyID = keyID
				gotStatus = status
				return nil
			},
		})

		req := requestWithRouteParam(
			http.MethodPatch,
			"/keys/7/status",
			`{"status":"suspended"}`,
			42,
			"7",
		)
		rr := httptest.NewRecorder()

		h.ToggleStatus(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
		if gotUserID != 42 || gotKeyID != 7 {
			t.Fatalf("unexpected IDs: user=%d key=%d", gotUserID, gotKeyID)
		}
		if gotStatus != apikey.APIKeyStatusSuspended {
			t.Fatalf("expected suspended, got %q", gotStatus)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{})

		req := requestWithRouteParam(
			http.MethodPatch,
			"/keys/not-a-number/status",
			`{"status":"active"}`,
			1,
			"not-a-number",
		)
		rr := httptest.NewRecorder()

		h.ToggleStatus(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{})

		req := requestWithRouteParam(
			http.MethodPatch,
			"/keys/7/status",
			`{"status":`,
			1,
			"7",
		)
		rr := httptest.NewRecorder()

		h.ToggleStatus(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{})

		req := requestWithRouteParam(
			http.MethodPatch,
			"/keys/7/status",
			`{"status":"invalid"}`,
			1,
			"7",
		)
		rr := httptest.NewRecorder()

		h.ToggleStatus(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("unauthorized returns forbidden", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			ToggleStatusFunc: func(
				_ context.Context,
				_, _ uint64,
				_ apikey.APIKeyStatus,
			) error {
				return apikey.ErrUnauthorizedAPIKey
			},
		})

		req := requestWithRouteParam(
			http.MethodPatch,
			"/keys/7/status",
			`{"status":"active"}`,
			1,
			"7",
		)
		rr := httptest.NewRecorder()

		h.ToggleStatus(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rr.Code)
		}
	})
}

func TestAPIKeyHandler_RevokeKey(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			RevokeKeyFunc: func(
				_ context.Context,
				userID, keyID uint64,
			) error {
				if userID != 42 || keyID != 8 {
					t.Fatalf("unexpected IDs: user=%d key=%d", userID, keyID)
				}
				return nil
			},
		})

		req := requestWithRouteParam(
			http.MethodDelete,
			"/keys/8",
			"",
			42,
			"8",
		)
		rr := httptest.NewRecorder()

		h.RevokeKey(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{})

		req := requestWithRouteParam(
			http.MethodDelete,
			"/keys/bad",
			"",
			1,
			"bad",
		)
		rr := httptest.NewRecorder()

		h.RevokeKey(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("unauthorized returns forbidden", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			RevokeKeyFunc: func(
				_ context.Context,
				_, _ uint64,
			) error {
				return apikey.ErrUnauthorizedAPIKey
			},
		})

		req := requestWithRouteParam(
			http.MethodDelete,
			"/keys/8",
			"",
			1,
			"8",
		)
		rr := httptest.NewRecorder()

		h.RevokeKey(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rr.Code)
		}
	})
}

func TestAPIKeyHandler_GetAPIKeyStats(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			GetAPIKeyStatsFunc: func(
				_ context.Context,
				userID, keyID uint64,
			) (*apikey.APIKeyUsageResponse, error) {
				if userID != 42 || keyID != 9 {
					t.Fatalf("unexpected IDs: user=%d key=%d", userID, keyID)
				}

				return &apikey.APIKeyUsageResponse{
					RPMUsed:    1,
					RPMLimit:   20,
					RPDUsed:    4,
					RPDLimit:   500,
					TokensUsed: 25,
					TokenQuota: 1000,
				}, nil
			},
		})

		req := requestWithRouteParam(
			http.MethodGet,
			"/keys/9/stats",
			"",
			42,
			"9",
		)
		rr := httptest.NewRecorder()

		h.GetAPIKeyStats(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
		if got := rr.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected application/json, got %q", got)
		}

		var stats apikey.APIKeyUsageResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}
		if stats.RPMUsed != 1 || stats.RPDUsed != 4 {
			t.Fatalf("unexpected stats: %+v", stats)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		t.Parallel()

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{})

		req := requestWithRouteParam(
			http.MethodGet,
			"/keys/bad/stats",
			"",
			1,
			"bad",
		)
		rr := httptest.NewRecorder()

		h.GetAPIKeyStats(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("rate limiter failure")

		h := apikey.NewAPIKeyHandler(&MockAPIKeyService{
			GetAPIKeyStatsFunc: func(
				_ context.Context,
				_, _ uint64,
			) (*apikey.APIKeyUsageResponse, error) {
				return nil, expectedErr
			},
		})

		req := requestWithRouteParam(
			http.MethodGet,
			"/keys/9/stats",
			"",
			1,
			"9",
		)
		rr := httptest.NewRecorder()

		h.GetAPIKeyStats(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rr.Code)
		}
	})
}
