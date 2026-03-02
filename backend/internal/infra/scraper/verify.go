package scraper

import "github.com/lugassawan/panen/backend/internal/domain/stock"

// Compile-time check that Yahoo implements stock.DataProvider.
var _ stock.DataProvider = (*Yahoo)(nil)
