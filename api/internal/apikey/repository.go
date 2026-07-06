package apikey

import "context"

type Repository interface {
	Create(ctx context.Context, apiKey *APIKey) (*APIKey, error)
	GetByID(ctx context.Context, id uint64) (*APIKey, error)
	GetBySelector(ctx context.Context, selector string) (*APIKey, error)
	ListByUserID(ctx context.Context, userID uint64) ([]*APIKey, error)
	Update(ctx context.Context, apiKey *APIKey) (*APIKey, error)
	Delete(ctx context.Context, id uint64) error
	IsUserActive(ctx context.Context, userID uint64) (bool, error)
}
