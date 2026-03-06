package presenter

import "github.com/lugassawan/panen/backend/domain/brokerconfig"

// BrokerConfigHandler serves broker configuration data loaded at startup.
type BrokerConfigHandler struct {
	configs []*brokerconfig.BrokerConfig
}

// NewBrokerConfigHandler creates a new BrokerConfigHandler.
func NewBrokerConfigHandler(configs []*brokerconfig.BrokerConfig) *BrokerConfigHandler {
	h := &BrokerConfigHandler{}
	h.Bind(configs)
	return h
}

func (h *BrokerConfigHandler) Bind(configs []*brokerconfig.BrokerConfig) {
	h.configs = configs
}

// ListBrokerConfigs returns all known broker configurations.
func (h *BrokerConfigHandler) ListBrokerConfigs() []*BrokerConfigResponse {
	result := make([]*BrokerConfigResponse, len(h.configs))
	for i, c := range h.configs {
		result[i] = &BrokerConfigResponse{
			Code:       c.Code,
			Name:       c.Name,
			BuyFeePct:  c.BuyFeePct,
			SellFeePct: c.SellFeePct,
			SellTaxPct: c.SellTaxPct,
			Notes:      c.Notes,
		}
	}
	return result
}
