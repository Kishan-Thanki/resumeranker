package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/domain"
)

type AnalysisRequestRepository struct {
	db *pgxpool.Pool
}

func NewAnalysisRequestRepository(db *pgxpool.Pool) *AnalysisRequestRepository {
	return &AnalysisRequestRepository{db: db}
}

func (r *AnalysisRequestRepository) Create(ctx context.Context, req *domain.AnalysisRequest) (*domain.AnalysisRequest, error) {
	const sql = `
		INSERT INTO analysis_requests (request_id, user_id, api_key_id, status, error, metadata, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		req.RequestID,
		req.UserID,
		req.APIKeyID,
		req.Status,
		req.Error,
		req.Metadata,
		req.StartedAt,
		req.CompletedAt,
	).Scan(
		&req.ID,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *AnalysisRequestRepository) GetByID(ctx context.Context, id uint64) (*domain.AnalysisRequest, error) {
	const sql = `
		SELECT id, request_id, user_id, api_key_id, status, error, metadata, started_at, completed_at, created_at, updated_at, deleted_at
		FROM analysis_requests
		WHERE id = $1 AND deleted_at IS NULL
	`
	req := &domain.AnalysisRequest{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&req.ID,
		&req.RequestID,
		&req.UserID,
		&req.APIKeyID,
		&req.Status,
		&req.Error,
		&req.Metadata,
		&req.StartedAt,
		&req.CompletedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
		&req.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *AnalysisRequestRepository) Update(ctx context.Context, req *domain.AnalysisRequest) (*domain.AnalysisRequest, error) {
	const sql = `
		UPDATE analysis_requests
		SET status = $1, error = $2, metadata = $3, started_at = $4, completed_at = $5, updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		req.Status,
		req.Error,
		req.Metadata,
		req.StartedAt,
		req.CompletedAt,
		req.ID,
	).Scan(&req.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *AnalysisRequestRepository) Delete(ctx context.Context, id uint64) error {
	const sql = `
		UPDATE analysis_requests
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}
