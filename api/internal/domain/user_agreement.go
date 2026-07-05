package domain

import "time"

type UserAgreement struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	AgreementID uint64    `json:"agreement_id"`
	AcceptedAt  time.Time `json:"accepted_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
