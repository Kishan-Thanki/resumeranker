package users

import (
	"context"
	"errors"

	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/password"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type auditService interface {
	LogEvent(ctx context.Context, event *audit.AuditEvent) error
}

type UserService struct {
	repo         Repository
	auditService auditService
}

func NewUserService(repo Repository, auditService auditService) *UserService {
	return &UserService{
		repo:         repo,
		auditService: auditService,
	}
}

func (s *UserService) Register(ctx context.Context, email, passwordStr string, role Role) (*User, error) {
	hashedPassword, err := password.HashIt(passwordStr)
	if err != nil {
		return nil, errors.Join(errors.New("failed to hash password"), err)
	}

	user := &User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
		Status:       AccountStatusActive,
	}

	createdUser, err := s.repo.CreateUser(ctx, user)
	if err == nil {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventUserRegistered,
			Description: "user registered successfully",
			UserID:      &createdUser.ID,
		})
	}
	return createdUser, err
}

func (s *UserService) Authenticate(ctx context.Context, email, passwordStr string) (*User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	match, err := password.VerifyHash(passwordStr, user.PasswordHash)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}

	if user.Status != AccountStatusActive {
		return nil, errors.New("account is not active")
	}

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventUserLoggedIn,
		Description: "user logged in successfully",
		UserID:      &user.ID,
	})
	return user, nil
}

func (s *UserService) AcceptTerms(ctx context.Context, userID uint64, version string) error {
	agreement, err := s.repo.GetAgreementByTypeAndVersion(ctx, AgreementTypeTermsOfService, version)
	if err != nil {
		return errors.Join(errors.New("failed to fetch agreement"), err)
	}

	userAgreement := &UserAgreement{
		UserID:      userID,
		AgreementID: agreement.ID,
	}

	_, err = s.repo.CreateUserAgreement(ctx, userAgreement)
	return err
}

func (s *UserService) HasAcceptedTerms(ctx context.Context, userID uint64, version string) (bool, error) {
	agreement, err := s.repo.GetAgreementByTypeAndVersion(ctx, AgreementTypeTermsOfService, version)
	if err != nil {
		return false, err
	}

	return s.repo.HasUserAcceptedAgreement(ctx, userID, agreement.ID)
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	match, err := password.VerifyHash(oldPassword, user.PasswordHash)
	if err != nil || !match {
		return errors.New("incorrect old password")
	}

	hashedPassword, err := password.HashIt(newPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.PasswordHash = hashedPassword
	_, err = s.repo.UpdateUser(ctx, user)
	if err == nil {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventUserPasswordChanged,
			Description: "user changed password successfully",
			UserID:      &userID,
		})
	}
	return err
}

func (s *UserService) ToggleStatus(ctx context.Context, userID uint64, status AccountStatus) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Status == status {
		return nil
	}

	user.Status = status
	_, err = s.repo.UpdateUser(ctx, user)
	return err
}

func (s *UserService) DeleteAccount(ctx context.Context, userID uint64) error {
	return s.repo.DeleteUser(ctx, userID)
}
