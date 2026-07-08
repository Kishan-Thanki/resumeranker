package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

type mockAuthManager struct{}

func (m *mockAuthManager) IssueSessionCookie(w http.ResponseWriter, userID uint64, role string) error {
	return nil
}

func (m *mockAuthManager) ClearSessionCookie(w http.ResponseWriter) {}

func TestUserHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		payload        interface{}
		mockRegister   func(ctx context.Context, email, password string, role users.Role, agreedToTerms bool) (*users.User, error)
		expectedStatus int
	}{
		{
			name: "success",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockRegister: func(ctx context.Context, email, password string, role users.Role, agreedToTerms bool) (*users.User, error) {
				return &users.User{ID: 1}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid payload",
			payload:        "not-json",
			mockRegister:   nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockRegister: func(ctx context.Context, email, password string, role users.Role, agreedToTerms bool) (*users.User, error) {
				return nil, errors.New("registration failed")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &MockUserService{
				RegisterFunc: tt.mockRegister,
			}
			h := users.NewUserHandler(svc, &mockAuthManager{}, 50)

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.payload)

			req := httptest.NewRequest(http.MethodPost, "/register", &buf)
			rr := httptest.NewRecorder()

			h.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		payload          interface{}
		mockAuthenticate func(ctx context.Context, email, password string) (*users.User, error)
		expectedStatus   int
	}{
		{
			name: "success",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockAuthenticate: func(ctx context.Context, email, password string) (*users.User, error) {
				return &users.User{ID: 1, Role: users.RoleUser}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			mockAuthenticate: func(ctx context.Context, email, password string) (*users.User, error) {
				return nil, users.ErrInvalidCredentials
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &MockUserService{
				AuthenticateFunc: tt.mockAuthenticate,
			}
			h := users.NewUserHandler(svc, &mockAuthManager{}, 50)

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.payload)

			req := httptest.NewRequest(http.MethodPost, "/login", &buf)
			rr := httptest.NewRecorder()

			h.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
