package database

import (
	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/user"
	"github.com/lugassawan/panen/backend/domain/watchlist"
)

// Compile-time interface compliance checks.
var (
	_ user.Repository                    = (*UserRepo)(nil)
	_ brokerage.Repository               = (*BrokerageRepo)(nil)
	_ portfolio.Repository               = (*PortfolioRepo)(nil)
	_ portfolio.HoldingRepository        = (*HoldingRepo)(nil)
	_ portfolio.BuyTransactionRepository = (*BuyTransactionRepo)(nil)
	_ stock.Repository                   = (*StockDataRepo)(nil)
	_ watchlist.Repository               = (*WatchlistRepo)(nil)
	_ watchlist.ItemRepository           = (*WatchlistItemRepo)(nil)
	_ checklist.Repository               = (*ChecklistResultRepo)(nil)
)
