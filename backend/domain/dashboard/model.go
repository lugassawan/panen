package dashboard

import "github.com/lugassawan/panen/backend/domain/transaction"

// Overview holds aggregated performance data across all user portfolios.
type Overview struct {
	TotalMarketValue    float64
	TotalCostBasis      float64
	TotalPLAmount       float64
	TotalPLPercent      float64
	TotalDividendIncome float64
	Portfolios          []PortfolioSummary
	TopGainers          []HoldingPL
	TopLosers           []HoldingPL
	PortfolioAllocation []AllocationItem
	SectorAllocation    []AllocationItem
	RecentTransactions  []transaction.Record
	WinRate             float64
	HoldingCount        int
	WinningCount        int
}

// PortfolioSummary holds per-portfolio aggregated values.
type PortfolioSummary struct {
	ID, Name, Mode         string
	MarketValue, CostBasis float64
	PLAmount, PLPercent    float64
	Weight                 float64
}

// HoldingPL holds profit/loss data for a single holding.
type HoldingPL struct {
	Ticker, PortfolioID, PortfolioName string
	MarketValue, CostBasis             float64
	PLAmount, PLPercent                float64
}

// AllocationItem holds a single slice of an allocation breakdown.
type AllocationItem struct {
	Label string
	Value float64
	Pct   float64
}
