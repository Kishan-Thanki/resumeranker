package users

import "context"

type Repository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id uint64) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context, limit, offset int32) ([]*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id uint64) error
	GetUserByVerificationToken(ctx context.Context, token string) (*User, error)
	GetUserByPasswordResetToken(ctx context.Context, token string) (*User, error)
	VerifyUserEmail(ctx context.Context, id uint64) (*User, error)

	CreateAgreement(ctx context.Context, agreement *Agreement) (*Agreement, error)
	GetAgreementByID(ctx context.Context, id uint64) (*Agreement, error)
	GetAgreementByTypeAndVersion(ctx context.Context, agType AgreementType, version string) (*Agreement, error)

	CreateUserAgreement(ctx context.Context, userAgreement *UserAgreement) (*UserAgreement, error)
	HasUserAcceptedAgreement(ctx context.Context, userID uint64, agreementID uint64) (bool, error)
	GetLatestAgreements(ctx context.Context) ([]*Agreement, error)
	GetPendingAgreementsForUser(ctx context.Context, userID uint64) ([]*Agreement, error)
}
