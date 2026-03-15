package sectorsapp

// financialReportResponse maps the SectorsApp financial report JSON.
// SectorsApp returns a unified report per period containing all three statements.
type financialReportResponse struct {
	Symbol  string `json:"symbol"`
	Year    int    `json:"year"`
	Quarter int    `json:"quarter"` // 0 for annual

	// Income statement fields
	Revenue           float64 `json:"revenue"`
	CostOfRevenue     float64 `json:"cost_of_revenue"`
	GrossProfit       float64 `json:"gross_profit"`
	OperatingExpenses float64 `json:"operating_expenses"`
	OperatingIncome   float64 `json:"operating_income"`
	NetIncome         float64 `json:"net_income"`
	EPS               float64 `json:"earnings_per_share"`
	EPSDiluted        float64 `json:"earnings_per_share_diluted"`
	EBITDA            float64 `json:"ebitda"`
	InterestExpense   float64 `json:"interest_expense"`
	IncomeTaxExpense  float64 `json:"income_tax_expense"`

	// Balance sheet fields
	TotalAssets        float64 `json:"total_assets"`
	TotalCurrentAssets float64 `json:"total_current_assets"`
	CashAndEquivalents float64 `json:"cash_and_equivalents"`
	Receivables        float64 `json:"receivables"`
	Inventory          float64 `json:"inventory"`
	IntangibleAssets   float64 `json:"intangible_assets"`
	TotalLiabilities   float64 `json:"total_liabilities"`
	TotalCurrentLiab   float64 `json:"total_current_liabilities"`
	LongTermDebt       float64 `json:"long_term_debt"`
	TotalDebt          float64 `json:"total_debt"`
	TotalEquity        float64 `json:"total_equity"`
	RetainedEarnings   float64 `json:"retained_earnings"`
	SharesOutstanding  float64 `json:"shares_outstanding"`

	// Cash flow fields
	OperatingCashFlow  float64 `json:"operating_cash_flow"`
	CapitalExpenditure float64 `json:"capital_expenditure"`
	FreeCashFlow       float64 `json:"free_cash_flow"`
	DividendsPaid      float64 `json:"dividends_paid"`
	NetBorrowings      float64 `json:"net_borrowings"`
	InvestingCashFlow  float64 `json:"investing_cash_flow"`
	FinancingCashFlow  float64 `json:"financing_cash_flow"`
	NetChangeInCash    float64 `json:"net_change_in_cash"`
}
