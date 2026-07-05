package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/domain"
)

type UserAgreementRepository struct {
	db *pgxpool.Pool
}

func NewUserAgreementRepository(db *pgxpool.Pool) *UserAgreementRepository {
	return &UserAgreementRepository{db: db}
}

func (r *UserAgreementRepository) Create(ctx context.Context, userAgreement *domain.UserAgreement) (*domain.UserAgreement, error) {
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

func (r *UserAgreementRepository) HasAccepted(ctx context.Context, userID uint64, agreementID uint64) (bool, error) {
	const sql = `
		SELECT EXISTS (
			SELECT 1 FROM user_agreements WHERE user_id = $1 AND agreement_id = $2
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, sql, userID, agreementID).Scan(&exists)
	return exists, err
}
