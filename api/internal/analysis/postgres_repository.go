package analysis

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/analysis/db"
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

func (r *PostgresRepository) CreateRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error) {
	ar, err := r.queries.CreateAnalysisRequest(ctx, db.CreateAnalysisRequestParams{
		RequestID:   req.RequestID,
		UserID:      int64(req.UserID),
		ApiKeyID:    int64(req.APIKeyID),
		Status:      string(req.Status),
		Error:       toPgText(req.Error),
		Metadata:    req.Metadata,
		StartedAt:   toPgTimestamp(req.StartedAt),
		CompletedAt: toPgTimestamp(req.CompletedAt),
	})
	if err != nil {
		return nil, err
	}
	req.ID = uint64(ar.ID)
	req.CreatedAt = ar.CreatedAt.Time
	req.UpdatedAt = ar.UpdatedAt.Time
	return req, nil
}

func (r *PostgresRepository) GetRequestByID(ctx context.Context, id uint64) (*AnalysisRequest, error) {
	ar, err := r.queries.GetAnalysisRequestByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return mapDBAnalysisRequestToModel(ar), nil
}

func (r *PostgresRepository) ListRequestsByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*AnalysisRequest, error) {
	dbReqs, err := r.queries.ListAnalysisRequestsByUserID(ctx, db.ListAnalysisRequestsByUserIDParams{
		UserID: int64(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	reqs := make([]*AnalysisRequest, len(dbReqs))
	for i, ar := range dbReqs {
		reqs[i] = mapDBAnalysisRequestToModel(ar)
	}
	return reqs, nil
}

func (r *PostgresRepository) UpdateRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error) {
	ar, err := r.queries.UpdateAnalysisRequest(ctx, db.UpdateAnalysisRequestParams{
		ID:          int64(req.ID),
		Status:      string(req.Status),
		Error:       toPgText(req.Error),
		Metadata:    req.Metadata,
		StartedAt:   toPgTimestamp(req.StartedAt),
		CompletedAt: toPgTimestamp(req.CompletedAt),
	})
	if err != nil {
		return nil, err
	}
	req.UpdatedAt = ar.UpdatedAt.Time
	return req, nil
}

func (r *PostgresRepository) DeleteRequest(ctx context.Context, id uint64) error {
	return r.queries.DeleteAnalysisRequest(ctx, int64(id))
}

func (r *PostgresRepository) CreateResult(ctx context.Context, result *AnalysisResult) (*AnalysisResult, error) {
	res, err := r.queries.CreateAnalysisResult(ctx, db.CreateAnalysisResultParams{
		AnalysisRequestID: int64(result.AnalysisRequestID),
		Model:             result.Model,
		Result:            result.Result,
		PromptTokens:      int32(result.PromptTokens),
		CompletionTokens:  int32(result.CompletionTokens),
		TotalTokens:       int32(result.TotalTokens),
	})
	if err != nil {
		return nil, err
	}
	result.ID = uint64(res.ID)
	result.CreatedAt = res.CreatedAt.Time
	result.UpdatedAt = res.UpdatedAt.Time
	return result, nil
}

func (r *PostgresRepository) GetResultByRequestID(ctx context.Context, requestID uint64) (*AnalysisResult, error) {
	res, err := r.queries.GetAnalysisResultByRequestID(ctx, int64(requestID))
	if err != nil {
		return nil, err
	}

	var deletedAt *time.Time
	if res.DeletedAt.Valid {
		deletedAt = &res.DeletedAt.Time
	}

	return &AnalysisResult{
		ID:                uint64(res.ID),
		AnalysisRequestID: uint64(res.AnalysisRequestID),
		Model:             res.Model,
		Result:            res.Result,
		PromptTokens:      uint32(res.PromptTokens),
		CompletionTokens:  uint32(res.CompletionTokens),
		TotalTokens:       uint32(res.TotalTokens),
		CreatedAt:         res.CreatedAt.Time,
		UpdatedAt:         res.UpdatedAt.Time,
		DeletedAt:         deletedAt,
	}, nil
}

func (r *PostgresRepository) DeleteResult(ctx context.Context, id uint64) error {
	return r.queries.DeleteAnalysisResult(ctx, int64(id))
}

func mapDBAnalysisRequestToModel(ar db.AnalysisRequest) *AnalysisRequest {
	var errorStr *string
	if ar.Error.Valid {
		errorStr = &ar.Error.String
	}
	var startedAt *time.Time
	if ar.StartedAt.Valid {
		startedAt = &ar.StartedAt.Time
	}
	var completedAt *time.Time
	if ar.CompletedAt.Valid {
		completedAt = &ar.CompletedAt.Time
	}
	var deletedAt *time.Time
	if ar.DeletedAt.Valid {
		deletedAt = &ar.DeletedAt.Time
	}

	return &AnalysisRequest{
		ID:          uint64(ar.ID),
		RequestID:   ar.RequestID,
		UserID:      uint64(ar.UserID),
		APIKeyID:    uint64(ar.ApiKeyID),
		Status:      AnalysisRequestStatus(ar.Status),
		Error:       errorStr,
		Metadata:    ar.Metadata,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		CreatedAt:   ar.CreatedAt.Time,
		UpdatedAt:   ar.UpdatedAt.Time,
		DeletedAt:   deletedAt,
	}
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toPgTimestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
