package users

import (
	"context"
	"errors"

	"github.com/kishan-thanki/resumeranker/api/internal/pgutil"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/users/db"
)

type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) (*User, error) {

	u, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:                  user.Email,
		PasswordHash:           user.PasswordHash,
		Role:                   string(user.Role),
		Status:                 string(user.Status),
		Metadata:               user.Metadata,
		IsVerified:             user.IsVerified,
		VerificationToken:      pgutil.ToPgText(user.VerificationToken),
		VerificationExpiresAt:  pgutil.ToPgTimestamptz(user.VerificationExpiresAt),
		PasswordResetToken:     pgutil.ToPgText(user.PasswordResetToken),
		PasswordResetExpiresAt: pgutil.ToPgTimestamptz(user.PasswordResetExpiresAt),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	user.ID = uint64(u.ID)
	user.CreatedAt = u.CreatedAt.Time
	user.UpdatedAt = u.UpdatedAt.Time

	return user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uint64) (*User, error) {
	u, err := r.queries.GetUserByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return userFromRow(u), nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return userFromRow(u), nil
}

func (r *PostgresRepository) GetUserByVerificationToken(ctx context.Context, token string) (*User, error) {
	pgToken := pgutil.ToPgText(&token)
	u, err := r.queries.GetUserByVerificationToken(ctx, pgToken)
	if err != nil {
		return nil, err
	}
	return userFromRow(u), nil
}

func (r *PostgresRepository) GetUserByPasswordResetToken(ctx context.Context, token string) (*User, error) {
	pgToken := pgutil.ToPgText(&token)
	u, err := r.queries.GetUserByPasswordResetToken(ctx, pgToken)
	if err != nil {
		return nil, err
	}
	return userFromRow(u), nil
}

func (r *PostgresRepository) ListUsers(ctx context.Context, limit, offset int32) ([]*User, error) {
	users, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*User, len(users))
	for i, u := range users {
		result[i] = userFromRow(u)
	}
	return result, nil
}

// UpdateUser updates an existing user in the database.
// Note: This method mutates the input *User struct by setting the UpdatedAt field.
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) (*User, error) {

	u, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:                     int64(user.ID),
		Email:                  user.Email,
		PasswordHash:           user.PasswordHash,
		Role:                   string(user.Role),
		Status:                 string(user.Status),
		Metadata:               user.Metadata,
		IsVerified:             user.IsVerified,
		VerificationToken:      pgutil.ToPgText(user.VerificationToken),
		VerificationExpiresAt:  pgutil.ToPgTimestamptz(user.VerificationExpiresAt),
		PasswordResetToken:     pgutil.ToPgText(user.PasswordResetToken),
		PasswordResetExpiresAt: pgutil.ToPgTimestamptz(user.PasswordResetExpiresAt),
	})
	if err != nil {
		return nil, err
	}

	user.UpdatedAt = u.UpdatedAt.Time

	return user, nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, id uint64) error {
	return r.queries.DeleteUser(ctx, int64(id))
}

func (r *PostgresRepository) VerifyUserEmail(ctx context.Context, id uint64) (*User, error) {
	u, err := r.queries.VerifyUserEmail(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return userFromRow(u), nil
}

func (r *PostgresRepository) CreateAgreement(ctx context.Context, agreement *Agreement) (*Agreement, error) {

	a, err := r.queries.CreateAgreement(ctx, db.CreateAgreementParams{
		Type:        string(agreement.Type),
		Version:     agreement.Version,
		Content:     agreement.Content,
		PublishedAt: pgutil.ToPgTimestamptz(&agreement.PublishedAt),
	})
	if err != nil {
		return nil, err
	}

	agreement.ID = uint64(a.ID)

	return agreement, nil
}

func (r *PostgresRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}

func (r *PostgresRepository) GetAgreementByID(ctx context.Context, id uint64) (*Agreement, error) {
	a, err := r.queries.GetAgreementByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return agreementFromRow(a), nil
}

func (r *PostgresRepository) GetAgreementByTypeAndVersion(ctx context.Context, agType AgreementType, version string) (*Agreement, error) {
	a, err := r.queries.GetAgreementByTypeAndVersion(ctx, db.GetAgreementByTypeAndVersionParams{
		Type:    string(agType),
		Version: version,
	})
	if err != nil {
		return nil, err
	}
	return agreementFromRow(a), nil
}

func (r *PostgresRepository) GetLatestAgreements(ctx context.Context) ([]*Agreement, error) {
	agreements, err := r.queries.GetLatestAgreements(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*Agreement, len(agreements))
	for i, a := range agreements {
		result[i] = agreementFromRow(a)
	}
	return result, nil
}

func (r *PostgresRepository) CreateUserAgreement(ctx context.Context, userAgreement *UserAgreement) (*UserAgreement, error) {

	ua, err := r.queries.CreateUserAgreement(ctx, db.CreateUserAgreementParams{
		UserID:      int64(userAgreement.UserID),
		AgreementID: int64(userAgreement.AgreementID),
	})
	if err != nil {
		return nil, err
	}

	userAgreement.ID = uint64(ua.ID)
	userAgreement.AcceptedAt = ua.CreatedAt.Time

	return userAgreement, nil
}

func (r *PostgresRepository) HasUserAcceptedAgreement(ctx context.Context, userID uint64, agreementID uint64) (bool, error) {
	return r.queries.HasUserAcceptedAgreement(ctx, db.HasUserAcceptedAgreementParams{
		UserID:      int64(userID),
		AgreementID: int64(agreementID),
	})
}

func (r *PostgresRepository) GetPendingAgreementsForUser(ctx context.Context, userID uint64) ([]*Agreement, error) {
	agreements, err := r.queries.GetPendingAgreementsForUser(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	result := make([]*Agreement, len(agreements))
	for i, a := range agreements {
		result[i] = agreementFromRow(a)
	}
	return result, nil
}

func userFromRow(u db.User) *User {
	return &User{
		ID:                     uint64(u.ID),
		Email:                  u.Email,
		PasswordHash:           u.PasswordHash,
		Role:                   Role(u.Role),
		Status:                 AccountStatus(u.Status),
		Metadata:               u.Metadata,
		IsVerified:             u.IsVerified,
		VerificationToken:      pgutil.FromPgText(u.VerificationToken),
		VerificationExpiresAt:  pgutil.FromPgTimestamptz(u.VerificationExpiresAt),
		PasswordResetToken:     pgutil.FromPgText(u.PasswordResetToken),
		PasswordResetExpiresAt: pgutil.FromPgTimestamptz(u.PasswordResetExpiresAt),
		CreatedAt:              u.CreatedAt.Time,
		UpdatedAt:              u.UpdatedAt.Time,
		DeletedAt:              pgutil.FromPgTimestamptz(u.DeletedAt),
	}
}

func agreementFromRow(a db.Agreement) *Agreement {
	return &Agreement{
		ID:          uint64(a.ID),
		Type:        AgreementType(a.Type),
		Version:     a.Version,
		Content:     a.Content,
		PublishedAt: a.PublishedAt.Time,
	}
}
