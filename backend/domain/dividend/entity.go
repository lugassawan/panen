package dividend

// Indicator classifies a dividend holding's current investment state.
type Indicator string

const (
	IndicatorBuyZone    Indicator = "BUY_ZONE"
	IndicatorAverageUp  Indicator = "AVERAGE_UP"
	IndicatorHold       Indicator = "HOLD"
	IndicatorOvervalued Indicator = "OVERVALUED"
)

// DividendMetrics holds computed dividend-specific metrics for a holding.
type DividendMetrics struct {
	Indicator      Indicator
	AnnualDPS      float64 // derived: Price * DY / 100
	YieldOnCost    float64 // DPS / AvgBuyPrice * 100
	ProjectedYoC   float64 // YoC after averaging up at current price
	PortfolioYield float64 // weighted portfolio-level yield
}

// RankItem represents a single stock (holding or watchlist candidate) in
// the dividend attractiveness ranking.
type RankItem struct {
	Ticker      string
	Indicator   Indicator
	DY          float64
	YieldOnCost float64
	PayoutRatio float64
	PositionPct float64
	Score       float64
	IsHolding   bool
}
