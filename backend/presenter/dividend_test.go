package presenter

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/usecase"
)

func newTestDividendHandler() *DividendHandler {
	ctx := context.Background()
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	stockRepo := newMockStockRepo()

	// Create a DIVIDEND portfolio with a holding.
	p := &portfolio.Portfolio{
		ID:                 "p1",
		Name:               "Dividend Portfolio",
		Mode:               portfolio.ModeDividend,
		BrokerageAccountID: "b1",
		RiskProfile:        portfolio.RiskProfileModerate,
	}
	_ = portfolioRepo.Create(ctx, p)

	h := &portfolio.Holding{
		ID:          "h1",
		PortfolioID: "p1",
		Ticker:      "BBCA",
		Lots:        10,
		AvgBuyPrice: 8000,
	}
	_ = holdingRepo.Create(ctx, h)

	// Seed stock data with dividend info.
	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID:            "s1",
		Ticker:        "BBCA",
		Price:         9000,
		DividendYield: 3.0,
		PayoutRatio:   40,
		FetchedAt:     time.Now().UTC(),
		Source:        "mock",
	})

	svc := usecase.NewDividendService(portfolioRepo, holdingRepo, stockRepo)
	return NewDividendHandler(ctx, svc)
}

func TestDividendHandlerGetRanking(t *testing.T) {
	handler := newTestDividendHandler()

	items, err := handler.GetDividendRanking("p1")
	if err != nil {
		t.Fatalf("GetDividendRanking() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one rank item")
	}

	found := false
	for _, item := range items {
		if item.Ticker == "BBCA" {
			found = true
			if !item.IsHolding {
				t.Error("expected BBCA IsHolding = true")
			}
			if item.DY != 3.0 {
				t.Errorf("DY = %v, want 3.0", item.DY)
			}
		}
	}
	if !found {
		t.Error("expected BBCA in ranking results")
	}
}

func TestDividendHandlerGetRankingWrongMode(t *testing.T) {
	ctx := context.Background()
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	stockRepo := newMockStockRepo()

	// Create a VALUE mode portfolio.
	p := &portfolio.Portfolio{
		ID:                 "p1",
		Name:               "Value Portfolio",
		Mode:               portfolio.ModeValue,
		BrokerageAccountID: "b1",
	}
	_ = portfolioRepo.Create(ctx, p)

	svc := usecase.NewDividendService(portfolioRepo, holdingRepo, stockRepo)
	handler := NewDividendHandler(ctx, svc)

	_, err := handler.GetDividendRanking("p1")
	if err == nil {
		t.Error("expected error for VALUE mode portfolio")
	}
}
