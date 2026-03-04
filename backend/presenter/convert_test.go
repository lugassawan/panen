package presenter

import (
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
	"github.com/lugassawan/panen/backend/usecase"
)

func TestNewStockValuationResponseWithBands(t *testing.T) {
	data := &stock.Data{
		Ticker:    "BBCA",
		Price:     9000,
		FetchedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Source:    "test",
	}
	result := &valuation.ValuationResult{
		GrahamNumber: 8000,
		Verdict:      valuation.VerdictUndervalued,
		PBVBand: &valuation.BandStats{
			Min: 1.0, Max: 3.0, Avg: 2.0, Median: 1.8,
		},
		PERBand: &valuation.BandStats{
			Min: 10.0, Max: 20.0, Avg: 15.0, Median: 14.0,
		},
	}

	resp := newStockValuationResponse(data, result, "MODERATE")

	if resp.PBVBand == nil {
		t.Fatal("expected PBVBand to be non-nil")
	}
	if resp.PBVBand.Min != 1.0 {
		t.Errorf("PBVBand.Min = %v, want 1.0", resp.PBVBand.Min)
	}
	if resp.PBVBand.Max != 3.0 {
		t.Errorf("PBVBand.Max = %v, want 3.0", resp.PBVBand.Max)
	}
	if resp.PBVBand.Avg != 2.0 {
		t.Errorf("PBVBand.Avg = %v, want 2.0", resp.PBVBand.Avg)
	}
	if resp.PBVBand.Median != 1.8 {
		t.Errorf("PBVBand.Median = %v, want 1.8", resp.PBVBand.Median)
	}

	if resp.PERBand == nil {
		t.Fatal("expected PERBand to be non-nil")
	}
	if resp.PERBand.Min != 10.0 {
		t.Errorf("PERBand.Min = %v, want 10.0", resp.PERBand.Min)
	}
	if resp.PERBand.Max != 20.0 {
		t.Errorf("PERBand.Max = %v, want 20.0", resp.PERBand.Max)
	}
	if resp.PERBand.Avg != 15.0 {
		t.Errorf("PERBand.Avg = %v, want 15.0", resp.PERBand.Avg)
	}
	if resp.PERBand.Median != 14.0 {
		t.Errorf("PERBand.Median = %v, want 14.0", resp.PERBand.Median)
	}
}

func TestNewStockValuationResponseNilBands(t *testing.T) {
	data := &stock.Data{
		Ticker:    "BBRI",
		Price:     5000,
		FetchedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Source:    "test",
	}
	result := &valuation.ValuationResult{
		GrahamNumber: 4500,
		Verdict:      valuation.VerdictFair,
		PBVBand:      nil,
		PERBand:      nil,
	}

	resp := newStockValuationResponse(data, result, "CONSERVATIVE")

	if resp.PBVBand != nil {
		t.Errorf("expected PBVBand to be nil, got %+v", resp.PBVBand)
	}
	if resp.PERBand != nil {
		t.Errorf("expected PERBand to be nil, got %+v", resp.PERBand)
	}
}

