package presenter

import (
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
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
