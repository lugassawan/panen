package crashplaybook

import "context"

// CrashCapitalRepository defines persistence operations for crash capital.
type CrashCapitalRepository interface {
	Upsert(ctx context.Context, cc *CrashCapital) error
	GetByPortfolioID(ctx context.Context, portfolioID string) (*CrashCapital, error)
}
