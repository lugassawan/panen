package financials

import "time"

// PeriodType distinguishes annual from quarterly financial data.
type PeriodType string

const (
	PeriodAnnual    PeriodType = "annual"
	PeriodQuarterly PeriodType = "quarterly"
)

// IncomeStatement holds income statement data for a single fiscal period.
type IncomeStatement struct {
	ID                string
	Ticker            string
	Source            string
	FiscalYear        int
	Quarter           int // 0 for annual
	Period            PeriodType
	FetchedAt         time.Time
	Revenue           float64
	CostOfRevenue     float64
	GrossProfit       float64
	OperatingExpenses float64
	OperatingIncome   float64
	NetIncome         float64
	EPS               float64
	EPSDiluted        float64
	EBITDA            float64
	InterestExpense   float64
	IncomeTaxExpense  float64
}

// BalanceSheet holds balance sheet data for a single fiscal period.
type BalanceSheet struct {
	ID                 string
	Ticker             string
	Source             string
	FiscalYear         int
	Quarter            int
	Period             PeriodType
	FetchedAt          time.Time
	TotalAssets        float64
	TotalCurrentAssets float64
	CashAndEquivalents float64
	Receivables        float64
	Inventory          float64
	IntangibleAssets   float64
	TotalLiabilities   float64
	TotalCurrentLiab   float64
	LongTermDebt       float64
	TotalDebt          float64
	TotalEquity        float64
	RetainedEarnings   float64
	SharesOutstanding  float64
}

// CashFlowStatement holds cash flow statement data for a single fiscal period.
type CashFlowStatement struct {
	ID                 string
	Ticker             string
	Source             string
	FiscalYear         int
	Quarter            int
	Period             PeriodType
	FetchedAt          time.Time
	OperatingCashFlow  float64
	CapitalExpenditure float64
	FreeCashFlow       float64
	DividendsPaid      float64
	NetBorrowings      float64
	InvestingCashFlow  float64
	FinancingCashFlow  float64
	NetChangeInCash    float64
}
