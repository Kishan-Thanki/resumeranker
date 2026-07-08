package users_test

import (
	"context"

	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/email"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

type MockRepository struct {
	CreateUserFunc                   func(ctx context.Context, user *users.User) (*users.User, error)
	GetUserByIDFunc                  func(ctx context.Context, id uint64) (*users.User, error)
	GetUserByEmailFunc               func(ctx context.Context, email string) (*users.User, error)
	ListUsersFunc                    func(ctx context.Context, limit, offset int32) ([]*users.User, error)
	UpdateUserFunc                   func(ctx context.Context, user *users.User) (*users.User, error)
	DeleteUserFunc                   func(ctx context.Context, id uint64) error
	GetUserByVerificationTokenFunc   func(ctx context.Context, token string) (*users.User, error)
	GetUserByPasswordResetTokenFunc  func(ctx context.Context, token string) (*users.User, error)
	VerifyUserEmailFunc              func(ctx context.Context, id uint64) (*users.User, error)
	CreateAgreementFunc              func(ctx context.Context, agreement *users.Agreement) (*users.Agreement, error)
	GetAgreementByIDFunc             func(ctx context.Context, id uint64) (*users.Agreement, error)
	GetAgreementByTypeAndVersionFunc func(ctx context.Context, agType users.AgreementType, version string) (*users.Agreement, error)
	CreateUserAgreementFunc          func(ctx context.Context, userAgreement *users.UserAgreement) (*users.UserAgreement, error)
	HasUserAcceptedAgreementFunc     func(ctx context.Context, userID uint64, agreementID uint64) (bool, error)
	GetLatestAgreementsFunc          func(ctx context.Context) ([]*users.Agreement, error)
	GetPendingAgreementsForUserFunc  func(ctx context.Context, userID uint64) ([]*users.Agreement, error)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *users.User) (*users.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return user, nil
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uint64) (*users.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockRepository) ListUsers(ctx context.Context, limit, offset int32) ([]*users.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, limit, offset)
	}
	return nil, nil
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *users.User) (*users.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	return user, nil
}

func (m *MockRepository) DeleteUser(ctx context.Context, id uint64) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

func (m *MockRepository) GetUserByVerificationToken(ctx context.Context, token string) (*users.User, error) {
	if m.GetUserByVerificationTokenFunc != nil {
		return m.GetUserByVerificationTokenFunc(ctx, token)
	}
	return nil, nil
}

func (m *MockRepository) GetUserByPasswordResetToken(ctx context.Context, token string) (*users.User, error) {
	if m.GetUserByPasswordResetTokenFunc != nil {
		return m.GetUserByPasswordResetTokenFunc(ctx, token)
	}
	return nil, nil
}

func (m *MockRepository) VerifyUserEmail(ctx context.Context, id uint64) (*users.User, error) {
	if m.VerifyUserEmailFunc != nil {
		return m.VerifyUserEmailFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) CreateAgreement(ctx context.Context, agreement *users.Agreement) (*users.Agreement, error) {
	if m.CreateAgreementFunc != nil {
		return m.CreateAgreementFunc(ctx, agreement)
	}
	return agreement, nil
}

func (m *MockRepository) GetAgreementByID(ctx context.Context, id uint64) (*users.Agreement, error) {
	if m.GetAgreementByIDFunc != nil {
		return m.GetAgreementByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) GetAgreementByTypeAndVersion(ctx context.Context, agType users.AgreementType, version string) (*users.Agreement, error) {
	if m.GetAgreementByTypeAndVersionFunc != nil {
		return m.GetAgreementByTypeAndVersionFunc(ctx, agType, version)
	}
	return nil, nil
}

func (m *MockRepository) CreateUserAgreement(ctx context.Context, userAgreement *users.UserAgreement) (*users.UserAgreement, error) {
	if m.CreateUserAgreementFunc != nil {
		return m.CreateUserAgreementFunc(ctx, userAgreement)
	}
	return userAgreement, nil
}

func (m *MockRepository) HasUserAcceptedAgreement(ctx context.Context, userID uint64, agreementID uint64) (bool, error) {
	if m.HasUserAcceptedAgreementFunc != nil {
		return m.HasUserAcceptedAgreementFunc(ctx, userID, agreementID)
	}
	return false, nil
}

func (m *MockRepository) GetLatestAgreements(ctx context.Context) ([]*users.Agreement, error) {
	if m.GetLatestAgreementsFunc != nil {
		return m.GetLatestAgreementsFunc(ctx)
	}
	return nil, nil
}

func (m *MockRepository) GetPendingAgreementsForUser(ctx context.Context, userID uint64) ([]*users.Agreement, error) {
	if m.GetPendingAgreementsForUserFunc != nil {
		return m.GetPendingAgreementsForUserFunc(ctx, userID)
	}
	return nil, nil
}

type MockAuditService struct {
	LogEventFunc func(ctx context.Context, event *audit.AuditEvent) error
}

func (m *MockAuditService) LogEvent(ctx context.Context, event *audit.AuditEvent) error {
	if m.LogEventFunc != nil {
		return m.LogEventFunc(ctx, event)
	}
	return nil
}

type MockEmailService struct {
	SendEmailFunc func(ctx context.Context, req *email.SendEmailRequest) error
}

func (m *MockEmailService) SendEmail(ctx context.Context, req *email.SendEmailRequest) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(ctx, req)
	}
	return nil
}

