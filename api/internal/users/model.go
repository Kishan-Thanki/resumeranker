package users

import "time"

type Role string

const (
	RoleOwner Role = "owner"
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleUser:
		return true
	}
	return false
}

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

type User struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`

	PasswordHash string `json:"-"`

	Role   Role          `json:"role"`
	Status AccountStatus `json:"status"`

	Metadata []byte `json:"metadata"`

	IsVerified bool `json:"is_verified"`

	VerificationToken     *string    `json:"-"`
	VerificationExpiresAt *time.Time `json:"-"`

	PasswordResetToken     *string    `json:"-"`
	PasswordResetExpiresAt *time.Time `json:"-"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}
