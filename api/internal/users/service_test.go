package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/password"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

func assertEqual(t *testing.T, got, want any, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

func TestUserService_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		email     string
		password  string
		role      users.Role
		mockRepo  func(ctx context.Context, user *users.User) (*users.User, error)
		wantError bool
	}{
		{
			name:     "success path",
			email:    "test@example.com",
			password: "strongpassword123",
			role:     users.RoleUser,
			mockRepo: func(ctx context.Context, user *users.User) (*users.User, error) {
				user.ID = 1
				return user, nil
			},
			wantError: false,
		},
		{
			name:     "repo error",
			email:    "test@example.com",
			password: "strongpassword123",
			role:     users.RoleUser,
			mockRepo: func(ctx context.Context, user *users.User) (*users.User, error) {
				return nil, errors.New("db error")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &MockRepository{
				CreateUserFunc: tt.mockRepo,
			}
			auditSvc := &MockAuditService{}
			emailSvc := &MockEmailService{}
			svc := users.NewUserService(repo, auditSvc, emailSvc)

			createdUser, err := svc.Register(context.Background(), tt.email, tt.password, tt.role, true)

			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if createdUser == nil {
				t.Fatal("expected user, got nil")
			}
			if createdUser.ID == 0 {
				t.Error("expected valid user ID")
			}
		})
	}
}

func TestUserService_Authenticate(t *testing.T) {
	t.Parallel()

	validHash, _ := password.HashIt("correctpassword")

	tests := []struct {
		name      string
		email     string
		password  string
		mockRepo  func(ctx context.Context, email string) (*users.User, error)
		wantError error
	}{
		{
			name:     "success path",
			email:    "test@example.com",
			password: "correctpassword",
			mockRepo: func(ctx context.Context, email string) (*users.User, error) {
				return &users.User{
					ID:           1,
					Email:        email,
					PasswordHash: validHash,
					Status:       users.AccountStatusActive,
					IsVerified:   true,
				}, nil
			},
			wantError: nil,
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "wrongpassword",
			mockRepo: func(ctx context.Context, email string) (*users.User, error) {
				return &users.User{
					ID:           1,
					Email:        email,
					PasswordHash: validHash,
					Status:       users.AccountStatusActive,
					IsVerified:   true,
				}, nil
			},
			wantError: users.ErrInvalidCredentials,
		},
		{
			name:     "user not found",
			email:    "notfound@example.com",
			password: "password",
			mockRepo: func(ctx context.Context, email string) (*users.User, error) {
				return nil, errors.New("not found")
			},
			wantError: users.ErrInvalidCredentials,
		},
		{
			name:     "account not active",
			email:    "test@example.com",
			password: "correctpassword",
			mockRepo: func(ctx context.Context, email string) (*users.User, error) {
				return &users.User{
					ID:           1,
					Email:        email,
					PasswordHash: validHash,
					Status:       users.AccountStatusSuspended,
					IsVerified:   true,
				}, nil
			},
			wantError: errors.New("account is not active"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &MockRepository{
				GetUserByEmailFunc: tt.mockRepo,
			}
			auditSvc := &MockAuditService{}
			emailSvc := &MockEmailService{}
			svc := users.NewUserService(repo, auditSvc, emailSvc)

			u, err := svc.Authenticate(context.Background(), tt.email, tt.password)

			if tt.wantError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantError)
				}
				assertEqual(t, err.Error(), tt.wantError.Error(), "error mismatch")
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u == nil {
				t.Fatal("expected user, got nil")
			}
		})
	}
}

func BenchmarkAuthenticate(b *testing.B) {
	b.ReportAllocs()

	validHash, _ := password.HashIt("benchmarkpassword")

	repo := &MockRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*users.User, error) {
			return &users.User{
				ID:           1,
				Email:        email,
				PasswordHash: validHash,
				Status:       users.AccountStatusActive,
				IsVerified:   true,
			}, nil
		},
	}
	auditSvc := &MockAuditService{}
	emailSvc := &MockEmailService{}
	svc := users.NewUserService(repo, auditSvc, emailSvc)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Authenticate(ctx, "test@example.com", "benchmarkpassword")
	}
}

func FuzzAuthenticate(f *testing.F) {
	validHash, _ := password.HashIt("fuzzpassword")

	f.Add("test@example.com", "fuzzpassword")
	f.Add("invalidemail", "wrongpassword")
	f.Add("", "")

	repo := &MockRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*users.User, error) {
			return &users.User{
				ID:           1,
				Email:        email,
				PasswordHash: validHash,
				Status:       users.AccountStatusActive,
				IsVerified:   true,
			}, nil
		},
	}
	auditSvc := &MockAuditService{}
	emailSvc := &MockEmailService{}
	svc := users.NewUserService(repo, auditSvc, emailSvc)

	f.Fuzz(func(t *testing.T, email, pass string) {
		_, _ = svc.Authenticate(context.Background(), email, pass)
	})
}
