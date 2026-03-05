package portfolio

import "context"

// Repository defines persistence operations for portfolios.
type Repository interface {
	Create(ctx context.Context, p *Portfolio) error
	GetByID(ctx context.Context, id string) (*Portfolio, error)
	ListAll(ctx context.Context) ([]*Portfolio, error)
	ListByBrokerageAccountID(ctx context.Context, brokerageAccountID string) ([]*Portfolio, error)
	Update(ctx context.Context, p *Portfolio) error
	Delete(ctx context.Context, id string) error
}

// HoldingRepository defines persistence operations for stock holdings.
type HoldingRepository interface {
	Create(ctx context.Context, holding *Holding) error
	GetByID(ctx context.Context, id string) (*Holding, error)
	GetByPortfolioAndTicker(ctx context.Context, portfolioID, ticker string) (*Holding, error)
	ListByPortfolioID(ctx context.Context, portfolioID string) ([]*Holding, error)
	Update(ctx context.Context, holding *Holding) error
	Delete(ctx context.Context, id string) error
}

// BuyTransactionRepository defines persistence operations for buy transactions.
// Transactions are immutable — no Update method.
type BuyTransactionRepository interface {
	Create(ctx context.Context, tx *BuyTransaction) error
	GetByID(ctx context.Context, id string) (*BuyTransaction, error)
	ListByHoldingID(ctx context.Context, holdingID string) ([]*BuyTransaction, error)
	Delete(ctx context.Context, id string) error
}
