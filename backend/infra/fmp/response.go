package fmp

// incomeStatementResponse maps the FMP income-statement JSON.
type incomeStatementResponse struct {
	Date              string  `json:"date"`
	Symbol            string  `json:"symbol"`
	CalendarYear      string  `json:"calendarYear"`
	Period            string  `json:"period"`
	Revenue           float64 `json:"revenue"`
	CostOfRevenue     float64 `json:"costOfRevenue"`
	GrossProfit       float64 `json:"grossProfit"`
	OperatingExpenses float64 `json:"operatingExpenses"`
	OperatingIncome   float64 `json:"operatingIncome"`
	NetIncome         float64 `json:"netIncome"`
	EPS               float64 `json:"eps"`
	EPSDiluted        float64 `json:"epsdiluted"`
	EBITDA            float64 `json:"ebitda"`
	InterestExpense   float64 `json:"interestExpense"`
	IncomeTaxExpense  float64 `json:"incomeTaxExpense"`
}

// balanceSheetResponse maps the FMP balance-sheet-statement JSON.
type balanceSheetResponse struct {
	Date                         string  `json:"date"`
	Symbol                       string  `json:"symbol"`
	CalendarYear                 string  `json:"calendarYear"`
	Period                       string  `json:"period"`
	TotalAssets                  float64 `json:"totalAssets"`
	TotalCurrentAssets           float64 `json:"totalCurrentAssets"`
	CashAndCashEquivalents       float64 `json:"cashAndCashEquivalents"`
	NetReceivables               float64 `json:"netReceivables"`
	Inventory                    float64 `json:"inventory"`
	IntangibleAssets             float64 `json:"intangibleAssets"`
	TotalLiabilities             float64 `json:"totalLiabilities"`
	TotalCurrentLiabilities      float64 `json:"totalCurrentLiabilities"`
	LongTermDebt                 float64 `json:"longTermDebt"`
	TotalDebt                    float64 `json:"totalDebt"`
	TotalStockholdersEquity      float64 `json:"totalStockholdersEquity"`
	RetainedEarnings             float64 `json:"retainedEarnings"`
	CommonStockSharesOutstanding float64 `json:"commonStockSharesOutstanding"`
}

// cashFlowResponse maps the FMP cash-flow-statement JSON.
type cashFlowResponse struct {
	Date                                     string  `json:"date"`
	Symbol                                   string  `json:"symbol"`
	CalendarYear                             string  `json:"calendarYear"`
	Period                                   string  `json:"period"`
	OperatingCashFlow                        float64 `json:"operatingCashFlow"`
	CapitalExpenditure                       float64 `json:"capitalExpenditure"`
	FreeCashFlow                             float64 `json:"freeCashFlow"`
	DividendsPaid                            float64 `json:"dividendsPaid"`
	DebtRepayment                            float64 `json:"debtRepayment"`
	NetCashUsedForInvestingActivites         float64 `json:"netCashUsedForInvestingActivites"`
	NetCashUsedProvidedByFinancingActivities float64 `json:"netCashUsedProvidedByFinancingActivities"`
	NetChangeInCash                          float64 `json:"netChangeInCash"`
}
