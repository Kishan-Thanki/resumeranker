package users_test

import (
	"context"

	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

type MockRepository struct {
	CreateUserFunc                   func(ctx context.Context, user *users.User) (*users.User, error)
	GetUserByIDFunc                  func(ctx context.Context, id uint64) (*users.User, error)
	GetUserByEmailFunc               func(ctx context.Context, email string) (*users.User, error)
	UpdateUserFunc                   func(ctx context.Context, user *users.User) (*users.User, error)
	DeleteUserFunc                   func(ctx context.Context, id uint64) error
	CreateAgreementFunc              func(ctx context.Context, agreement *users.Agreement) (*users.Agreement, error)
	GetAgreementByIDFunc             func(ctx context.Context, id uint64) (*users.Agreement, error)
	GetAgreementByTypeAndVersionFunc func(ctx context.Context, agType users.AgreementType, version string) (*users.Agreement, error)
	CreateUserAgreementFunc          func(ctx context.Context, userAgreement *users.UserAgreement) (*users.UserAgreement, error)
	HasUserAcceptedAgreementFunc     func(ctx context.Context, userID uint64, agreementID uint64) (bool, error)
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

type MockAuditService struct {
	LogEventFunc func(ctx context.Context, event *audit.AuditEvent) error
}

func (m *MockAuditService) LogEvent(ctx context.Context, event *audit.AuditEvent) error {
	if m.LogEventFunc != nil {
		return m.LogEventFunc(ctx, event)
	}
	return nil
}

type MockUserService struct {
	RegisterFunc         func(ctx context.Context, email, password string, role users.Role) (*users.User, error)
	AuthenticateFunc     func(ctx context.Context, email, password string) (*users.User, error)
	AcceptTermsFunc      func(ctx context.Context, userID uint64, version string) error
	HasAcceptedTermsFunc func(ctx context.Context, userID uint64, version string) (bool, error)
	ChangePasswordFunc   func(ctx context.Context, userID uint64, oldPassword, newPassword string) error
	ToggleStatusFunc     func(ctx context.Context, userID uint64, status users.AccountStatus) error
	DeleteAccountFunc    func(ctx context.Context, userID uint64) error
}

func (m *MockUserService) Register(ctx context.Context, email, password string, role users.Role) (*users.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, email, password, role)
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
