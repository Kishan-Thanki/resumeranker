package apikey

import "time"

type APIKeyStatus string

const (
	APIKeyStatusActive    APIKeyStatus = "active"
	APIKeyStatusInactive  APIKeyStatus = "inactive"
	APIKeyStatusSuspended APIKeyStatus = "suspended"
)

type APIKey struct {
	ID                uint64       `json:"id"`
	UserID            uint64       `json:"user_id"`
	Name              string       `json:"name"`
	KeyPrefix         string       `json:"key_prefix"`
	KeySelector       string       `json:"-"`
	KeyHash           string       `json:"-"`
	Status            APIKeyStatus `json:"status"`
	RequestsPerMinute uint64       `json:"requests_per_minute"`
	RequestsPerDay    uint64       `json:"requests_per_day"`
	TokenQuota        uint64       `json:"token_quota"`
	TokensUsed        uint64       `json:"tokens_used"`
	ExpiresAt         *time.Time   `json:"expires_at"`
	LastUsedAt        *time.Time   `json:"last_used_at"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	DeletedAt         *time.Time   `json:"deleted_at"`
}

type APIKeyUsageResponse struct {
	RPMUsed    int    `json:"rpm_used"`
	RPMLimit   uint64 `json:"rpm_limit"`
	RPDUsed    int    `json:"rpd_used"`
	RPDLimit   uint64 `json:"rpd_limit"`
	TokensUsed uint64 `json:"tokens_used"`
	TokenQuota uint64 `json:"token_quota"`
}
