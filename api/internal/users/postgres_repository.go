package users

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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
		VerificationToken:      toPgText(user.VerificationToken),
		VerificationExpiresAt:  toPgTimestamp(user.VerificationExpiresAt),
		PasswordResetToken:     toPgText(user.PasswordResetToken),
		PasswordResetExpiresAt: toPgTimestamp(user.PasswordResetExpiresAt),
	})
	if err != nil {
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

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}

	var verifyExpiresAt *time.Time
	if u.VerificationExpiresAt.Valid {
		verifyExpiresAt = &u.VerificationExpiresAt.Time
	}

	var resetExpiresAt *time.Time
	if u.PasswordResetExpiresAt.Valid {
		resetExpiresAt = &u.PasswordResetExpiresAt.Time
	}

	return &User{
		ID:                     uint64(u.ID),
		Email:                  u.Email,
		PasswordHash:           u.PasswordHash,
		Role:                   Role(u.Role),
		Status:                 AccountStatus(u.Status),
		Metadata:               u.Metadata,
		IsVerified:             u.IsVerified,
		VerificationToken:      fromPgText(u.VerificationToken),
		VerificationExpiresAt:  verifyExpiresAt,
		PasswordResetToken:     fromPgText(u.PasswordResetToken),
		PasswordResetExpiresAt: resetExpiresAt,
		CreatedAt:              u.CreatedAt.Time,
		UpdatedAt:              u.UpdatedAt.Time,
		DeletedAt:              deletedAt,
	}, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {

	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}

	var verifyExpiresAt *time.Time
	if u.VerificationExpiresAt.Valid {
		verifyExpiresAt = &u.VerificationExpiresAt.Time
	}

	var resetExpiresAt *time.Time
	if u.PasswordResetExpiresAt.Valid {
		resetExpiresAt = &u.PasswordResetExpiresAt.Time
	}

	return &User{
		ID:                     uint64(u.ID),
		Email:                  u.Email,
		PasswordHash:           u.PasswordHash,
		Role:                   Role(u.Role),
		Status:                 AccountStatus(u.Status),
		Metadata:               u.Metadata,
		IsVerified:             u.IsVerified,
		VerificationToken:      fromPgText(u.VerificationToken),
		VerificationExpiresAt:  verifyExpiresAt,
		PasswordResetToken:     fromPgText(u.PasswordResetToken),
		PasswordResetExpiresAt: resetExpiresAt,
		CreatedAt:              u.CreatedAt.Time,
		UpdatedAt:              u.UpdatedAt.Time,
		DeletedAt:              deletedAt,
	}, nil
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
		var deletedAt *time.Time
		if u.DeletedAt.Valid {
			deletedAt = &u.DeletedAt.Time
		}

		var verifyExpiresAt *time.Time
		if u.VerificationExpiresAt.Valid {
			verifyExpiresAt = &u.VerificationExpiresAt.Time
		}

		var resetExpiresAt *time.Time
		if u.PasswordResetExpiresAt.Valid {
			resetExpiresAt = &u.PasswordResetExpiresAt.Time
		}

		result[i] = &User{
			ID:                     uint64(u.ID),
			Email:                  u.Email,
			PasswordHash:           u.PasswordHash,
			Role:                   Role(u.Role),
			Status:                 AccountStatus(u.Status),
			Metadata:               u.Metadata,
			IsVerified:             u.IsVerified,
			VerificationToken:      fromPgText(u.VerificationToken),
			VerificationExpiresAt:  verifyExpiresAt,
			PasswordResetToken:     fromPgText(u.PasswordResetToken),
			PasswordResetExpiresAt: resetExpiresAt,
			CreatedAt:              u.CreatedAt.Time,
			UpdatedAt:              u.UpdatedAt.Time,
			DeletedAt:              deletedAt,
		}
	}
	return result, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) (*User, error) {

	u, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:                     int64(user.ID),
		Email:                  user.Email,
		PasswordHash:           user.PasswordHash,
		Role:                   string(user.Role),
		Status:                 string(user.Status),
		Metadata:               user.Metadata,
		IsVerified:             user.IsVerified,
		VerificationToken:      toPgText(user.VerificationToken),
		VerificationExpiresAt:  toPgTimestamp(user.VerificationExpiresAt),
		PasswordResetToken:     toPgText(user.PasswordResetToken),
		PasswordResetExpiresAt: toPgTimestamp(user.PasswordResetExpiresAt),
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

func (r *PostgresRepository) GetUserByVerificationToken(ctx context.Context, token string) (*User, error) {

	pgToken := toPgText(&token)
	u, err := r.queries.GetUserByVerificationToken(ctx, pgToken)
	if err != nil {
		return nil, err
	}

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}

	var verifyExpiresAt *time.Time
	if u.VerificationExpiresAt.Valid {
		verifyExpiresAt = &u.VerificationExpiresAt.Time
	}

	var resetExpiresAt *time.Time
	if u.PasswordResetExpiresAt.Valid {
		resetExpiresAt = &u.PasswordResetExpiresAt.Time
	}

	return &User{
		ID:                     uint64(u.ID),
		Email:                  u.Email,
		PasswordHash:           u.PasswordHash,
		Role:                   Role(u.Role),
		Status:                 AccountStatus(u.Status),
		Metadata:               u.Metadata,
		IsVerified:             u.IsVerified,
		VerificationToken:      fromPgText(u.VerificationToken),
		VerificationExpiresAt:  verifyExpiresAt,
		PasswordResetToken:     fromPgText(u.PasswordResetToken),
		PasswordResetExpiresAt: resetExpiresAt,
		CreatedAt:              u.CreatedAt.Time,
		UpdatedAt:              u.UpdatedAt.Time,
		DeletedAt:              deletedAt,
	}, nil
}

