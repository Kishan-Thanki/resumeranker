package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/domain"
)

type AnalysisResultRepository struct {
	db *pgxpool.Pool
}

func NewAnalysisResultRepository(db *pgxpool.Pool) *AnalysisResultRepository {
	return &AnalysisResultRepository{db: db}
}

func (r *AnalysisResultRepository) Create(ctx context.Context, result *domain.AnalysisResult) (*domain.AnalysisResult, error) {
	const sql = `
		INSERT INTO analysis_results (analysis_request_id, model, result, prompt_tokens, completion_tokens, total_tokens)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, sql,
		result.AnalysisRequestID,
		result.Model,
		result.Result,
		result.PromptTokens,
		result.CompletionTokens,
		result.TotalTokens,
	).Scan(
		&result.ID,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *AnalysisResultRepository) GetByRequestID(ctx context.Context, requestID uint64) (*domain.AnalysisResult, error) {
	const sql = `
		SELECT id, analysis_request_id, model, result, prompt_tokens, completion_tokens, total_tokens, created_at, updated_at, deleted_at
		FROM analysis_results
		WHERE analysis_request_id = $1 AND deleted_at IS NULL
	`
	result := &domain.AnalysisResult{}
	err := r.db.QueryRow(ctx, sql, requestID).Scan(
		&result.ID,
		&result.AnalysisRequestID,
		&result.Model,
		&result.Result,
		&result.PromptTokens,
		&result.CompletionTokens,
		&result.TotalTokens,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *AnalysisResultRepository) Delete(ctx context.Context, id uint64) error {
	const sql = `
		UPDATE analysis_results
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}
