package apikey

import "time"

type GenerateKeyRequest struct {
	Name  string `json:"name"`
	Quota uint64 `json:"quota"`
}

type GenerateKeyResponse struct {
	Message string `json:"message"`
	Key     string `json:"key"`
	KeyID   uint64 `json:"key_id"`
}

type UpdateKeyStatusRequest struct {
	Status APIKeyStatus `json:"status"`
}

type APIKeyDTO struct {
	ID                uint64       `json:"id"`
	Name              string       `json:"name"`
	KeyPrefix         string       `json:"key_prefix"`
	Status            APIKeyStatus `json:"status"`
	RequestsPerMinute uint64       `json:"requests_per_minute"`
	RequestsPerDay    uint64       `json:"requests_per_day"`
	TokenQuota        uint64       `json:"token_quota"`
	TokensUsed        uint64       `json:"tokens_used"`
	ExpiresAt         *time.Time   `json:"expires_at"`
	LastUsedAt        *time.Time   `json:"last_used_at"`
	CreatedAt         time.Time    `json:"created_at"`
}

type APIKeyUsageResponse struct {
	RPMUsed    int    `json:"rpm_used"`
	RPMLimit   uint64 `json:"rpm_limit"`
	RPDUsed    int    `json:"rpd_used"`
	RPDLimit   uint64 `json:"rpd_limit"`
	TokensUsed uint64 `json:"tokens_used"`
	TokenQuota uint64 `json:"token_quota"`
}