func (r *PostgresRepository) VerifyUserEmail(ctx context.Context, id uint64) (*User, error) {

	u, err := r.queries.VerifyUserEmail(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}

	var verifyExpiresAt *time.Time
	if u.VerificationExpiresAt.Valid {
		verifyExpiresAt = &u.VerificationExpiresAt.Time
	}

	var resetExpiresAt *time.Time
	if u.PasswordResetExpiresAt.Valid {
		resetExpiresAt = &u.PasswordResetExpiresAt.Time
	}

	return &User{
		ID:                     uint64(u.ID),
		Email:                  u.Email,
		PasswordHash:           u.PasswordHash,
		Role:                   Role(u.Role),
		Status:                 AccountStatus(u.Status),
		Metadata:               u.Metadata,
		IsVerified:             u.IsVerified,
		VerificationToken:      fromPgText(u.VerificationToken),
		VerificationExpiresAt:  verifyExpiresAt,
		PasswordResetToken:     fromPgText(u.PasswordResetToken),
		PasswordResetExpiresAt: resetExpiresAt,
		CreatedAt:              u.CreatedAt.Time,
		UpdatedAt:              u.UpdatedAt.Time,
		DeletedAt:              deletedAt,
	}, nil
}

func (r *PostgresRepository) GetUserByPasswordResetToken(ctx context.Context, token string) (*User, error) {

	pgToken := toPgText(&token)
	u, err := r.queries.GetUserByPasswordResetToken(ctx, pgToken)
	if err != nil {
		return nil, err
	}

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}

	var verifyExpiresAt *time.Time
	if u.VerificationExpiresAt.Valid {
		verifyExpiresAt = &u.VerificationExpiresAt.Time
	}

	var resetExpiresAt *time.Time
	if u.PasswordResetExpiresAt.Valid {
		resetExpiresAt = &u.PasswordResetExpiresAt.Time
	}

	return &User{
		ID:                     uint64(u.ID),
		Email:                  u.Email,
		PasswordHash:           u.PasswordHash,
		Role:                   Role(u.Role),
		Status:                 AccountStatus(u.Status),
		Metadata:               u.Metadata,
		IsVerified:             u.IsVerified,
		VerificationToken:      fromPgText(u.VerificationToken),
		VerificationExpiresAt:  verifyExpiresAt,
		PasswordResetToken:     fromPgText(u.PasswordResetToken),
		PasswordResetExpiresAt: resetExpiresAt,
		CreatedAt:              u.CreatedAt.Time,
		UpdatedAt:              u.UpdatedAt.Time,
		DeletedAt:              deletedAt,
	}, nil
}

