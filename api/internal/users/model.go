package users

import (
	"encoding/json"
	"time"
)

type Role string

const (
	RoleOwner Role = "owner"
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusDeleted   AccountStatus = "deleted"
)

type User struct {
	ID                     uint64          `json:"id"`
	Email                  string          `json:"email"`
	PasswordHash           string          `json:"-"`
	Role                   Role            `json:"role"`
	Status                 AccountStatus   `json:"status"`
	Metadata               json.RawMessage `json:"metadata,omitempty"`
	IsVerified             bool            `json:"is_verified"`
	VerificationToken      *string         `json:"-"`
	VerificationExpiresAt  *time.Time      `json:"-"`
	PasswordResetToken     *string         `json:"-"`
	PasswordResetExpiresAt *time.Time      `json:"-"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	DeletedAt              *time.Time      `json:"deleted_at"`
}

type AgreementType string

const (
	AgreementTypeTermsOfService AgreementType = "terms_of_service"
	AgreementTypePrivacyPolicy  AgreementType = "privacy_policy"
)

type Agreement struct {
	ID          uint64        `json:"id"`
	Type        AgreementType `json:"type"`
	Version     string        `json:"version"`
	Content     string        `json:"content"`
	PublishedAt time.Time     `json:"published_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type UserAgreement struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	AgreementID uint64    `json:"agreement_id"`
	AcceptedAt  time.Time `json:"accepted_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
