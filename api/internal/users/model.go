package users

import (
	"errors"
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
	AccountStatusPending   AccountStatus = "pending"
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
)

func (s AccountStatus) IsValid() bool {
	switch s {
	case AccountStatusPending, AccountStatusActive, AccountStatusSuspended:
		return true
	}
	return false
}

type AgreementType string

const (
	AgreementTypeTermsOfService AgreementType = "terms_of_service"
	AgreementTypePrivacyPolicy  AgreementType = "privacy_policy"
)

var (
	ErrAccountSuspended   = errors.New("account is suspended")
	ErrMustAgreeToTerms   = errors.New("must agree to terms of service and privacy policy")
	ErrIncorrectPassword  = errors.New("incorrect old password")
	ErrInvalidStatus      = errors.New("invalid account status")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

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

type User struct {
	ID                     uint64        `json:"id"`
	Email                  string        `json:"email"`
	PasswordHash           string        `json:"-"`
	Role                   Role          `json:"role"`
	Status                 AccountStatus `json:"status"`
	Metadata               []byte        `json:"metadata"`
	IsVerified             bool          `json:"is_verified"`
	VerificationToken      *string       `json:"-"`
	VerificationExpiresAt  *time.Time    `json:"-"`
	PasswordResetToken     *string       `json:"-"`
	PasswordResetExpiresAt *time.Time    `json:"-"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
	DeletedAt              *time.Time    `json:"-"`
}