func (r *PostgresRepository) CreateAgreement(ctx context.Context, agreement *Agreement) (*Agreement, error) {

	a, err := r.queries.CreateAgreement(ctx, db.CreateAgreementParams{
		Type:        string(agreement.Type),
		Version:     agreement.Version,
		Content:     agreement.Content,
		PublishedAt: toPgTimestamp(&agreement.PublishedAt),
	})
	if err != nil {
		return nil, err
	}

	agreement.ID = uint64(a.ID)
	agreement.CreatedAt = a.CreatedAt.Time
	agreement.UpdatedAt = a.UpdatedAt.Time

	return agreement, nil
}

func (r *PostgresRepository) GetAgreementByID(ctx context.Context, id uint64) (*Agreement, error) {

	a, err := r.queries.GetAgreementByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	return &Agreement{
		ID:          uint64(a.ID),
		Type:        AgreementType(a.Type),
		Version:     a.Version,
		Content:     a.Content,
		PublishedAt: a.PublishedAt.Time,
		CreatedAt:   a.CreatedAt.Time,
		UpdatedAt:   a.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) GetAgreementByTypeAndVersion(ctx context.Context, agType AgreementType, version string) (*Agreement, error) {

	a, err := r.queries.GetAgreementByTypeAndVersion(ctx, db.GetAgreementByTypeAndVersionParams{
		Type:    string(agType),
		Version: version,
	})
	if err != nil {
		return nil, err
	}

	return &Agreement{
		ID:          uint64(a.ID),
		Type:        AgreementType(a.Type),
		Version:     a.Version,
		Content:     a.Content,
		PublishedAt: a.PublishedAt.Time,
		CreatedAt:   a.CreatedAt.Time,
		UpdatedAt:   a.UpdatedAt.Time,
	}, nil
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
	userAgreement.CreatedAt = ua.CreatedAt.Time
	userAgreement.UpdatedAt = ua.UpdatedAt.Time

	return userAgreement, nil
}

func (r *PostgresRepository) HasUserAcceptedAgreement(ctx context.Context, userID uint64, agreementID uint64) (bool, error) {
	return r.queries.HasUserAcceptedAgreement(ctx, db.HasUserAcceptedAgreementParams{
		UserID:      int64(userID),
		AgreementID: int64(agreementID),
	})
}

func (r *PostgresRepository) GetLatestAgreements(ctx context.Context) ([]*Agreement, error) {
	agreements, err := r.queries.GetLatestAgreements(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*Agreement, len(agreements))
	for i, a := range agreements {
		result[i] = &Agreement{
			ID:          uint64(a.ID),
			Type:        AgreementType(a.Type),
			Version:     a.Version,
			Content:     a.Content,
			PublishedAt: a.PublishedAt.Time,
			CreatedAt:   a.CreatedAt.Time,
			UpdatedAt:   a.UpdatedAt.Time,
		}
	}
	return result, nil
}

func (r *PostgresRepository) GetPendingAgreementsForUser(ctx context.Context, userID uint64) ([]*Agreement, error) {
	agreements, err := r.queries.GetPendingAgreementsForUser(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	result := make([]*Agreement, len(agreements))
	for i, a := range agreements {
		result[i] = &Agreement{
			ID:          uint64(a.ID),
			Type:        AgreementType(a.Type),
			Version:     a.Version,
			Content:     a.Content,
			PublishedAt: a.PublishedAt.Time,
			CreatedAt:   a.CreatedAt.Time,
			UpdatedAt:   a.UpdatedAt.Time,
		}
	}
	return result, nil
}

func toPgTimestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func fromPgText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
