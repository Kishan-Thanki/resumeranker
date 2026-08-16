package users_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
	"github.com/kishan-thanki/resumeranker/api/internal/config"
	"github.com/kishan-thanki/resumeranker/api/internal/password"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

func testConfig() *config.Config {
	return &config.Config{
		Domain:                   "https://example.com",
		EmailContact:             "support@example.com",
		VerifyTokenDurationHours: 24,
		ResetTokenDurationHours:  1,
	}
}

func newUserService(
	repo *MockUserRepository,
	agreementRepo *MockAgreementRepository,
	auditSvc *MockAuditService,
	emailSvc *MockEmailService,
	apiKeySvc *MockAPIKeyService,
	wg *sync.WaitGroup,
) *users.UserService {
	return users.NewUserService(
		repo,
		agreementRepo,
		auditSvc,
		emailSvc,
		apiKeySvc,
		testConfig(),
		wg,
	)
}

func TestUserService_Register(t *testing.T) {
	t.Parallel()

	t.Run("requires agreement", func(t *testing.T) {
		t.Parallel()

		svc := newUserService(
			&MockUserRepository{},
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		_, err := svc.Register(
			context.Background(),
			"user@example.com",
			"password123",
			users.RoleUser,
			false,
		)

		if !errors.Is(err, users.ErrMustAgreeToTerms) {
			t.Fatalf("expected ErrMustAgreeToTerms, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		repo := &MockUserRepository{
			CreateUserFunc: func(
				_ context.Context,
				user *users.User,
			) (*users.User, error) {
				user.ID = 1
				return user, nil
			},
		}

		agreementRepo := &MockAgreementRepository{
			GetLatestAgreementsFunc: func(
				_ context.Context,
			) ([]*users.Agreement, error) {
				return []*users.Agreement{
					{
						ID:   10,
						Type: users.AgreementTypeTermsOfService,
					},
					{
						ID:   11,
						Type: users.AgreementTypePrivacyPolicy,
					},
				}, nil
			},
		}

		apiKeySvc := &MockAPIKeyService{
			GenerateKeyFunc: func(
				_ context.Context,
				userID uint64,
				name string,
				quota uint64,
			) (string, *apikey.APIKey, error) {
				return "mock", &apikey.APIKey{
					ID:     100,
					UserID: userID,
					Name:   name,
				}, nil
			},
		}

		svc := newUserService(
			repo,
			agreementRepo,
			&MockAuditService{},
			&MockEmailService{},
			apiKeySvc,
			&wg,
		)

		created, err := svc.Register(
			context.Background(),
			"user@example.com",
			"password123",
			users.RoleUser,
			true,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if created == nil {
			t.Fatal("expected created user")
		}

		if created.ID != 1 {
			t.Fatalf("expected user ID 1, got %d", created.ID)
		}

		if created.PasswordHash == "" {
			t.Fatal("expected password hash")
		}

		if created.VerificationToken == nil {
			t.Fatal("expected verification token")
		}

		wg.Wait()
	})

	t.Run("agreement lookup failure rolls back user", func(t *testing.T) {
		t.Parallel()

		var deleted bool

		repo := &MockUserRepository{
			CreateUserFunc: func(
				_ context.Context,
				user *users.User,
			) (*users.User, error) {
				user.ID = 1
				return user, nil
			},
			DeleteUserFunc: func(
				_ context.Context,
				id uint64,
			) error {
				if id == 1 {
					deleted = true
				}
				return nil
			},
		}

		agreementRepo := &MockAgreementRepository{
			GetLatestAgreementsFunc: func(
				_ context.Context,
			) ([]*users.Agreement, error) {
				return nil, errors.New("agreement lookup failed")
			},
		}

		svc := newUserService(
			repo,
			agreementRepo,
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		_, err := svc.Register(
			context.Background(),
			"user@example.com",
			"password123",
			users.RoleUser,
			true,
		)

		if err == nil {
			t.Fatal("expected error")
		}

		if !deleted {
			t.Fatal("expected created user to be deleted during rollback")
		}
	})

	t.Run("agreement acceptance failure rolls back user", func(t *testing.T) {
		t.Parallel()

		var deleted bool

		repo := &MockUserRepository{
			CreateUserFunc: func(
				_ context.Context,
				user *users.User,
			) (*users.User, error) {
				user.ID = 1
				return user, nil
			},
			DeleteUserFunc: func(
				_ context.Context,
				id uint64,
			) error {
				if id == 1 {
					deleted = true
				}
				return nil
			},
		}

		agreementRepo := &MockAgreementRepository{
			GetLatestAgreementsFunc: func(
				_ context.Context,
			) ([]*users.Agreement, error) {
				return []*users.Agreement{{ID: 10}}, nil
			},
			CreateUserAgreementFunc: func(
				_ context.Context,
				_ *users.UserAgreement,
			) (*users.UserAgreement, error) {
				return nil, errors.New("agreement acceptance failed")
			},
		}

		svc := newUserService(
			repo,
			agreementRepo,
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		_, err := svc.Register(
			context.Background(),
			"user@example.com",
			"password123",
			users.RoleUser,
			true,
		)

		if err == nil {
			t.Fatal("expected error")
		}

		if !deleted {
			t.Fatal("expected created user to be deleted during rollback")
		}
	})
}

func TestUserService_Authenticate(t *testing.T) {
	t.Parallel()

	validHash, err := password.HashIt("correctpassword")
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	tests := []struct {
		name     string
		user     *users.User
		repoErr  error
		password string
		wantErr  error
	}{
		{
			name: "success",
			user: &users.User{
				ID:           1,
				Email:        "test@example.com",
				PasswordHash: validHash,
				Status:       users.AccountStatusActive,
				IsVerified:   true,
			},
			password: "correctpassword",
		},
		{
			name: "invalid password",
			user: &users.User{
				ID:           1,
				PasswordHash: validHash,
				Status:       users.AccountStatusActive,
				IsVerified:   true,
			},
			password: "wrongpassword",
			wantErr:  users.ErrInvalidCredentials,
		},
		{
			name:     "user not found",
			repoErr:  errors.New("not found"),
			password: "correctpassword",
			wantErr:  users.ErrInvalidCredentials,
		},
		{
			name: "unverified",
			user: &users.User{
				ID:           1,
				PasswordHash: validHash,
				Status:       users.AccountStatusActive,
				IsVerified:   false,
			},
			password: "correctpassword",
			wantErr:  errors.New("email is not verified"),
		},
		{
			name: "suspended",
			user: &users.User{
				ID:           1,
				PasswordHash: validHash,
				Status:       users.AccountStatusSuspended,
				IsVerified:   true,
			},
			password: "correctpassword",
			wantErr:  users.ErrAccountSuspended,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &MockUserRepository{
				GetUserByEmailFunc: func(
					_ context.Context,
					_ string,
				) (*users.User, error) {
					return tt.user, tt.repoErr
				},
			}

			svc := newUserService(
				repo,
				&MockAgreementRepository{},
				&MockAuditService{},
				&MockEmailService{},
				&MockAPIKeyService{},
				&sync.WaitGroup{},
			)

			got, err := svc.Authenticate(
				context.Background(),
				"test@example.com",
				tt.password,
			)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected user")
			}
		})
	}
}

func TestUserService_VerifyEmail(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expires := time.Now().Add(time.Hour)
		var verifiedID uint64

		repo := &MockUserRepository{
			GetUserByVerificationTokenFunc: func(
				_ context.Context,
				token string,
			) (*users.User, error) {
				if token != "token" {
					t.Fatalf("unexpected token %q", token)
				}
				return &users.User{
					ID:                    1,
					VerificationExpiresAt: &expires,
				}, nil
			},
			VerifyUserEmailFunc: func(
				_ context.Context,
				id uint64,
			) (*users.User, error) {
				verifiedID = id
				return &users.User{ID: id, IsVerified: true}, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		if err := svc.VerifyEmail(context.Background(), "token"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if verifiedID != 1 {
			t.Fatalf("expected verified ID 1, got %d", verifiedID)
		}
	})

	t.Run("expired", func(t *testing.T) {
		t.Parallel()

		expires := time.Now().Add(-time.Hour)

		repo := &MockUserRepository{
			GetUserByVerificationTokenFunc: func(
				_ context.Context,
				_ string,
			) (*users.User, error) {
				return &users.User{
					ID:                    1,
					VerificationExpiresAt: &expires,
				}, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		if err := svc.VerifyEmail(context.Background(), "token"); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUserService_ForgotPassword(t *testing.T) {
	t.Parallel()

	t.Run("unknown email is intentionally silent", func(t *testing.T) {
		t.Parallel()

		repo := &MockUserRepository{
			GetUserByEmailFunc: func(
				_ context.Context,
				_ string,
			) (*users.User, error) {
				return nil, errors.New("not found")
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		if err := svc.ForgotPassword(context.Background(), "missing@example.com"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("updates reset token", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		var updated *users.User

		repo := &MockUserRepository{
			GetUserByEmailFunc: func(
				_ context.Context,
				_ string,
			) (*users.User, error) {
				return &users.User{
					ID:    1,
					Email: "user@example.com",
				}, nil
			},
			UpdateUserFunc: func(
				_ context.Context,
				user *users.User,
			) (*users.User, error) {
				updated = user
				return user, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&wg,
		)

		if err := svc.ForgotPassword(context.Background(), "user@example.com"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updated == nil {
			t.Fatal("expected user update")
		}
		if updated.PasswordResetToken == nil {
			t.Fatal("expected password reset token")
		}
		if updated.PasswordResetExpiresAt == nil {
			t.Fatal("expected password reset expiration")
		}

		wg.Wait()
	})
}

func TestUserService_ResetPassword(t *testing.T) {
	t.Parallel()

	validHash, err := password.HashIt("old-password")
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	t.Run("success clears reset state", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		expires := time.Now().Add(time.Hour)
		repo := &MockUserRepository{
			GetUserByPasswordResetTokenFunc: func(
				_ context.Context,
				_ string,
			) (*users.User, error) {
				return &users.User{
					ID:                     1,
					Email:                  "user@example.com",
					PasswordHash:           validHash,
					PasswordResetToken:     stringPtr("token"),
					PasswordResetExpiresAt: &expires,
				}, nil
			},
			UpdateUserFunc: func(
				_ context.Context,
				user *users.User,
			) (*users.User, error) {
				return user, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&wg,
		)

		if err := svc.ResetPassword(
			context.Background(),
			"token",
			"new-password",
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wg.Wait()
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()

		expires := time.Now().Add(-time.Hour)

		repo := &MockUserRepository{
			GetUserByPasswordResetTokenFunc: func(
				_ context.Context,
				_ string,
			) (*users.User, error) {
				return &users.User{
					ID:                     1,
					PasswordResetExpiresAt: &expires,
				}, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		if err := svc.ResetPassword(
			context.Background(),
			"token",
			"new-password",
		); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUserService_ListUsers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		limit      int32
		offset     int32
		wantLimit  int32
		wantOffset int32
	}{
		{name: "valid", limit: 10, offset: 5, wantLimit: 10, wantOffset: 5},
		{name: "default limit", limit: 0, offset: 5, wantLimit: 50, wantOffset: 5},
		{name: "max limit", limit: 101, offset: 5, wantLimit: 50, wantOffset: 5},
		{name: "negative offset", limit: 10, offset: -5, wantLimit: 10, wantOffset: 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotLimit, gotOffset int32

			repo := &MockUserRepository{
				ListUsersFunc: func(
					_ context.Context,
					limit, offset int32,
				) ([]*users.User, error) {
					gotLimit = limit
					gotOffset = offset
					return []*users.User{}, nil
				},
				CountUsersFunc: func(_ context.Context) (int64, error) {
					return 0, nil
				},
			}

			svc := newUserService(
				repo,
				&MockAgreementRepository{},
				&MockAuditService{},
				&MockEmailService{},
				&MockAPIKeyService{},
				&sync.WaitGroup{},
			)

			if _, _, err := svc.ListUsers(
				context.Background(),
				tt.limit,
				tt.offset,
			); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotLimit != tt.wantLimit {
				t.Fatalf("expected limit %d, got %d", tt.wantLimit, gotLimit)
			}

			if gotOffset != tt.wantOffset {
				t.Fatalf("expected offset %d, got %d", tt.wantOffset, gotOffset)
			}
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	t.Parallel()

	validHash, err := password.HashIt("old-password")
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()

		repo := &MockUserRepository{
			GetUserByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*users.User, error) {
				return &users.User{
					ID:           1,
					PasswordHash: validHash,
				}, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		if err := svc.ChangePassword(
			context.Background(),
			1,
			"wrong-password",
			"new-password",
		); !errors.Is(err, users.ErrIncorrectPassword) {
			t.Fatalf("expected ErrIncorrectPassword, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		var updated *users.User

		repo := &MockUserRepository{
			GetUserByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*users.User, error) {
				return &users.User{
					ID:           1,
					Email:        "user@example.com",
					PasswordHash: validHash,
				}, nil
			},
			UpdateUserFunc: func(
				_ context.Context,
				user *users.User,
			) (*users.User, error) {
				updated = user
				return user, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&wg,
		)

		if err := svc.ChangePassword(
			context.Background(),
			1,
			"old-password",
			"new-password",
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updated == nil || updated.PasswordHash == "" {
			t.Fatal("expected password hash update")
		}

		wg.Wait()
	})
}

func TestUserService_ToggleStatus(t *testing.T) {
	t.Parallel()

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()

		svc := newUserService(
			&MockUserRepository{},
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		if err := svc.ToggleStatus(
			context.Background(),
			1,
			users.AccountStatus("invalid"),
		); !errors.Is(err, users.ErrInvalidStatus) {
			t.Fatalf("expected ErrInvalidStatus, got %v", err)
		}
	})

	t.Run("same status is a no-op", func(t *testing.T) {
		t.Parallel()

		var updates int

		repo := &MockUserRepository{
			GetUserByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*users.User, error) {
				return &users.User{
					ID:     1,
					Status: users.AccountStatusActive,
				}, nil
			},
			UpdateUserFunc: func(
				_ context.Context,
				_ *users.User,
			) (*users.User, error) {
				updates++
				return nil, errors.New("must not update")
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		if err := svc.ToggleStatus(
			context.Background(),
			1,
			users.AccountStatusActive,
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updates != 0 {
			t.Fatalf("expected no update, got %d", updates)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		var updated *users.User

		repo := &MockUserRepository{
			GetUserByIDFunc: func(
				_ context.Context,
				_ uint64,
			) (*users.User, error) {
				return &users.User{
					ID:     1,
					Email:  "user@example.com",
					Status: users.AccountStatusPending,
				}, nil
			},
			UpdateUserFunc: func(
				_ context.Context,
				user *users.User,
			) (*users.User, error) {
				updated = user
				return user, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&wg,
		)

		if err := svc.ToggleStatus(
			context.Background(),
			1,
			users.AccountStatusActive,
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updated == nil || updated.Status != users.AccountStatusActive {
			t.Fatal("expected status to be updated")
		}

		wg.Wait()
	})
}

func TestUserService_GetMeAndDeleteAccount(t *testing.T) {
	t.Parallel()

	t.Run("GetMe", func(t *testing.T) {
		t.Parallel()

		repo := &MockUserRepository{
			GetUserByIDFunc: func(
				_ context.Context,
				id uint64,
			) (*users.User, error) {
				return &users.User{ID: id}, nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&sync.WaitGroup{},
		)

		got, err := svc.GetMe(context.Background(), 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == nil || got.ID != 42 {
			t.Fatalf("expected user 42, got %#v", got)
		}
	})

	t.Run("DeleteAccount", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		var deletedID uint64

		repo := &MockUserRepository{
			GetUserByIDFunc: func(
				_ context.Context,
				id uint64,
			) (*users.User, error) {
				return &users.User{
					ID:    id,
					Email: "user@example.com",
				}, nil
			},
			DeleteUserFunc: func(
				_ context.Context,
				id uint64,
			) error {
				deletedID = id
				return nil
			},
		}

		svc := newUserService(
			repo,
			&MockAgreementRepository{},
			&MockAuditService{},
			&MockEmailService{},
			&MockAPIKeyService{},
			&wg,
		)

		if err := svc.DeleteAccount(context.Background(), 42); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deletedID != 42 {
			t.Fatalf("expected deleted ID 42, got %d", deletedID)
		}

		wg.Wait()
	})
}

func stringPtr(value string) *string {
	return &value
}
