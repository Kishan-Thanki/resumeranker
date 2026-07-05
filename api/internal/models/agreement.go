package models

import "time"

type AgreementType string

const (
	AgreementTypeTermsOfService AgreementType = "terms_of_service"
	AgreementTypePrivacyPolicy  AgreementType = "privacy_policy"
)

type Agreement struct {
	ID          uint64        `json:"id"`
	Type        AgreementType `json:"type"`
	Version     string        `json:"version"`
	DocumentURL string        `json:"document_url"`
	PublishedAt time.Time     `json:"published_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
