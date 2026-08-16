package users

import "time"

type RegisterRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	AgreedToTerms bool   `json:"agreed_to_terms"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ToggleStatusRequest struct {
	Status AccountStatus `json:"status"`
}

type AgreementResponse struct {
	ID          uint64 `json:"id"`
	Type        string `json:"type"`
	Version     string `json:"version"`
	Content     string `json:"content"`
	PublishedAt string `json:"published_at"`
}

type AcceptAgreementsRequest struct {
	AgreementIDs []uint64 `json:"agreement_ids"`
}

type PublishAgreementRequest struct {
	Type    AgreementType `json:"type"`
	Version string        `json:"version"`
	Content string        `json:"content"`
}

type UserResponse struct {
	ID         uint64        `json:"id"`
	Email      string        `json:"email"`
	Role       Role          `json:"role"`
	Status     AccountStatus `json:"status"`
	IsVerified bool          `json:"is_verified"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

func toAgreementResponse(a *Agreement) AgreementResponse {
	return AgreementResponse{
		ID:          a.ID,
		Type:        string(a.Type),
		Version:     a.Version,
		Content:     a.Content,
		PublishedAt: a.PublishedAt.Format(time.RFC3339),
	}
}

func toUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:         u.ID,
		Email:      u.Email,
		Role:       u.Role,
		Status:     u.Status,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  u.UpdatedAt.Format(time.RFC3339),
	}
}
