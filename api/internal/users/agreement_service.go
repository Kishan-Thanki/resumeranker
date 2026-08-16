package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/config"
	emailpkg "github.com/kishan-thanki/resumeranker/api/internal/email"
)

var (
	ErrInvalidAgreementType    = errors.New("invalid agreement type")
	ErrInvalidAgreementVersion = errors.New("agreement version is required")
	ErrInvalidAgreementContent = errors.New("agreement content is required")
	ErrInvalidAgreementID      = errors.New("agreement id must be greater than zero")
	ErrNoAgreementsProvided    = errors.New("at least one agreement is required")
)

type AgreementService struct {
	repo     AgreementRepository
	emailSvc emailService
	cfg      *config.Config
}

func NewAgreementService(
	repo AgreementRepository,
	emailSvc emailService,
	cfg *config.Config,
) *AgreementService {
	return &AgreementService{
		repo:     repo,
		emailSvc: emailSvc,
		cfg:      cfg,
	}
}

func (s *AgreementService) PublishAgreement(
	ctx context.Context,
	agType AgreementType,
	version,
	content string,
	notifyUsers bool,
	userRepo UserRepository,
) (*Agreement, error) {
	if !agType.IsValid() {
		return nil, ErrInvalidAgreementType
	}

	version = strings.TrimSpace(version)
	if version == "" {
		return nil, ErrInvalidAgreementVersion
	}

	if strings.TrimSpace(content) == "" {
		return nil, ErrInvalidAgreementContent
	}

	agreement := &Agreement{
		Type:        agType,
		Version:     version,
		Content:     content,
		PublishedAt: time.Now().UTC(),
	}

	created, err := s.repo.CreateAgreement(ctx, agreement)
	if err != nil {
		return nil, fmt.Errorf("failed to create agreement: %w", err)
	}

	if notifyUsers && userRepo != nil {
		s.notifyUsersOfAgreementUpdate(agType, version, userRepo)
	}

	return created, nil
}

func (s *AgreementService) notifyUsersOfAgreementUpdate(
	agType AgreementType,
	version string,
	userRepo UserRepository,
) {
	go func() {
		bgCtx := context.Background()
		const pageSize int32 = 100
		offset := int32(0)
		errorCount := 0

		for {
			userList, err := userRepo.ListUsers(bgCtx, pageSize, offset)
			if err != nil {
				slog.Error(
					"failed to list users for agreement notification",
					"error", err,
					"offset", offset,
				)

				errorCount++
				if errorCount > 3 {
					return
				}

				time.Sleep(2 * time.Second)
				continue
			}

			if len(userList) == 0 {
				return
			}

			errorCount = 0

			for _, user := range userList {
				if user == nil || user.Status != AccountStatusActive {
					continue
				}

				if s.emailSvc == nil {
					continue
				}

				htmlBody := fmt.Sprintf(
					"<p>An agreement you accepted (%s) has been updated to version %s.</p>",
					string(agType),
					version,
				)

				if err := s.emailSvc.SendEmail(
					bgCtx,
					&emailpkg.SendEmailRequest{
						To:      []string{user.Email},
						Subject: "Agreement Update",
						Text:    "An agreement has been updated.",
						HTML:    htmlBody,
					},
				); err != nil {
					slog.Error(
						"failed to send agreement update email",
						"user_id", user.ID,
						"error", err,
					)
				}
			}

			offset += pageSize
		}
	}()
}

func (s *AgreementService) GetLatestAgreements(
	ctx context.Context,
) ([]*Agreement, error) {
	return s.repo.GetLatestAgreements(ctx)
}

func (s *AgreementService) GetPendingAgreements(
	ctx context.Context,
	userID uint64,
) ([]*Agreement, error) {
	return s.repo.GetPendingAgreementsForUser(ctx, userID)
}

func (s *AgreementService) AcceptTerms(
	ctx context.Context,
	userID uint64,
	termsVersion,
	privacyVersion string,
) error {
	termsVersion = strings.TrimSpace(termsVersion)
	privacyVersion = strings.TrimSpace(privacyVersion)

	if termsVersion == "" {
		return errors.New("invalid terms of service version")
	}
	if privacyVersion == "" {
		return errors.New("invalid privacy policy version")
	}

	terms, err := s.repo.GetAgreementByTypeAndVersion(
		ctx,
		AgreementTypeTermsOfService,
		termsVersion,
	)
	if err != nil {
		return errors.New("invalid terms of service version")
	}

	privacy, err := s.repo.GetAgreementByTypeAndVersion(
		ctx,
		AgreementTypePrivacyPolicy,
		privacyVersion,
	)
	if err != nil {
		return errors.New("invalid privacy policy version")
	}

	if _, err := s.repo.CreateUserAgreement(
		ctx,
		&UserAgreement{
			UserID:      userID,
			AgreementID: terms.ID,
			AcceptedAt:  time.Now().UTC(),
		},
	); err != nil {
		return fmt.Errorf("failed to accept terms of service: %w", err)
	}

	if _, err := s.repo.CreateUserAgreement(
		ctx,
		&UserAgreement{
			UserID:      userID,
			AgreementID: privacy.ID,
			AcceptedAt:  time.Now().UTC(),
		},
	); err != nil {
		return fmt.Errorf("failed to accept privacy policy: %w", err)
	}

	return nil
}

func (s *AgreementService) HasAcceptedTerms(
	ctx context.Context,
	userID uint64,
	termsVersion,
	privacyVersion string,
) (bool, error) {
	termsVersion = strings.TrimSpace(termsVersion)
	privacyVersion = strings.TrimSpace(privacyVersion)

	if termsVersion == "" || privacyVersion == "" {
		return false, nil
	}

	terms, err := s.repo.GetAgreementByTypeAndVersion(
		ctx,
		AgreementTypeTermsOfService,
		termsVersion,
	)
	if err != nil {
		return false, err
	}

	privacy, err := s.repo.GetAgreementByTypeAndVersion(
		ctx,
		AgreementTypePrivacyPolicy,
		privacyVersion,
	)
	if err != nil {
		return false, err
	}

	acceptedTerms, err := s.repo.HasUserAcceptedAgreement(
		ctx,
		userID,
		terms.ID,
	)
	if err != nil {
		return false, err
	}
	if !acceptedTerms {
		return false, nil
	}

	acceptedPrivacy, err := s.repo.HasUserAcceptedAgreement(
		ctx,
		userID,
		privacy.ID,
	)
	if err != nil {
		return false, err
	}

	return acceptedPrivacy, nil
}

func (s *AgreementService) AcceptAgreements(
	ctx context.Context,
	userID uint64,
	agreementIDs []uint64,
) error {
	if len(agreementIDs) == 0 {
		return ErrNoAgreementsProvided
	}

	seen := make(map[uint64]struct{}, len(agreementIDs))

	for _, id := range agreementIDs {
		if id == 0 {
			return ErrInvalidAgreementID
		}

		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}

		if _, err := s.repo.CreateUserAgreement(
			ctx,
			&UserAgreement{
				UserID:      userID,
				AgreementID: id,
				AcceptedAt:  time.Now().UTC(),
			},
		); err != nil {
			return fmt.Errorf("failed to accept agreement %d: %w", id, err)
		}
	}

	return nil
}
