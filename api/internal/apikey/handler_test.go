package apikey_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
)

func TestAPIKeyHandler_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		svc := &MockAPIKeyService{
			GenerateKeyFunc: func(ctx context.Context, userID uint64, name string, quota uint64) (string, *apikey.APIKey, error) {
				return "rr_plain_key", &apikey.APIKey{ID: 1, Name: name, TokenQuota: quota}, nil
			},
		}
		h := apikey.NewAPIKeyHandler(svc)

		reqBody := `{"name":"test-key", "token_quota": 1000}`
		req := httptest.NewRequest(http.MethodPost, "/keys", bytes.NewBufferString(reqBody))

		ctx := context.WithValue(req.Context(), "user_id", uint64(1))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		h.GenerateKey(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", rr.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&resp)

		if resp["key"] != "rr_plain_key" {
			t.Errorf("expected plain_text_key rr_plain_key, got %v", resp["key"])
		}
	})

	t.Run("missing user context", func(t *testing.T) {
		svc := &MockAPIKeyService{}
		h := apikey.NewAPIKeyHandler(svc)

		reqBody := `{"name":"test-key", "token_quota": 1000}`
		req := httptest.NewRequest(http.MethodPost, "/keys", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		h.GenerateKey(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})
}
