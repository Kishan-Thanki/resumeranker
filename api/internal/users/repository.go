package users

import "context"

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id uint64) error

	GetUserByID(ctx context.Context, id uint64) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	GetUserByVerificationToken(ctx context.Context, token string) (*User, error)
	GetUserByPasswordResetToken(ctx context.Context, token string) (*User, error)
	VerifyUserEmail(ctx context.Context, id uint64) (*User, error)

	ListUsers(ctx context.Context, limit, offset int32) ([]*User, error)
	CountUsers(ctx context.Context) (int64, error)
}

type AgreementRepository interface {
	CreateAgreement(ctx context.Context, agreement *Agreement) (*Agreement, error)
	GetAgreementByID(ctx context.Context, id uint64) (*Agreement, error)
	GetAgreementByTypeAndVersion(ctx context.Context, agType AgreementType, version string) (*Agreement, error)
	GetLatestAgreements(ctx context.Context) ([]*Agreement, error)

	CreateUserAgreement(ctx context.Context, userAgreement *UserAgreement) (*UserAgreement, error)
	HasUserAcceptedAgreement(ctx context.Context, userID uint64, agreementID uint64) (bool, error)
	GetPendingAgreementsForUser(ctx context.Context, userID uint64) ([]*Agreement, error)
}
