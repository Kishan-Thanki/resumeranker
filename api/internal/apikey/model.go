package apikey

import "time"

type APIKeyStatus string

const (
	APIKeyStatusActive    APIKeyStatus = "active"
	APIKeyStatusInactive  APIKeyStatus = "inactive"
	APIKeyStatusSuspended APIKeyStatus = "suspended"
)

func (s APIKeyStatus) IsValid() bool {
	switch s {
	case APIKeyStatusActive, APIKeyStatusInactive, APIKeyStatusSuspended:
		return true
	}
	return false
}

// APIKey represents a credential owned by a user.
//
// The raw API key value is intentionally never stored. KeyPrefix and
// KeySelector support identification/lookup, while KeyHash is used to
// verify a presented secret.
type APIKey struct {
	ID          uint64 `json:"id"`
	UserID      uint64 `json:"user_id"`
	Name        string `json:"name"`
	KeyPrefix   string `json:"key_prefix"`
	KeySelector string `json:"key_selector"`
	KeyHash     string `json:"-"`

	Status APIKeyStatus `json:"status"`

	// RequestsPerMinute and RequestsPerDay define the configured request
	// limits for this key. Their enforcement window semantics belong to
	// the rate-limiting service.
	RequestsPerMinute uint64 `json:"requests_per_minute"`
	RequestsPerDay    uint64 `json:"requests_per_day"`

	// TokenQuota defines the configured token allowance for the key.
	// TokensUsed represents accumulated usage tracked by the application.
	// The reset/window policy for token usage belongs to the quota service.
	TokenQuota uint64 `json:"token_quota"`
	TokensUsed uint64 `json:"tokens_used"`

	ExpiresAt  *time.Time `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}
