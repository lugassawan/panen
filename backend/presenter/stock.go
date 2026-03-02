package presenter

// LookupStock fetches or refreshes stock data and returns valuation results.
func (a *App) LookupStock(ticker, riskProfile string) (*StockValuationResponse, error) {
	rp, err := toValuationRisk(riskProfile)
	if err != nil {
		return nil, err
	}
	data, result, err := a.stocks.Lookup(a.ctx, ticker, rp)
	if err != nil {
		return nil, err
	}
	return buildStockResponse(data, result, riskProfile), nil
}

// GetStockValuation returns cached stock valuation without fetching new data.
func (a *App) GetStockValuation(ticker, riskProfile string) (*StockValuationResponse, error) {
	rp, err := toValuationRisk(riskProfile)
	if err != nil {
		return nil, err
	}
	data, result, err := a.stocks.GetCached(a.ctx, ticker, rp)
	if err != nil {
		return nil, err
	}
	return buildStockResponse(data, result, riskProfile), nil
}
