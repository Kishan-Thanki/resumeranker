package users

import "errors"

var (
	ErrAccountSuspended   = errors.New("account is suspended")
	ErrMustAgreeToTerms   = errors.New("must agree to terms of service and privacy policy")
	ErrIncorrectPassword  = errors.New("incorrect old password")
	ErrInvalidStatus      = errors.New("invalid account status")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)
