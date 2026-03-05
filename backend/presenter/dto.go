package presenter

const dateLayout = "2006-01-02"

// BandStatsResponse is the frontend-facing response for PBV/PER band statistics.
type BandStatsResponse struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Avg    float64 `json:"avg"`
	Median float64 `json:"median"`
}

// StockValuationResponse is the frontend-facing response for stock lookup.
type StockValuationResponse struct {
	Ticker         string             `json:"ticker"`
	Price          float64            `json:"price"`
	High52Week     float64            `json:"high52Week"`
	Low52Week      float64            `json:"low52Week"`
	EPS            float64            `json:"eps"`
	BVPS           float64            `json:"bvps"`
	ROE            float64            `json:"roe"`
	DER            float64            `json:"der"`
	PBV            float64            `json:"pbv"`
	PER            float64            `json:"per"`
	DividendYield  float64            `json:"dividendYield"`
	PayoutRatio    float64            `json:"payoutRatio"`
	GrahamNumber   float64            `json:"grahamNumber"`
	MarginOfSafety float64            `json:"marginOfSafety"`
	EntryPrice     float64            `json:"entryPrice"`
	ExitTarget     float64            `json:"exitTarget"`
	PBVBand        *BandStatsResponse `json:"pbvBand,omitempty"`
	PERBand        *BandStatsResponse `json:"perBand,omitempty"`
	Verdict        string             `json:"verdict"`
	RiskProfile    string             `json:"riskProfile"`
	FetchedAt      string             `json:"fetchedAt"`
	Source         string             `json:"source"`
}

// BrokerageAccountResponse is the frontend-facing response for a brokerage account.
type BrokerageAccountResponse struct {
	ID          string  `json:"id"`
	BrokerName  string  `json:"brokerName"`
	BrokerCode  string  `json:"brokerCode"`
	BuyFeePct   float64 `json:"buyFeePct"`
	SellFeePct  float64 `json:"sellFeePct"`
	SellTaxPct  float64 `json:"sellTaxPct"`
	IsManualFee bool    `json:"isManualFee"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// BrokerConfigResponse is the frontend-facing response for a broker configuration.
type BrokerConfigResponse struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	BuyFeePct  float64 `json:"buyFeePct"`
	SellFeePct float64 `json:"sellFeePct"`
	SellTaxPct float64 `json:"sellTaxPct"`
	Notes      string  `json:"notes"`
}

// PortfolioResponse is the frontend-facing response for a portfolio.
type PortfolioResponse struct {
	ID              string  `json:"id"`
	BrokerageAcctID string  `json:"brokerageAcctId"`
	Name            string  `json:"name"`
	Mode            string  `json:"mode"`
	RiskProfile     string  `json:"riskProfile"`
	Capital         float64 `json:"capital"`
	MonthlyAddition float64 `json:"monthlyAddition"`
	MaxStocks       int     `json:"maxStocks"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// HoldingDetailResponse is the frontend-facing response for a holding with valuation.
type HoldingDetailResponse struct {
	ID              string                   `json:"id"`
	Ticker          string                   `json:"ticker"`
	AvgBuyPrice     float64                  `json:"avgBuyPrice"`
	Lots            int                      `json:"lots"`
	CurrentPrice    *float64                 `json:"currentPrice,omitempty"`
	GrahamNumber    *float64                 `json:"grahamNumber,omitempty"`
	EntryPrice      *float64                 `json:"entryPrice,omitempty"`
	ExitTarget      *float64                 `json:"exitTarget,omitempty"`
	Verdict         *string                  `json:"verdict,omitempty"`
	MarginOfSafety  *float64                 `json:"marginOfSafety,omitempty"`
	TrailingStop    *TrailingStopResponse    `json:"trailingStop,omitempty"`
	DividendMetrics *DividendMetricsResponse `json:"dividendMetrics,omitempty"`
}

// DividendMetricsResponse is the frontend-facing response for dividend metrics.
type DividendMetricsResponse struct {
	Indicator      string  `json:"indicator"`
	AnnualDPS      float64 `json:"annualDPS"`
	YieldOnCost    float64 `json:"yieldOnCost"`
	ProjectedYoC   float64 `json:"projectedYoC"`
	PortfolioYield float64 `json:"portfolioYield"`
}

// DividendRankItemResponse is the frontend-facing response for a dividend ranking item.
type DividendRankItemResponse struct {
	Ticker      string  `json:"ticker"`
	Indicator   string  `json:"indicator"`
	DY          float64 `json:"dividendYield"`
	YoC         float64 `json:"yieldOnCost"`
	PayoutRatio float64 `json:"payoutRatio"`
	PositionPct float64 `json:"positionPct"`
	Score       float64 `json:"score"`
	IsHolding   bool    `json:"isHolding"`
}

// FundamentalExitResponse is the frontend-facing response for a fundamental exit criterion.
type FundamentalExitResponse struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Detail    string `json:"detail"`
	Triggered bool   `json:"triggered"`
}

