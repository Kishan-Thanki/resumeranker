package users

type RegisterRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	AgreedToTerms bool   `json:"agreed_to_terms"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	ID      uint64 `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Content string `json:"content"`
}

type AcceptAgreementsRequest struct {
	AgreementIDs []uint64 `json:"agreement_ids"`
}
