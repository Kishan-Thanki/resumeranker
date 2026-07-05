package service

import (
	"context"
	"errors"

	"github.com/kishan-thanki/resumeranker/api/internal/domain"
	"github.com/kishan-thanki/resumeranker/api/pkg/hashutil"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uint64) (*domain.User, error)
}

type AgreementRepository interface {
	GetByTypeAndVersion(ctx context.Context, agType domain.AgreementType, version string) (*domain.Agreement, error)
}

type UserAgreementRepository interface {
	Create(ctx context.Context, userAgreement *domain.UserAgreement) (*domain.UserAgreement, error)
	HasAccepted(ctx context.Context, userID uint64, agreementID uint64) (bool, error)
}

type UserService struct {
	userRepo      UserRepository
	agreementRepo AgreementRepository
	userAgRepo    UserAgreementRepository
}

func NewUserService(ur UserRepository, ar AgreementRepository, uar UserAgreementRepository) *UserService {
	return &UserService{
		userRepo:      ur,
		agreementRepo: ar,
		userAgRepo:    uar,
	}
}

func (s *UserService) Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error) {
	hash, err := hashutil.HashIt(password)
	if err != nil {
		return nil, errors.Join(errors.New("failed to hash password"), err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		Status:       domain.AccountStatusActive,
	}

	return s.userRepo.Create(ctx, user)
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	match, err := hashutil.VerifyHash(password, user.PasswordHash)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}

	if user.Status != domain.AccountStatusActive {
		return nil, errors.New("account is not active")
	}

	return user, nil
}

func (s *UserService) AcceptTerms(ctx context.Context, userID uint64, version string) error {
	agreement, err := s.agreementRepo.GetByTypeAndVersion(ctx, domain.AgreementTypeTermsOfService, version)
	if err != nil {
		return errors.Join(errors.New("failed to fetch agreement"), err)
	}

	userAgreement := &domain.UserAgreement{
		UserID:      userID,
		AgreementID: agreement.ID,
	}

	_, err = s.userAgRepo.Create(ctx, userAgreement)
	return err
}

func (s *UserService) HasAcceptedTerms(ctx context.Context, userID uint64, version string) (bool, error) {
	agreement, err := s.agreementRepo.GetByTypeAndVersion(ctx, domain.AgreementTypeTermsOfService, version)
	if err != nil {
		return false, err
	}

	return s.userAgRepo.HasAccepted(ctx, userID, agreement.ID)
}
