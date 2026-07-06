package users

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
	const sql = `
		INSERT INTO users (email, password_hash, role, status, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.Metadata,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uint64) (*User, error) {
	const sql = `
		SELECT id, email, password_hash, role, status, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	user := &User{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.Metadata,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const sql = `
		SELECT id, email, password_hash, role, status, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	user := &User{}
	err := r.db.QueryRow(ctx, sql, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.Metadata,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) (*User, error) {
	const sql = `
		UPDATE users
		SET email = $1, password_hash = $2, role = $3, status = $4, metadata = $5, updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.Metadata,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, id uint64) error {
	const sql = `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *PostgresRepository) CreateAgreement(ctx context.Context, agreement *Agreement) (*Agreement, error) {
	const sql = `
		INSERT INTO agreements (type, version, document_url, published_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		agreement.Type,
		agreement.Version,
		agreement.DocumentURL,
		agreement.PublishedAt,
	).Scan(
		&agreement.ID,
		&agreement.CreatedAt,
		&agreement.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return agreement, nil
}

func (r *PostgresRepository) GetAgreementByID(ctx context.Context, id uint64) (*Agreement, error) {
	const sql = `
		SELECT id, type, version, document_url, published_at, created_at, updated_at
		FROM agreements
		WHERE id = $1
	`
	agreement := &Agreement{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&agreement.ID,
		&agreement.Type,
		&agreement.Version,
		&agreement.DocumentURL,
		&agreement.PublishedAt,
		&agreement.CreatedAt,
		&agreement.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return agreement, nil
}

func (r *PostgresRepository) GetAgreementByTypeAndVersion(ctx context.Context, agType AgreementType, version string) (*Agreement, error) {
	const sql = `
		SELECT id, type, version, document_url, published_at, created_at, updated_at
		FROM agreements
		WHERE type = $1 AND version = $2
	`
	agreement := &Agreement{}
	err := r.db.QueryRow(ctx, sql, agType, version).Scan(
		&agreement.ID,
		&agreement.Type,
		&agreement.Version,
		&agreement.DocumentURL,
		&agreement.PublishedAt,
		&agreement.CreatedAt,
		&agreement.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return agreement, nil
}

func (r *PostgresRepository) CreateUserAgreement(ctx context.Context, userAgreement *UserAgreement) (*UserAgreement, error) {
	const sql = `
		INSERT INTO user_agreements (user_id, agreement_id, accepted_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		userAgreement.UserID,
		userAgreement.AgreementID,
		userAgreement.AcceptedAt,
	).Scan(
		&userAgreement.ID,
		&userAgreement.CreatedAt,
		&userAgreement.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return userAgreement, nil
}

func (r *PostgresRepository) HasUserAcceptedAgreement(ctx context.Context, userID uint64, agreementID uint64) (bool, error) {
	const sql = `
		SELECT EXISTS (
			SELECT 1 FROM user_agreements WHERE user_id = $1 AND agreement_id = $2
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, sql, userID, agreementID).Scan(&exists)
	return exists, err
}
