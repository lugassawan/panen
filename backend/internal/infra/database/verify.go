package database

import (
	"github.com/lugassawan/panen/backend/internal/domain/brokerage"
	"github.com/lugassawan/panen/backend/internal/domain/portfolio"
	"github.com/lugassawan/panen/backend/internal/domain/stock"
	"github.com/lugassawan/panen/backend/internal/domain/user"
)

// Compile-time interface compliance checks.
var (
	_ user.Repository                    = (*UserRepo)(nil)
	_ brokerage.Repository               = (*BrokerageRepo)(nil)
	_ portfolio.Repository               = (*PortfolioRepo)(nil)
	_ portfolio.HoldingRepository        = (*HoldingRepo)(nil)
	_ portfolio.BuyTransactionRepository = (*BuyTransactionRepo)(nil)
	_ stock.Repository                   = (*StockDataRepo)(nil)
)
