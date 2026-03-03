package brokerconfig

// BrokerConfig holds fee defaults and metadata for an IDX-registered broker.
type BrokerConfig struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	BuyFeePct  float64 `json:"buyFeePct"`
	SellFeePct float64 `json:"sellFeePct"`
	SellTaxPct float64 `json:"sellTaxPct"`
	Notes      string  `json:"notes"`
}
