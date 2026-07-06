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

type UserService struct {
	repo      Repository
	auditRepo audit.Repository
}

func NewUserService(repo Repository, auditRepo audit.Repository) *UserService {
	return &UserService{
		repo:      repo,
		auditRepo: auditRepo,
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
		_, _ = s.auditRepo.Create(ctx, &audit.AuditEvent{
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

	_, _ = s.auditRepo.Create(ctx, &audit.AuditEvent{
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
