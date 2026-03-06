package alert

import (
	"context"
	"time"
)

// Repository defines persistence operations for fundamental alerts.
type Repository interface {
	Create(ctx context.Context, alert *FundamentalAlert) error
	GetByTicker(ctx context.Context, ticker string) ([]*FundamentalAlert, error)
	GetActive(ctx context.Context) ([]*FundamentalAlert, error)
	GetActiveByTicker(ctx context.Context, ticker string) ([]*FundamentalAlert, error)
	Acknowledge(ctx context.Context, id string) error
	Resolve(ctx context.Context, id string) error
	CountActive(ctx context.Context) (int, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