func TestNewBandStatsResponse(t *testing.T) {
	t.Run("non-nil input", func(t *testing.T) {
		band := &valuation.BandStats{
			Min: 5.0, Max: 25.0, Avg: 12.5, Median: 11.0,
		}
		resp := newBandStatsResponse(band)
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Min != 5.0 {
			t.Errorf("Min = %v, want 5.0", resp.Min)
		}
		if resp.Max != 25.0 {
			t.Errorf("Max = %v, want 25.0", resp.Max)
		}
		if resp.Avg != 12.5 {
			t.Errorf("Avg = %v, want 12.5", resp.Avg)
		}
		if resp.Median != 11.0 {
			t.Errorf("Median = %v, want 11.0", resp.Median)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		resp := newBandStatsResponse(nil)
		if resp != nil {
			t.Errorf("expected nil response, got %+v", resp)
		}
	})
}

func TestNewBrokerageAccountResponse(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	acct := &brokerage.Account{
		ID:          "b1",
		BrokerName:  "Ajaib",
		BrokerCode:  "AJAIB",
		BuyFeePct:   0.15,
		SellFeePct:  0.25,
		SellTaxPct:  0.1,
		IsManualFee: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	resp := newBrokerageAccountResponse(acct)

	if resp.ID != "b1" {
		t.Errorf("ID = %q, want b1", resp.ID)
	}
	if resp.BrokerName != "Ajaib" {
		t.Errorf("BrokerName = %q, want Ajaib", resp.BrokerName)
	}
	if resp.BrokerCode != "AJAIB" {
		t.Errorf("BrokerCode = %q, want AJAIB", resp.BrokerCode)
	}
	if resp.BuyFeePct != 0.15 {
		t.Errorf("BuyFeePct = %v, want 0.15", resp.BuyFeePct)
	}
	if resp.SellFeePct != 0.25 {
		t.Errorf("SellFeePct = %v, want 0.25", resp.SellFeePct)
	}
	if resp.SellTaxPct != 0.1 {
		t.Errorf("SellTaxPct = %v, want 0.1", resp.SellTaxPct)
	}
	if !resp.IsManualFee {
		t.Error("IsManualFee = false, want true")
	}
	if resp.CreatedAt != "2025-06-15T10:30:00Z" {
		t.Errorf("CreatedAt = %q, want 2025-06-15T10:30:00Z", resp.CreatedAt)
	}
}

func TestNewCheckResultResponse(t *testing.T) {
	cr := checklist.CheckResult{
		Key:    "roe_above_min",
		Label:  "ROE above minimum",
		Type:   checklist.CheckTypeAuto,
		Status: checklist.CheckStatusPass,
		Detail: "ROE: 15.00% (min: 10.00%)",
	}

	resp := newCheckResultResponse(cr)

	if resp.Key != "roe_above_min" {
		t.Errorf("Key = %q, want roe_above_min", resp.Key)
	}
	if resp.Label != "ROE above minimum" {
		t.Errorf("Label = %q, want ROE above minimum", resp.Label)
	}
	if resp.Type != "AUTO" {
		t.Errorf("Type = %q, want AUTO", resp.Type)
	}
	if resp.Status != "PASS" {
		t.Errorf("Status = %q, want PASS", resp.Status)
	}
	if resp.Detail != "ROE: 15.00% (min: 10.00%)" {
		t.Errorf("Detail = %q, want ROE: 15.00%% (min: 10.00%%)", resp.Detail)
	}
}

func TestNewSuggestionResponse(t *testing.T) {
	t.Run("non-nil input", func(t *testing.T) {
		s := &checklist.Suggestion{
			Action:          checklist.ActionBuy,
			Ticker:          "BBCA",
			Lots:            5,
			PricePerShare:   9000,
			GrossCost:       4500000,
			Fee:             6750,
			Tax:             0,
			NetCost:         4506750,
			NewAvgBuyPrice:  9000,
			NewPositionLots: 5,
			NewPositionPct:  20.5,
			CapitalGainPct:  0,
		}

		resp := newSuggestionResponse(s)

		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Action != "BUY" {
			t.Errorf("Action = %q, want BUY", resp.Action)
		}
		if resp.Ticker != "BBCA" {
			t.Errorf("Ticker = %q, want BBCA", resp.Ticker)
		}
		if resp.Lots != 5 {
			t.Errorf("Lots = %d, want 5", resp.Lots)
		}
		if resp.PricePerShare != 9000 {
			t.Errorf("PricePerShare = %v, want 9000", resp.PricePerShare)
		}
		if resp.GrossCost != 4500000 {
			t.Errorf("GrossCost = %v, want 4500000", resp.GrossCost)
		}
		if resp.Fee != 6750 {
			t.Errorf("Fee = %v, want 6750", resp.Fee)
		}
		if resp.Tax != 0 {
			t.Errorf("Tax = %v, want 0", resp.Tax)
		}
		if resp.NetCost != 4506750 {
			t.Errorf("NetCost = %v, want 4506750", resp.NetCost)
		}
		if resp.NewAvgBuyPrice != 9000 {
			t.Errorf("NewAvgBuyPrice = %v, want 9000", resp.NewAvgBuyPrice)
		}
		if resp.NewPositionLots != 5 {
			t.Errorf("NewPositionLots = %d, want 5", resp.NewPositionLots)
		}
		if resp.NewPositionPct != 20.5 {
			t.Errorf("NewPositionPct = %v, want 20.5", resp.NewPositionPct)
		}
		if resp.CapitalGainPct != 0 {
			t.Errorf("CapitalGainPct = %v, want 0", resp.CapitalGainPct)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		resp := newSuggestionResponse(nil)
		if resp != nil {
			t.Errorf("expected nil response, got %+v", resp)
		}
	})
}

func TestNewChecklistEvaluationResponseWithSuggestion(t *testing.T) {
	eval := &usecase.ChecklistEvaluation{
		Action: checklist.ActionBuy,
		Ticker: "BBCA",
		Checks: []checklist.CheckResult{
			{
				Key:    "roe_above_min",
				Label:  "ROE above minimum",
				Type:   checklist.CheckTypeAuto,
				Status: checklist.CheckStatusPass,
				Detail: "ROE: 15.00%",
			},
			{
				Key:    "confirm_research",
				Label:  "Research confirmed",
				Type:   checklist.CheckTypeManual,
				Status: checklist.CheckStatusPass,
			},
		},
		AllPassed: true,
		Suggestion: &checklist.Suggestion{
			Action:        checklist.ActionBuy,
			Ticker:        "BBCA",
			Lots:          3,
			PricePerShare: 9000,
			GrossCost:     2700000,
			Fee:           4050,
			NetCost:       2704050,
		},
	}

	resp := newChecklistEvaluationResponse(eval)

	if resp.Action != "BUY" {
		t.Errorf("Action = %q, want BUY", resp.Action)
	}
	if resp.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want BBCA", resp.Ticker)
	}
	if len(resp.Checks) != 2 {
		t.Fatalf("len(Checks) = %d, want 2", len(resp.Checks))
	}
	if resp.Checks[0].Key != "roe_above_min" {
		t.Errorf("Checks[0].Key = %q, want roe_above_min", resp.Checks[0].Key)
	}
	if resp.Checks[0].Type != "AUTO" {
		t.Errorf("Checks[0].Type = %q, want AUTO", resp.Checks[0].Type)
	}
	if resp.Checks[1].Key != "confirm_research" {
		t.Errorf("Checks[1].Key = %q, want confirm_research", resp.Checks[1].Key)
	}
	if resp.Checks[1].Type != "MANUAL" {
		t.Errorf("Checks[1].Type = %q, want MANUAL", resp.Checks[1].Type)
	}
	if !resp.AllPassed {
		t.Error("AllPassed = false, want true")
	}
	if resp.Suggestion == nil {
		t.Fatal("expected non-nil Suggestion")
	}
	if resp.Suggestion.Lots != 3 {
		t.Errorf("Suggestion.Lots = %d, want 3", resp.Suggestion.Lots)
	}
}

func TestNewChecklistEvaluationResponseWithoutSuggestion(t *testing.T) {
	eval := &usecase.ChecklistEvaluation{
		Action: checklist.ActionHold,
		Ticker: "BBRI",
		Checks: []checklist.CheckResult{
			{
				Key:    "fundamentals_stable",
				Label:  "Fundamentals stable",
				Type:   checklist.CheckTypeAuto,
				Status: checklist.CheckStatusFail,
				Detail: "ROE: 8.00% (fail)",
			},
		},
		AllPassed:  false,
		Suggestion: nil,
	}

	resp := newChecklistEvaluationResponse(eval)

	if resp.Action != "HOLD" {
		t.Errorf("Action = %q, want HOLD", resp.Action)
	}
	if resp.AllPassed {
		t.Error("AllPassed = true, want false")
	}
	if resp.Suggestion != nil {
		t.Errorf("expected nil Suggestion, got %+v", resp.Suggestion)
	}
	if len(resp.Checks) != 1 {
		t.Fatalf("len(Checks) = %d, want 1", len(resp.Checks))
	}
	if resp.Checks[0].Status != "FAIL" {
		t.Errorf("Checks[0].Status = %q, want FAIL", resp.Checks[0].Status)
	}
}
