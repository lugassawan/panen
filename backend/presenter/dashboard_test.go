package presenter

import (
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/dashboard"
	"github.com/lugassawan/panen/backend/domain/transaction"
)

func TestNewDashboardOverviewResponse(t *testing.T) {
	overview := &dashboard.Overview{
		TotalMarketValue:    16000000,
		TotalCostBasis:      14000000,
		TotalPLAmount:       2000000,
		TotalPLPercent:      14.28,
		TotalDividendIncome: 480000,
		WinRate:             100,
		HoldingCount:        1,
		WinningCount:        1,
		Portfolios: []dashboard.PortfolioSummary{
			{
				ID:          "p1",
				Name:        "Value",
				Mode:        "VALUE",
				MarketValue: 9000000,
				CostBasis:   8000000,
				PLAmount:    1000000,
				PLPercent:   12.5,
				Weight:      56.25,
			},
		},
		TopGainers: []dashboard.HoldingPL{
			{
				Ticker:        "BBCA",
				PortfolioID:   "p1",
				PortfolioName: "Value",
				MarketValue:   9000000,
				CostBasis:     8000000,
				PLAmount:      1000000,
				PLPercent:     12.5,
			},
		},
		TopLosers: []dashboard.HoldingPL{},
		PortfolioAllocation: []dashboard.AllocationItem{
			{Label: "Value", Value: 9000000, Pct: 56.25},
		},
		SectorAllocation: []dashboard.AllocationItem{
			{Label: "Banking", Value: 9000000, Pct: 56.25},
		},
		RecentTransactions: []transaction.Record{
			{ID: "txn1", Type: transaction.TypeBuy, Date: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), Ticker: "BBCA"},
		},
	}

	got := newDashboardOverviewResponse(overview)

	if got.TotalMarketValue != 16000000 {
		t.Errorf("TotalMarketValue = %f, want 16000000", got.TotalMarketValue)
	}
	if got.TotalPLPercent != 14.28 {
		t.Errorf("TotalPLPercent = %f, want 14.28", got.TotalPLPercent)
	}
	if len(got.Portfolios) != 1 {
		t.Fatalf("Portfolios = %d, want 1", len(got.Portfolios))
	}
	if got.Portfolios[0].Name != "Value" {
		t.Errorf("Portfolio name = %q, want %q", got.Portfolios[0].Name, "Value")
	}
	if len(got.TopGainers) != 1 {
		t.Errorf("TopGainers = %d, want 1", len(got.TopGainers))
	}
	if len(got.TopLosers) != 0 {
		t.Errorf("TopLosers = %d, want 0", len(got.TopLosers))
	}
	if len(got.PortfolioAllocation) != 1 {
		t.Errorf("PortfolioAllocation = %d, want 1", len(got.PortfolioAllocation))
	}
	if len(got.RecentTransactions) != 1 {
		t.Fatalf("RecentTransactions = %d, want 1", len(got.RecentTransactions))
	}
	if got.RecentTransactions[0].Ticker != "BBCA" {
		t.Errorf("transaction ticker = %q, want %q", got.RecentTransactions[0].Ticker, "BBCA")
	}
	if got.WinRate != 100 {
		t.Errorf("WinRate = %f, want 100", got.WinRate)
	}
	if got.HoldingCount != 1 {
		t.Errorf("HoldingCount = %d, want 1", got.HoldingCount)
	}
	if got.WinningCount != 1 {
		t.Errorf("WinningCount = %d, want 1", got.WinningCount)
	}
}
