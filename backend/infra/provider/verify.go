package provider

import (
	domainProvider "github.com/lugassawan/panen/backend/domain/provider"
	"github.com/lugassawan/panen/backend/domain/stock"
)

// Compile-time checks that our providers implement the required interfaces.
var (
	_ stock.DataProvider      = (*IDXProvider)(nil)
	_ stock.DataProvider      = (*Registry)(nil)
	_ domainProvider.Registry = (*Registry)(nil)
)
