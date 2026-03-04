package checklist

import "context"

// Repository defines persistence operations for checklist results.
type Repository interface {
	Upsert(ctx context.Context, r *ChecklistResult) error
	Get(ctx context.Context, portfolioID, ticker string, action ActionType) (*ChecklistResult, error)
	Delete(ctx context.Context, id string) error
	DeleteByPortfolioID(ctx context.Context, portfolioID string) error
}
