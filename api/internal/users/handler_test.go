package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kishan-thanki/resumeranker/api/internal/ctxkey"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

func TestUserHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		registerErr    error
		returnNilUser  bool
		expectedStatus int
	}{
		{
			name:           "success",
			body:           `{"email":"test@example.com","password":"password123","agreed_to_terms":true}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json",
			body:           `{`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing email",
			body:           `{"email":"","password":"password123","agreed_to_terms":true}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing password",
			body:           `{"email":"test@example.com","password":"","agreed_to_terms":true}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "duplicate email",
			body:           `{"email":"test@example.com","password":"password123","agreed_to_terms":true}`,
			registerErr:    users.ErrUserAlreadyExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "must agree to terms",
			body:           `{"email":"test@example.com","password":"password123","agreed_to_terms":false}`,
			registerErr:    users.ErrMustAgreeToTerms,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			body:           `{"email":"test@example.com","password":"password123","agreed_to_terms":true}`,
			registerErr:    errors.New("database failure"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "nil user",
			body:           `{"email":"test@example.com","password":"password123","agreed_to_terms":true}`,
			returnNilUser:  true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &MockUserService{
				RegisterFunc: func(
					_ context.Context,
					_ string,
					_ string,
					_ users.Role,
					_ bool,
				) (*users.User, error) {
					if tt.registerErr != nil {
						return nil, tt.registerErr
					}
					if tt.returnNilUser {
						return nil, nil
					}
					return &users.User{ID: 1}, nil
				},
			}

			handler := users.NewUserHandler(
				svc,
				&MockAuthManager{},
				50,
			)

			req := httptest.NewRequest(
				http.MethodPost,
				"/register",
				bytes.NewBufferString(tt.body),
			)
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		authErr        error
		issueCookieErr error
		expectedStatus int
	}{
		{
			name:           "success",
			body:           `{"email":"test@example.com","password":"password123"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid json",
			body:           `{`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing email",
			body:           `{"email":"","password":"password123"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid credentials",
			body:           `{"email":"test@example.com","password":"wrong"}`,
			authErr:        users.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "suspended",
			body:           `{"email":"test@example.com","password":"password123"}`,
			authErr:        users.ErrAccountSuspended,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "session failure",
			body:           `{"email":"test@example.com","password":"password123"}`,
			issueCookieErr: errors.New("cookie failure"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &MockUserService{
				AuthenticateFunc: func(
					_ context.Context,
					_ string,
					_ string,
				) (*users.User, error) {
					if tt.authErr != nil {
						return nil, tt.authErr
					}
					return &users.User{
						ID:   1,
						Role: users.RoleUser,
					}, nil
				},
			}

			authManager := &MockAuthManager{
				IssueSessionCookieFunc: func(
					_ http.ResponseWriter,
					_ uint64,
					_ string,
				) error {
					return tt.issueCookieErr
				},
			}

			handler := users.NewUserHandler(svc, authManager, 50)

			req := httptest.NewRequest(
				http.MethodPost,
				"/login",
				bytes.NewBufferString(tt.body),
			)
			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestUserHandler_ForgotPassword(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		svc := &MockUserService{}
		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodPost,
			"/password/forgot",
			bytes.NewBufferString(`{"email":"test@example.com"}`),
		)
		rr := httptest.NewRecorder()

		handler.ForgotPassword(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("missing email", func(t *testing.T) {
		svc := &MockUserService{}
		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodPost,
			"/password/forgot",
			bytes.NewBufferString(`{"email":""}`),
		)
		rr := httptest.NewRecorder()

		handler.ForgotPassword(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})
}

func TestUserHandler_ResetPassword(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		svc := &MockUserService{}
		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodPost,
			"/password/reset",
			bytes.NewBufferString(`{"token":"token","new_password":"new-password"}`),
		)
		rr := httptest.NewRecorder()

		handler.ResetPassword(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		svc := &MockUserService{}
		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodPost,
			"/password/reset",
			bytes.NewBufferString(`{"token":"","new_password":"new-password"}`),
		)
		rr := httptest.NewRecorder()

		handler.ResetPassword(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := &MockUserService{
			ResetPasswordFunc: func(
				_ context.Context,
				_,
				_ string,
			) error {
				return errors.New("invalid token")
			},
		}
		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodPost,
			"/password/reset",
			bytes.NewBufferString(`{"token":"token","new_password":"new-password"}`),
		)
		rr := httptest.NewRecorder()

		handler.ResetPassword(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})
}

func TestUserHandler_VerifyEmail(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		svc := &MockUserService{}
		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodPost,
			"/verify",
			bytes.NewBufferString(`{"token":"verification-token"}`),
		)
		rr := httptest.NewRecorder()

		handler.VerifyEmail(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		svc := &MockUserService{}
		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodPost,
			"/verify",
			bytes.NewBufferString(`{"token":""}`),
		)
		rr := httptest.NewRecorder()

		handler.VerifyEmail(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})
}

func TestUserHandler_Logout(t *testing.T) {
	t.Parallel()

	var cleared bool

	authManager := &MockAuthManager{
		ClearSessionCookieFunc: func(_ http.ResponseWriter) {
			cleared = true
		},
	}

	handler := users.NewUserHandler(&MockUserService{}, authManager, 50)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !cleared {
		t.Fatal("expected session cookie to be cleared")
	}
}

func TestUserHandler_GetMe(t *testing.T) {
	t.Parallel()

	t.Run("missing user context", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		rr := httptest.NewRecorder()

		handler.GetMe(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc := &MockUserService{
			GetMeFunc: func(
				_ context.Context,
				userID uint64,
			) (*users.User, error) {
				return &users.User{
					ID:    userID,
					Email: "test@example.com",
				}, nil
			},
		}

		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		req = req.WithContext(
			context.WithValue(req.Context(), ctxkey.UserID, uint64(42)),
		)
		rr := httptest.NewRecorder()

		handler.GetMe(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})
}

func TestUserHandler_ChangePassword(t *testing.T) {
	t.Parallel()

	t.Run("missing context", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(http.MethodPut, "/password", nil)
		rr := httptest.NewRecorder()

		handler.ChangePassword(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		svc := &MockUserService{
			ChangePasswordFunc: func(
				_ context.Context,
				_ uint64,
				_,
				_ string,
			) error {
				return users.ErrIncorrectPassword
			},
		}

		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodPut,
			"/password",
			bytes.NewBufferString(`{"old_password":"old","new_password":"new"}`),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), ctxkey.UserID, uint64(1)),
		)
		rr := httptest.NewRecorder()

		handler.ChangePassword(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/password",
			bytes.NewBufferString(`{"old_password":"old","new_password":"new"}`),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), ctxkey.UserID, uint64(1)),
		)
		rr := httptest.NewRecorder()

		handler.ChangePassword(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})
}

func TestUserHandler_ToggleStatus(t *testing.T) {
	t.Parallel()

	t.Run("invalid id", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(http.MethodPut, "/users/abc/status", nil)
		rr := httptest.NewRecorder()

		handler.ToggleStatus(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/1/status",
			bytes.NewBufferString(`{"status":"invalid"}`),
		)
		rr := httptest.NewRecorder()

		req = withChiURLParam(req, "id", "1")

		handler.ToggleStatus(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/1/status",
			bytes.NewBufferString(`{"status":"suspended"}`),
		)
		req = withChiURLParam(req, "id", "1")

		rr := httptest.NewRecorder()

		handler.ToggleStatus(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})
}

func TestUserHandler_DeleteAccount(t *testing.T) {
	t.Parallel()

	handler := users.NewUserHandler(
		&MockUserService{},
		&MockAuthManager{},
		50,
	)

	req := httptest.NewRequest(http.MethodDelete, "/users/me", nil)
	req = req.WithContext(
		context.WithValue(req.Context(), ctxkey.UserID, uint64(42)),
	)
	rr := httptest.NewRecorder()

	handler.DeleteAccount(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	t.Parallel()

	t.Run("invalid limit", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(http.MethodGet, "/admin/users?limit=abc", nil)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("invalid offset", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(http.MethodGet, "/admin/users?offset=abc", nil)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("invalid limit range", func(t *testing.T) {
		handler := users.NewUserHandler(
			&MockUserService{},
			&MockAuthManager{},
			50,
		)

		req := httptest.NewRequest(http.MethodGet, "/admin/users?limit=101", nil)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc := &MockUserService{
			ListUsersFunc: func(
				_ context.Context,
				limit,
				offset int32,
			) ([]*users.User, int64, error) {
				if limit != 25 || offset != 5 {
					t.Fatalf("unexpected pagination: limit=%d offset=%d", limit, offset)
				}

				return []*users.User{
					{
						ID:    1,
						Email: "a@example.com",
					},
					{
						ID:    2,
						Email: "b@example.com",
					},
				}, 10, nil
			},
		}

		handler := users.NewUserHandler(svc, &MockAuthManager{}, 50)

		req := httptest.NewRequest(
			http.MethodGet,
			"/admin/users?limit=25&offset=5",
			nil,
		)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var response map[string]json.RawMessage
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if _, ok := response["users"]; !ok {
			t.Fatal("expected users field")
		}
		if _, ok := response["total"]; !ok {
			t.Fatal("expected total field")
		}
	})
}

func withChiURLParam(
	r *http.Request,
	key,
	value string,
) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(
		r.Context(),
		chi.RouteCtxKey,
		rctx,
	))
}
