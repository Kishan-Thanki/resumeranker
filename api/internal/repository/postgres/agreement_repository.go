package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/domain"
)

type AgreementRepository struct {
	db *pgxpool.Pool
}

func NewAgreementRepository(db *pgxpool.Pool) *AgreementRepository {
	return &AgreementRepository{db: db}
}

func (r *AgreementRepository) Create(ctx context.Context, agreement *domain.Agreement) (*domain.Agreement, error) {
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

func (r *AgreementRepository) GetByID(ctx context.Context, id uint64) (*domain.Agreement, error) {
	const sql = `
		SELECT id, type, version, document_url, published_at, created_at, updated_at
		FROM agreements
		WHERE id = $1
	`
	agreement := &domain.Agreement{}
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

func (r *AgreementRepository) GetByTypeAndVersion(ctx context.Context, agType domain.AgreementType, version string) (*domain.Agreement, error) {
	const sql = `
		SELECT id, type, version, document_url, published_at, created_at, updated_at
		FROM agreements
		WHERE type = $1 AND version = $2
	`
	agreement := &domain.Agreement{}
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