// TrailingStopResponse is the frontend-facing response for trailing stop data.
type TrailingStopResponse struct {
	PeakPrice        float64                   `json:"peakPrice"`
	StopPercentage   float64                   `json:"stopPercentage"`
	StopPrice        float64                   `json:"stopPrice"`
	Triggered        bool                      `json:"triggered"`
	FundamentalExits []FundamentalExitResponse `json:"fundamentalExits"`
}

// PortfolioDetailResponse is the frontend-facing response for a portfolio with holdings.
type PortfolioDetailResponse struct {
	Portfolio PortfolioResponse       `json:"portfolio"`
	Holdings  []HoldingDetailResponse `json:"holdings"`
}

// WatchlistResponse is the frontend-facing response for a watchlist.
type WatchlistResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// WatchlistItemResponse is the frontend-facing response for a watchlist item with optional data.
type WatchlistItemResponse struct {
	Ticker        string   `json:"ticker"`
	Sector        string   `json:"sector"`
	Price         *float64 `json:"price,omitempty"`
	ROE           *float64 `json:"roe,omitempty"`
	DER           *float64 `json:"der,omitempty"`
	EPS           *float64 `json:"eps,omitempty"`
	DividendYield *float64 `json:"dividendYield,omitempty"`
	PayoutRatio   *float64 `json:"payoutRatio,omitempty"`
	GrahamNumber  *float64 `json:"grahamNumber,omitempty"`
	EntryPrice    *float64 `json:"entryPrice,omitempty"`
	ExitTarget    *float64 `json:"exitTarget,omitempty"`
	Verdict       *string  `json:"verdict,omitempty"`
	FetchedAt     *string  `json:"fetchedAt,omitempty"`
}

// RefreshStatusResponse is the frontend-facing response for refresh status.
type RefreshStatusResponse struct {
	State       string `json:"state"`
	LastRefresh string `json:"lastRefresh"`
	Error       string `json:"error,omitempty"`
}

// RefreshSettingsResponse is the frontend-facing response for refresh settings.
type RefreshSettingsResponse struct {
	AutoRefreshEnabled bool   `json:"autoRefreshEnabled"`
	IntervalMinutes    int    `json:"intervalMinutes"`
	LastRefreshedAt    string `json:"lastRefreshedAt"`
}

// CheckResultResponse is the frontend-facing response for a single check result.
type CheckResultResponse struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// SuggestionResponse is the frontend-facing response for a trade suggestion.
type SuggestionResponse struct {
	Action          string  `json:"action"`
	Ticker          string  `json:"ticker"`
	Lots            int     `json:"lots"`
	PricePerShare   float64 `json:"pricePerShare"`
	GrossCost       float64 `json:"grossCost"`
	Fee             float64 `json:"fee"`
	Tax             float64 `json:"tax"`
	NetCost         float64 `json:"netCost"`
	NewAvgBuyPrice  float64 `json:"newAvgBuyPrice"`
	NewPositionLots int     `json:"newPositionLots"`
	NewPositionPct  float64 `json:"newPositionPct"`
	CapitalGainPct  float64 `json:"capitalGainPct"`
}

