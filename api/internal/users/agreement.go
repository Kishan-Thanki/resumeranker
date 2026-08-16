package users

import "time"

type AgreementType string

const (
	AgreementTypeTermsOfService AgreementType = "terms_of_service"
	AgreementTypePrivacyPolicy  AgreementType = "privacy_policy"
)

func (t AgreementType) IsValid() bool {
	switch t {
	case AgreementTypeTermsOfService, AgreementTypePrivacyPolicy:
		return true
	}

	return false
}

type Agreement struct {
	ID          uint64        `json:"id"`
	Type        AgreementType `json:"type"`
	Version     string        `json:"version"`
	Content     string        `json:"content"`
	PublishedAt time.Time     `json:"published_at"`
}

type UserAgreement struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	AgreementID uint64    `json:"agreement_id"`
	AcceptedAt  time.Time `json:"accepted_at"`
}
