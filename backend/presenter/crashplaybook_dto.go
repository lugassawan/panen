package presenter

// MarketStatusResponse is the frontend-facing response for IHSG market status.
type MarketStatusResponse struct {
	Condition   string  `json:"condition"`
	IHSGPrice   float64 `json:"ihsgPrice"`
	IHSGPeak    float64 `json:"ihsgPeak"`
	DrawdownPct float64 `json:"drawdownPct"`
	FetchedAt   string  `json:"fetchedAt"`
}

// ResponseLevelResponse is the frontend-facing response for a single crash response tier.
type ResponseLevelResponse struct {
	Level        string  `json:"level"`
	TriggerPrice float64 `json:"triggerPrice"`
	DeployPct    float64 `json:"deployPct"`
}

// StockPlaybookResponse is the frontend-facing response for a stock's crash playbook.
type StockPlaybookResponse struct {
	Ticker       string                  `json:"ticker"`
	CurrentPrice float64                 `json:"currentPrice"`
	EntryPrice   float64                 `json:"entryPrice"`
	Levels       []ResponseLevelResponse `json:"levels"`
	ActiveLevel  *string                 `json:"activeLevel,omitempty"`
}

// PortfolioPlaybookResponse is the frontend-facing response for a portfolio crash playbook.
type PortfolioPlaybookResponse struct {
	Market     MarketStatusResponse    `json:"market"`
	Stocks     []StockPlaybookResponse `json:"stocks"`
	RefreshMin int                     `json:"refreshMin"`
}

// DiagnosticResponse is the frontend-facing response for falling knife diagnostic.
type DiagnosticResponse struct {
	MarketCrashed  bool   `json:"marketCrashed"`
	CompanyBadNews *bool  `json:"companyBadNews"`
	FundamentalsOK *bool  `json:"fundamentalsOK"`
	BelowEntry     bool   `json:"belowEntry"`
	Signal         string `json:"signal"`
}

// CrashCapitalResponse is the frontend-facing response for crash capital.
type CrashCapitalResponse struct {
	PortfolioID string  `json:"portfolioId"`
	Amount      float64 `json:"amount"`
	Deployed    float64 `json:"deployed"`
}

// DeploymentLevelPlanResponse is the frontend-facing response for a level's deployment plan.
type DeploymentLevelPlanResponse struct {
	Level  string  `json:"level"`
	Pct    float64 `json:"pct"`
	Amount float64 `json:"amount"`
}

// DeploymentPlanResponse is the frontend-facing response for the full deployment plan.
type DeploymentPlanResponse struct {
	Total     float64                       `json:"total"`
	Deployed  float64                       `json:"deployed"`
	Remaining float64                       `json:"remaining"`
	Levels    []DeploymentLevelPlanResponse `json:"levels"`
}

// DeploymentSettingsResponse is the frontend-facing response for deployment percentages.
type DeploymentSettingsResponse struct {
	Normal  float64 `json:"normal"`
	Crash   float64 `json:"crash"`
	Extreme float64 `json:"extreme"`
}
