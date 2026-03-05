package payday

import "context"

// Repository defines persistence operations for payday events.
type Repository interface {
	Create(ctx context.Context, event *PaydayEvent) error
	GetByMonthAndPortfolio(ctx context.Context, month, portfolioID string) (*PaydayEvent, error)
	ListByMonth(ctx context.Context, month string) ([]*PaydayEvent, error)
	ListByPortfolioID(ctx context.Context, portfolioID string) ([]*PaydayEvent, error)
	Update(ctx context.Context, event *PaydayEvent) error
}

// CashFlowRepository defines persistence operations for cash flow records.
type CashFlowRepository interface {
	Create(ctx context.Context, cf *CashFlow) error
	ListByPortfolioID(ctx context.Context, portfolioID string) ([]*CashFlow, error)
	Delete(ctx context.Context, id string) error
}