type MockUserService struct {
	RegisterFunc         func(ctx context.Context, email, password string, role users.Role, agreedToTerms bool) (*users.User, error)
	AuthenticateFunc     func(ctx context.Context, email, password string) (*users.User, error)
	AcceptTermsFunc      func(ctx context.Context, userID uint64, version string) error
	HasAcceptedTermsFunc func(ctx context.Context, userID uint64, version string) (bool, error)
	GetPendingAgreementsFunc func(ctx context.Context, userID uint64) ([]*users.Agreement, error)
	AcceptAgreementsFunc func(ctx context.Context, userID uint64, agreementIDs []uint64) error
	ChangePasswordFunc   func(ctx context.Context, userID uint64, oldPassword, newPassword string) error
	ToggleStatusFunc     func(ctx context.Context, userID uint64, status users.AccountStatus) error
	DeleteAccountFunc    func(ctx context.Context, userID uint64) error
	GetLatestAgreementsFunc func(ctx context.Context) ([]*users.Agreement, error)
	VerifyEmailFunc      func(ctx context.Context, token string) error
	PublishAgreementFunc func(ctx context.Context, agType users.AgreementType, version, content string) (*users.Agreement, error)
	ForgotPasswordFunc   func(ctx context.Context, email string) error
	ResetPasswordFunc    func(ctx context.Context, token, newPassword string) error
	GetMeFunc            func(ctx context.Context, userID uint64) (*users.User, error)
	ListUsersFunc        func(ctx context.Context, limit, offset int32) ([]*users.User, error)
}

func (m *MockUserService) Register(ctx context.Context, email, password string, role users.Role, agreedToTerms bool) (*users.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, email, password, role, agreedToTerms)
	}
	return nil, nil
}

func (m *MockUserService) Authenticate(ctx context.Context, email, password string) (*users.User, error) {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(ctx, email, password)
	}
	return nil, nil
}

func (m *MockUserService) AcceptTerms(ctx context.Context, userID uint64, version string) error {
	if m.AcceptTermsFunc != nil {
		return m.AcceptTermsFunc(ctx, userID, version)
	}
	return nil
}

func (m *MockUserService) HasAcceptedTerms(ctx context.Context, userID uint64, version string) (bool, error) {
	if m.HasAcceptedTermsFunc != nil {
		return m.HasAcceptedTermsFunc(ctx, userID, version)
	}
	return false, nil
}

func (m *MockUserService) GetPendingAgreements(ctx context.Context, userID uint64) ([]*users.Agreement, error) {
	if m.GetPendingAgreementsFunc != nil {
		return m.GetPendingAgreementsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserService) AcceptAgreements(ctx context.Context, userID uint64, agreementIDs []uint64) error {
	if m.AcceptAgreementsFunc != nil {
		return m.AcceptAgreementsFunc(ctx, userID, agreementIDs)
	}
	return nil
}

func (m *MockUserService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(ctx, userID, oldPassword, newPassword)
	}
	return nil
}

func (m *MockUserService) ToggleStatus(ctx context.Context, userID uint64, status users.AccountStatus) error {
	if m.ToggleStatusFunc != nil {
		return m.ToggleStatusFunc(ctx, userID, status)
	}
	return nil
}

func (m *MockUserService) DeleteAccount(ctx context.Context, userID uint64) error {
	if m.DeleteAccountFunc != nil {
		return m.DeleteAccountFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserService) GetLatestAgreements(ctx context.Context) ([]*users.Agreement, error) {
	if m.GetLatestAgreementsFunc != nil {
		return m.GetLatestAgreementsFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserService) VerifyEmail(ctx context.Context, token string) error {
	if m.VerifyEmailFunc != nil {
		return m.VerifyEmailFunc(ctx, token)
	}
	return nil
}

func (m *MockUserService) PublishAgreement(ctx context.Context, agType users.AgreementType, version, content string) (*users.Agreement, error) {
	if m.PublishAgreementFunc != nil {
		return m.PublishAgreementFunc(ctx, agType, version, content)
	}
	return nil, nil
}

func (m *MockUserService) ForgotPassword(ctx context.Context, email string) error {
	if m.ForgotPasswordFunc != nil {
		return m.ForgotPasswordFunc(ctx, email)
	}
	return nil
}

func (m *MockUserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(ctx, token, newPassword)
	}
	return nil
}

func (m *MockUserService) GetMe(ctx context.Context, userID uint64) (*users.User, error) {
	if m.GetMeFunc != nil {
		return m.GetMeFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int32) ([]*users.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, limit, offset)
	}
	return nil, nil
}
