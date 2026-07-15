package users

import (
	"context"

	"github.com/kishan-thanki/resumeranker/api/internal/config"
)

type Fixtures struct {
	TermsOfService struct {
		Version string `json:"version"`
		Content string `json:"content"`
	} `json:"terms_of_service"`
	PrivacyPolicy struct {
		Version string `json:"version"`
		Content string `json:"content"`
	} `json:"privacy_policy"`
	Users []struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		Role          string `json:"role"`
		Status        string `json:"status"`
		AgreedToTerms bool   `json:"agreed_to_terms"`
	} `json:"users"`
}

func SeedFromFixtures(ctx context.Context, userSvc *UserService, agreementSvc *AgreementService, cfg *config.Config, fixtures *Fixtures) error {
	for _, user := range fixtures.Users {
		_, err := userSvc.Register(ctx, user.Email, user.Password, Role(user.Role), user.AgreedToTerms)
		if err != nil {
			return err
		}
	}

	_, err := agreementSvc.PublishAgreement(ctx, AgreementTypeTermsOfService, fixtures.TermsOfService.Version, fixtures.TermsOfService.Content, false, nil)
	if err != nil {
		return err
	}

	_, err = agreementSvc.PublishAgreement(ctx, AgreementTypePrivacyPolicy, fixtures.PrivacyPolicy.Version, fixtures.PrivacyPolicy.Content, false, nil)
	if err != nil {
		return err
	}

	return nil
}
