package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
)

func TestGetDividendRankingHappyPath(t *testing.T) {
	ctx := context.Background()
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	stockRepo := newMockStockRepo()

	p := &portfolio.Portfolio{
		ID:          "p1",
		Mode:        portfolio.ModeDividend,
		RiskProfile: portfolio.RiskProfileModerate,
		Universe:    []string{"BBCA", "TLKM", "UNVR"},
	}
	_ = portfolioRepo.Create(ctx, p)

	h := &portfolio.Holding{
		ID:          "h1",
		PortfolioID: "p1",
		Ticker:      "BBCA",
		AvgBuyPrice: 8000,
		Lots:        10,
	}
	_ = holdingRepo.Create(ctx, h)

	for _, ticker := range []string{"BBCA", "TLKM", "UNVR"} {
		_ = stockRepo.Upsert(ctx, &stock.Data{
			Ticker:        ticker,
			Source:        "test",
			Price:         8500,
			EPS:           500,
			BVPS:          3000,
			PBV:           2.8,
			PER:           17,
			DividendYield: 3.0,
			PayoutRatio:   40,
		})
	}

	svc := NewDividendService(portfolioRepo, holdingRepo, stockRepo)
	items, err := svc.GetDividendRanking(ctx, "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	// Verify BBCA is marked as a holding
	found := false
	for _, item := range items {
		if item.Ticker == "BBCA" {
			found = true
			if !item.IsHolding {
				t.Error("expected BBCA IsHolding=true")
			}
			if item.YieldOnCost == 0 {
				t.Error("expected non-zero YieldOnCost for held ticker")
			}
		}
	}
	if !found {
		t.Error("expected BBCA in rank items")
	}

	// Verify universe tickers are not holdings
	for _, item := range items {
		if item.Ticker == "TLKM" || item.Ticker == "UNVR" {
			if item.IsHolding {
				t.Errorf("expected %s IsHolding=false", item.Ticker)
			}
		}
	}
}

func TestGetDividendRankingWrongMode(t *testing.T) {
	ctx := context.Background()
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	stockRepo := newMockStockRepo()

	p := &portfolio.Portfolio{
		ID:   "p1",
		Mode: portfolio.ModeValue,
	}
	_ = portfolioRepo.Create(ctx, p)

	svc := NewDividendService(portfolioRepo, holdingRepo, stockRepo)
	_, err := svc.GetDividendRanking(ctx, "p1")
	if err == nil {
		t.Fatal("expected error for VALUE mode portfolio")
	}
	if !strings.Contains(err.Error(), "not in DIVIDEND mode") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCollectTickerDataSkipsHeldFromUniverse(t *testing.T) {
	ctx := context.Background()
	stockRepo := newMockStockRepo()

	// Set up stock data for both held and universe tickers
	for _, ticker := range []string{"BBCA", "TLKM", "UNVR"} {
		_ = stockRepo.Upsert(ctx, &stock.Data{
			Ticker:        ticker,
			Source:        "test",
			Price:         8500,
			DividendYield: 3.0,
			EPS:           500,
			BVPS:          3000,
			PBV:           2.8,
			PER:           17,
		})
	}

	svc := NewDividendService(nil, nil, stockRepo)

	holdings := []*portfolio.Holding{
		{ID: "h1", PortfolioID: "p1", Ticker: "BBCA", AvgBuyPrice: 8000, Lots: 10},
	}
	p := &portfolio.Portfolio{
		ID:          "p1",
		Mode:        portfolio.ModeDividend,
		RiskProfile: portfolio.RiskProfileModerate,
		Universe:    []string{"BBCA", "TLKM", "UNVR"},
	}

	holdingSet, infoMap, totalValue := svc.collectTickerData(ctx, holdings, p)

	// BBCA should be in holdingSet
	if _, ok := holdingSet["BBCA"]; !ok {
		t.Error("expected BBCA in holdingSet")
	}

	// All 3 tickers should be in infoMap (BBCA from holdings, TLKM+UNVR from universe)
	if len(infoMap) != 3 {
		t.Errorf("got %d entries in infoMap, want 3", len(infoMap))
	}

	// totalValue should be BBCA's position value: 8500 * 10 * 100
	expectedValue := 8500.0 * 10 * 100
	if totalValue != expectedValue {
		t.Errorf("totalValue = %v, want %v", totalValue, expectedValue)
	}
}

func TestBuildRankItemsComputation(t *testing.T) {
	infoMap := map[string]tickerInfo{
		"BBCA": {
			data: &stock.Data{
				Ticker:        "BBCA",
				Price:         8500,
				DividendYield: 3.0,
				PayoutRatio:   40,
				EPS:           500,
				BVPS:          3000,
				PBV:           2.8,
				PER:           17,
			},
		},
		"TLKM": {
			data: &stock.Data{
				Ticker:        "TLKM",
				Price:         4000,
				DividendYield: 5.0,
				PayoutRatio:   60,
				EPS:           200,
				BVPS:          1500,
				PBV:           2.7,
				PER:           20,
			},
		},
	}

	holdingSet := map[string]*portfolio.Holding{
		"BBCA": {ID: "h1", PortfolioID: "p1", Ticker: "BBCA", AvgBuyPrice: 8000, Lots: 10},
	}

	// Total value: BBCA = 8500 * 10 * 100 = 8_500_000
	totalValue := 8500.0 * 10 * 100

	thresholds := checklist.ThresholdsForRisk(portfolio.RiskProfileModerate)
	items := buildRankItems(infoMap, holdingSet, totalValue, thresholds)

	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}

	for _, item := range items {
		switch item.Ticker {
		case "BBCA":
			if !item.IsHolding {
				t.Error("expected BBCA IsHolding=true")
			}
			if item.DY != 3.0 {
				t.Errorf("BBCA DY = %v, want 3.0", item.DY)
			}
			if item.PayoutRatio != 40 {
				t.Errorf("BBCA PayoutRatio = %v, want 40", item.PayoutRatio)
			}
			// YoC = annualDPS / avgBuyPrice * 100
			// annualDPS = 8500 * 3.0 / 100 = 255
			// YoC = 255 / 8000 * 100 = 3.1875
			if item.YieldOnCost == 0 {
				t.Error("expected non-zero YieldOnCost for BBCA")
			}
			// PositionPct should be 100% since it's the only holding
			if item.PositionPct != 100 {
				t.Errorf("BBCA PositionPct = %v, want 100", item.PositionPct)
			}
		case "TLKM":
			if item.IsHolding {
				t.Error("expected TLKM IsHolding=false")
			}
			if item.DY != 5.0 {
				t.Errorf("TLKM DY = %v, want 5.0", item.DY)
			}
			if item.YieldOnCost != 0 {
				t.Errorf("expected zero YieldOnCost for non-held ticker, got %v", item.YieldOnCost)
			}
			if item.PositionPct != 0 {
				t.Errorf("expected zero PositionPct for non-held ticker, got %v", item.PositionPct)
			}
		}
	}
}