// UpdateCheckResponse is the frontend-facing response for an update check.
type UpdateCheckResponse struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseURL"`
}

// ChecklistEvaluationResponse is the frontend-facing response for a checklist evaluation.
type ChecklistEvaluationResponse struct {
	Action     string                `json:"action"`
	Ticker     string                `json:"ticker"`
	Checks     []CheckResultResponse `json:"checks"`
	AllPassed  bool                  `json:"allPassed"`
	Suggestion *SuggestionResponse   `json:"suggestion,omitempty"`
}

// MonthlyPaydayResponse is the frontend-facing response for a monthly payday status.
type MonthlyPaydayResponse struct {
	Month         string                        `json:"month"`
	PaydayDay     int                           `json:"paydayDay"`
	Portfolios    []PortfolioPaydayItemResponse `json:"portfolios"`
	TotalExpected float64                       `json:"totalExpected"`
}

// PortfolioPaydayItemResponse is the frontend-facing response for a portfolio's payday status.
type PortfolioPaydayItemResponse struct {
	PortfolioID   string  `json:"portfolioId"`
	PortfolioName string  `json:"portfolioName"`
	Mode          string  `json:"mode"`
	Expected      float64 `json:"expected"`
	Actual        float64 `json:"actual"`
	Status        string  `json:"status"`
	DeferUntil    *string `json:"deferUntil,omitempty"`
}

// CashFlowSummaryResponse is the frontend-facing response for a cash flow summary.
type CashFlowSummaryResponse struct {
	Items         []CashFlowItemResponse `json:"items"`
	TotalInflow   float64                `json:"totalInflow"`
	TotalDeployed float64                `json:"totalDeployed"`
	Balance       float64                `json:"balance"`
}

// ScreenerCheckResponse is the frontend-facing response for a single screener check.
type ScreenerCheckResponse struct {
	Key    string  `json:"key"`
	Label  string  `json:"label"`
	Status string  `json:"status"`
	Value  float64 `json:"value"`
	Limit  float64 `json:"limit"`
}

// ScreenerItemResponse is the frontend-facing response for a screened stock.
type ScreenerItemResponse struct {
	Ticker        string                  `json:"ticker"`
	Sector        string                  `json:"sector"`
	Price         *float64                `json:"price,omitempty"`
	ROE           *float64                `json:"roe,omitempty"`
	DER           *float64                `json:"der,omitempty"`
	EPS           *float64                `json:"eps,omitempty"`
	PBV           *float64                `json:"pbv,omitempty"`
	PER           *float64                `json:"per,omitempty"`
	DividendYield *float64                `json:"dividendYield,omitempty"`
	GrahamNumber  *float64                `json:"grahamNumber,omitempty"`
	EntryPrice    *float64                `json:"entryPrice,omitempty"`
	ExitTarget    *float64                `json:"exitTarget,omitempty"`
	Verdict       *string                 `json:"verdict,omitempty"`
	Checks        []ScreenerCheckResponse `json:"checks"`
	Passed        bool                    `json:"passed"`
	Score         float64                 `json:"score"`
	FetchedAt     *string                 `json:"fetchedAt,omitempty"`
}

// PricePointResponse is the frontend-facing response for a single price history point.
type PricePointResponse struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

// CashFlowItemResponse is the frontend-facing response for a single cash flow entry.
type CashFlowItemResponse struct {
	ID          string  `json:"id"`
	PortfolioID string  `json:"portfolioId"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Note        string  `json:"note"`
	CreatedAt   string  `json:"createdAt"`
}
