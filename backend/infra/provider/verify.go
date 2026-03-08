package provider

import "github.com/lugassawan/panen/backend/domain/stock"

// Compile-time checks that our providers implement stock.DataProvider.
var (
	_ stock.DataProvider = (*IDXProvider)(nil)
	_ stock.DataProvider = (*Registry)(nil)
)
