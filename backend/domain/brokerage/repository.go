package brokerage

import "context"

// Repository defines persistence operations for brokerage accounts.
type Repository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	ListByProfileID(ctx context.Context, profileID string) ([]*Account, error)
	ListNonManualByProfileID(ctx context.Context, profileID string) ([]*Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id string) error
}
