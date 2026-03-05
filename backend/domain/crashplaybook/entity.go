package crashplaybook

import "time"

// MarketStatus holds the current broad market state derived from IHSG (^JKSE).
type MarketStatus struct {
	Condition   MarketCondition
	IHSGPrice   float64
	IHSGPeak    float64
	DrawdownPct float64
	FetchedAt   time.Time
}

// ResponseLevel defines a single pre-calculated crash response tier.
type ResponseLevel struct {
	Level        CrashLevel
	TriggerPrice float64
	DeployPct    float64
}

// StockPlaybook holds the crash playbook for a single holding.
type StockPlaybook struct {
	Ticker       string
	CurrentPrice float64
	EntryPrice   float64
	Levels       []ResponseLevel
	ActiveLevel  *CrashLevel
}

// CrashCapital represents pre-committed capital reserved for crash deployments.
type CrashCapital struct {
	ID          string
	PortfolioID string
	Amount      float64
	Deployed    float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FallingKnifeDiagnostic holds the 4-check evaluation for a stock.
type FallingKnifeDiagnostic struct {
	MarketCrashed  bool
	CompanyBadNews *bool
	FundamentalsOK *bool
	BelowEntry     bool
	Signal         DiagnosticSignal
}
