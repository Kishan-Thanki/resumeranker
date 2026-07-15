package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kishan-thanki/resumeranker/api/internal/config"
	emailpkg "github.com/kishan-thanki/resumeranker/api/internal/email"
)

type AgreementService struct {
	repo     AgreementRepository
	emailSvc emailService
	cfg      *config.Config
}

func NewAgreementService(repo AgreementRepository, emailSvc emailService, cfg *config.Config) *AgreementService {
	return &AgreementService{
		repo:     repo,
		emailSvc: emailSvc,
		cfg:      cfg,
	}
}

func (s *AgreementService) PublishAgreement(ctx context.Context, agType AgreementType, version, content string, notifyUsers bool, userRepo UserRepository) (*Agreement, error) {
	agreement := &Agreement{
		Type:        agType,
		Version:     version,
		Content:     content,
		PublishedAt: time.Now(),
	}

	created, err := s.repo.CreateAgreement(ctx, agreement)
	if err != nil {
		return nil, fmt.Errorf("failed to create agreement: %w", err)
	}

	if notifyUsers && userRepo != nil {
		go func() {
			bgCtx := context.Background()
			limit := int32(100)
			offset := int32(0)
			errorCount := 0

			for {
				users, err := userRepo.ListUsers(bgCtx, limit, offset)
				if err != nil {
					slog.Error("failed to list users for agreement notification", "error", err, "offset", offset)
					errorCount++
					if errorCount > 3 {
						break
					}
					time.Sleep(2 * time.Second)
					continue
				}

				if len(users) == 0 {
					break
				}

				for _, user := range users {
					if user.Status == AccountStatusActive {

						htmlBody := fmt.Sprintf("<p>An agreement you accepted (%s) has been updated to version %s.</p>", string(agType), version)
						err := s.emailSvc.SendEmail(context.Background(), &emailpkg.SendEmailRequest{
							To:      []string{user.Email},
							Subject: "Agreement Update",
							Text:    "An agreement has been updated.",
							HTML:    htmlBody,
						})
						if err != nil {
							slog.Error("failed to send agreement update email", "user_id", user.ID, "error", err)
						}
					}
				}
				offset += limit
				errorCount = 0
			}
		}()
	}

	return created, nil
}

func (s *AgreementService) GetLatestAgreements(ctx context.Context) ([]*Agreement, error) {
	return s.repo.GetLatestAgreements(ctx)
}

func (s *AgreementService) GetPendingAgreements(ctx context.Context, userID uint64) ([]*Agreement, error) {
	return s.repo.GetPendingAgreementsForUser(ctx, userID)
}

func (s *AgreementService) AcceptTerms(ctx context.Context, userID uint64, termsVersion, privacyVersion string) error {
	terms, err := s.repo.GetAgreementByTypeAndVersion(ctx, AgreementTypeTermsOfService, termsVersion)
	if err != nil {
		return errors.New("invalid terms of service version")
	}

	privacy, err := s.repo.GetAgreementByTypeAndVersion(ctx, AgreementTypePrivacyPolicy, privacyVersion)
	if err != nil {
		return errors.New("invalid privacy policy version")
	}

	_, err = s.repo.CreateUserAgreement(ctx, &UserAgreement{
		UserID:      userID,
		AgreementID: terms.ID,
		AcceptedAt:  time.Now(),
	})
	if err != nil {
		return err
	}

	_, err = s.repo.CreateUserAgreement(ctx, &UserAgreement{
		UserID:      userID,
		AgreementID: privacy.ID,
		AcceptedAt:  time.Now(),
	})
	return err
}

func (s *AgreementService) HasAcceptedTerms(ctx context.Context, userID uint64, termsVersion, privacyVersion string) (bool, error) {
	terms, err := s.repo.GetAgreementByTypeAndVersion(ctx, AgreementTypeTermsOfService, termsVersion)
	if err != nil {
		return false, err
	}

	privacy, err := s.repo.GetAgreementByTypeAndVersion(ctx, AgreementTypePrivacyPolicy, privacyVersion)
	if err != nil {
		return false, err
	}

	acceptedTerms, err := s.repo.HasUserAcceptedAgreement(ctx, userID, terms.ID)
	if err != nil || !acceptedTerms {
		return false, err
	}

	acceptedPrivacy, err := s.repo.HasUserAcceptedAgreement(ctx, userID, privacy.ID)
	if err != nil || !acceptedPrivacy {
		return false, err
	}

	return true, nil
}

func (s *AgreementService) AcceptAgreements(ctx context.Context, userID uint64, agreementIDs []uint64) error {
	for _, id := range agreementIDs {
		_, err := s.repo.CreateUserAgreement(ctx, &UserAgreement{
			UserID:      userID,
			AgreementID: id,
			AcceptedAt:  time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to accept agreement %d: %w", id, err)
		}
	}
	return nil
}
