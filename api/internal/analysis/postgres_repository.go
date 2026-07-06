package analysis

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

func (r *PostgresRepository) CreateRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error) {
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

func (r *PostgresRepository) GetRequestByID(ctx context.Context, id uint64) (*AnalysisRequest, error) {
	const sql = `
		SELECT id, request_id, user_id, api_key_id, status, error, metadata, started_at, completed_at, created_at, updated_at, deleted_at
		FROM analysis_requests
		WHERE id = $1 AND deleted_at IS NULL
	`
	req := &AnalysisRequest{}
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

func (r *PostgresRepository) ListRequestsByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*AnalysisRequest, error) {
	const sql = `
		SELECT id, request_id, user_id, api_key_id, status, error, metadata, started_at, completed_at, created_at, updated_at, deleted_at
		FROM analysis_requests
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, sql, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*AnalysisRequest
	for rows.Next() {
		req := &AnalysisRequest{}
		err := rows.Scan(
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
		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *PostgresRepository) UpdateRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error) {
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

func (r *PostgresRepository) DeleteRequest(ctx context.Context, id uint64) error {
	const sql = `
		UPDATE analysis_requests
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *PostgresRepository) CreateResult(ctx context.Context, result *AnalysisResult) (*AnalysisResult, error) {
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

func (r *PostgresRepository) GetResultByRequestID(ctx context.Context, requestID uint64) (*AnalysisResult, error) {
	const sql = `
		SELECT id, analysis_request_id, model, result, prompt_tokens, completion_tokens, total_tokens, created_at, updated_at, deleted_at
		FROM analysis_results
		WHERE analysis_request_id = $1 AND deleted_at IS NULL
	`
	result := &AnalysisResult{}
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

func (r *PostgresRepository) DeleteResult(ctx context.Context, id uint64) error {
	const sql = `
		UPDATE analysis_results
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}
