package brokerconfig

import (
	"encoding/json"
	"errors"

	"github.com/lugassawan/panen/backend/domain/brokerconfig"
	"github.com/lugassawan/panen/backend/infra/liveconfig"
	"github.com/lugassawan/panen/configs"
)

// NewLoader creates a liveconfig.Loader for broker configurations.
func NewLoader(dataDir string, deps liveconfig.Deps) *liveconfig.Loader[[]*brokerconfig.BrokerConfig] {
	return liveconfig.NewLoader(dataDir, liveconfig.Config[[]*brokerconfig.BrokerConfig]{
		Name:          "brokers",
		RemotePath:    "brokers.json",
		CacheFileName: "brokers.json",
		BundledData:   configs.BrokersJSON,
		ParseFunc:     parseBrokers,
	}, deps)
}

func parseBrokers(data []byte) ([]*brokerconfig.BrokerConfig, error) {
	var cfgs []*brokerconfig.BrokerConfig
	if err := json.Unmarshal(data, &cfgs); err != nil {
		return nil, err
	}
	if len(cfgs) == 0 {
		return nil, errors.New("empty broker configs")
	}
	return cfgs, nil
}
