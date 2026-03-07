package presenter

import (
	"testing"

	"github.com/lugassawan/panen/backend/domain/brokerconfig"
)

func TestListBrokerConfigs(t *testing.T) {
	configs := []*brokerconfig.BrokerConfig{
		{Code: "AJ", Name: "Ajaib", BuyFeePct: 0.15, SellFeePct: 0.25, SellTaxPct: 0.1, Notes: "default"},
		{Code: "ST", Name: "Stockbit", BuyFeePct: 0.10, SellFeePct: 0.20, SellTaxPct: 0.1, Notes: ""},
	}
	handler := NewBrokerConfigHandler(configs)
	result := handler.ListBrokerConfigs()

	if len(result) != 2 {
		t.Fatalf("got %d configs, want 2", len(result))
	}

	tests := []struct {
		idx  int
		code string
		name string
		buy  float64
		sell float64
		tax  float64
	}{
		{0, "AJ", "Ajaib", 0.15, 0.25, 0.1},
		{1, "ST", "Stockbit", 0.10, 0.20, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			r := result[tt.idx]
			if r.Code != tt.code {
				t.Errorf("Code = %q, want %q", r.Code, tt.code)
			}
			if r.Name != tt.name {
				t.Errorf("Name = %q, want %q", r.Name, tt.name)
			}
			if r.BuyFeePct != tt.buy {
				t.Errorf("BuyFeePct = %v, want %v", r.BuyFeePct, tt.buy)
			}
			if r.SellFeePct != tt.sell {
				t.Errorf("SellFeePct = %v, want %v", r.SellFeePct, tt.sell)
			}
			if r.SellTaxPct != tt.tax {
				t.Errorf("SellTaxPct = %v, want %v", r.SellTaxPct, tt.tax)
			}
		})
	}
}

func TestListBrokerConfigsEmpty(t *testing.T) {
	handler := NewBrokerConfigHandler(nil)
	result := handler.ListBrokerConfigs()
	if len(result) != 0 {
		t.Errorf("got %d configs for nil input, want 0", len(result))
	}
}

func TestSearchBrokerConfigs(t *testing.T) {
	configs := []*brokerconfig.BrokerConfig{
		{Code: "AJ", Name: "Ajaib", BuyFeePct: 0.15, SellFeePct: 0.25, SellTaxPct: 0.1},
		{Code: "ST", Name: "Stockbit", BuyFeePct: 0.10, SellFeePct: 0.20, SellTaxPct: 0.1},
		{Code: "IP", Name: "IPOT", BuyFeePct: 0.19, SellFeePct: 0.29, SellTaxPct: 0.1},
	}
	handler := NewBrokerConfigHandler(configs)

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "empty query returns all", query: "", want: 3},
		{name: "match by name", query: "Ajaib", want: 1},
		{name: "match by code", query: "IP", want: 1},
		{name: "match by name substring", query: "Stock", want: 1},
		{name: "case insensitive", query: "ajaib", want: 1},
		{name: "no results", query: "XYZ", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.SearchBrokerConfigs(tt.query)
			if len(result) != tt.want {
				t.Errorf("SearchBrokerConfigs(%q) returned %d results, want %d", tt.query, len(result), tt.want)
			}
		})
	}
}
