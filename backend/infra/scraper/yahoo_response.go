package scraper

// chartResponse maps the Yahoo Finance v8 chart API response.
type chartResponse struct {
	Chart struct {
		Result []chartResult `json:"result"`
		Error  *yahooError   `json:"error"`
	} `json:"chart"`
}

type chartResult struct {
	Meta       chartMeta  `json:"meta"`
	Timestamp  []int64    `json:"timestamp"`
	Indicators indicators `json:"indicators"`
}

type chartMeta struct {
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}

type indicators struct {
	Quote []quoteIndicator `json:"quote"`
}

type quoteIndicator struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

// quoteSummaryResponse maps the Yahoo Finance v10 quoteSummary API response.
type quoteSummaryResponse struct {
	QuoteSummary struct {
		Result []quoteSummaryResult `json:"result"`
		Error  *yahooError          `json:"error"`
	} `json:"quoteSummary"`
}

type quoteSummaryResult struct {
	DefaultKeyStatistics defaultKeyStatistics `json:"defaultKeyStatistics"`
	FinancialData        financialData        `json:"financialData"`
	SummaryDetail        summaryDetail        `json:"summaryDetail"`
}

type defaultKeyStatistics struct {
	TrailingEps rawValue `json:"trailingEps"`
	BookValue   rawValue `json:"bookValue"`
	PriceToBook rawValue `json:"priceToBook"`
	TrailingPE  rawValue `json:"trailingPE"`
}

type financialData struct {
	ReturnOnEquity rawValue `json:"returnOnEquity"`
	DebtToEquity   rawValue `json:"debtToEquity"`
}

type summaryDetail struct {
	DividendYield rawValue `json:"dividendYield"`
	PayoutRatio   rawValue `json:"payoutRatio"`
}

type rawValue struct {
	Raw float64 `json:"raw"`
}

type yahooError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}
