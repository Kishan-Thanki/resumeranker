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
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
		Status:       string(user.Status),
		Metadata:     user.Metadata,
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

	return &User{
		ID:           uint64(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         Role(u.Role),
		Status:       AccountStatus(u.Status),
		Metadata:     u.Metadata,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
		DeletedAt:    deletedAt,
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

	return &User{
		ID:           uint64(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         Role(u.Role),
		Status:       AccountStatus(u.Status),
		Metadata:     u.Metadata,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
		DeletedAt:    deletedAt,
	}, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) (*User, error) {
	u, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:           int64(user.ID),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
		Status:       string(user.Status),
		Metadata:     user.Metadata,
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

func (r *PostgresRepository) CreateAgreement(ctx context.Context, agreement *Agreement) (*Agreement, error) {
	a, err := r.queries.CreateAgreement(ctx, db.CreateAgreementParams{
		Type:        string(agreement.Type),
		Version:     agreement.Version,
		DocumentUrl: agreement.DocumentURL,
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
		DocumentURL: a.DocumentUrl,
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
		DocumentURL: a.DocumentUrl,
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

func toPgTimestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
