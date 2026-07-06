package analysis

import "context"

type Repository interface {
	CreateRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error)
	GetRequestByID(ctx context.Context, id uint64) (*AnalysisRequest, error)
	ListRequestsByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*AnalysisRequest, error)
	UpdateRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisRequest, error)
	DeleteRequest(ctx context.Context, id uint64) error

	CreateResult(ctx context.Context, result *AnalysisResult) (*AnalysisResult, error)
	GetResultByRequestID(ctx context.Context, requestID uint64) (*AnalysisResult, error)
	DeleteResult(ctx context.Context, id uint64) error
}
