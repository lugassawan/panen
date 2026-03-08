# Portfolio Management

Panen organizes your investments into portfolios, each with a single investment mode and risk profile. This guide covers how portfolios work and the workflows for each mode.

## Portfolio Structure

```
Brokerage Account (e.g., Ajaib)
├── Value Portfolio (0 or 1)
│   ├── Risk Profile: Conservative / Moderate / Aggressive
│   ├── Stock Universe: IDX30, LQ45, etc.
│   └── Holdings: BBRI, ASII, UNTR
└── Dividend Portfolio (0 or 1)
    ├── Risk Profile: Conservative / Moderate / Aggressive
    ├── Stock Universe: IDXHIDIV20, LQ45, etc.
    └── Holdings: BBCA, TLKM, BTPS
```

### Key Rules

| Rule | Detail |
|------|--------|
| Portfolios per broker | Max 1 Value + 1 Dividend |
| Stock uniqueness | A stock can exist in only one portfolio per brokerage |
| Cross-broker overlap | The same stock can appear in portfolios at different brokers |
| Portfolio mode | VALUE or DIVIDEND only -- no combined mode |
| Risk profile | Set per portfolio, can differ between portfolios |

### Why No Combined Mode?

This is a deliberate product decision. Separating Value and Dividend modes forces clarity about **why** you hold each stock:

- **Clarity of purpose** -- "Am I buying for growth or income?" If you cannot answer, you should not buy yet
- **Conviction under pressure** -- During crashes, knowing your thesis prevents panic selling
- **Measurable outcomes** -- Value is measured by capital gain percentage; Dividend by yield and income growth

## Value Mode

Value Mode targets capital growth by identifying undervalued stocks and selling when they reach fair value.

### Key Metrics

| Metric | Purpose |
|--------|---------|
| PBV (Price-to-Book Value) | Compare price to net asset value |
| PER (Price-to-Earnings Ratio) | Compare price to earnings |
| Graham Number | Intrinsic value estimate from EPS and BVPS |
| Margin of Safety | Buffer below intrinsic value for entry |

### Workflow: Starter (No Stocks Yet)

The app shows a stock ranking table sorted by value attractiveness:

1. **Valuation verdicts** for each stock (undervalued / fair / overvalued)
2. **Entry price calculator** using Graham Number, PBV/PER 5-year bands, and margin-of-safety-adjusted zones
3. **Capital allocation suggestions** weighted by value score

### Workflow: Owner (Has Holdings)

The portfolio detail page provides:

1. **Portfolio table** -- Average buy price, current price, P/L, valuation status per holding
2. **Per-stock action signals** -- Buy more (if still undervalued), hold, or sell based on valuation zones
3. **Portfolio insights** -- Sector concentration, rebalance suggestions
4. **Exit strategy** -- Conservative, moderate, and aggressive exit price levels
5. **Trailing stop** -- Suggested stop-loss level tracking a percentage below peak price (see [Valuation Models](valuation-models.md))

### Monthly Addition in Value Mode

Capital accumulates in a **war chest** until entry targets are hit. The app does not deploy monthly capital automatically -- it waits for stocks to enter their buy zones, then suggests allocation.

## Dividend Mode

Dividend Mode targets passive income through dividend-paying stocks with strong fundamentals and growing payouts.

### Key Metrics

| Metric | Purpose |
|--------|---------|
| DY (Dividend Yield) | Annual dividend relative to current price |
| YoC (Yield on Cost) | Annual dividend relative to your average buy price |
| DGR (Dividend Growth Rate) | 5-year CAGR of dividend per share |
| DPR (Dividend Payout Ratio) | Percentage of earnings paid as dividends |
| Consistency | Number of consecutive years of dividend payments |

### Workflow: Starter (Building Income)

The app shows a dividend stock ranking:

1. **Dividend attractiveness** -- DY, 5-year average DY, DPR, consistency, DGR
2. **Dividend trap warnings** -- Flags stocks with high yield but declining growth or unsustainable payout ratios
3. **Income simulator** -- Projects dividend income over 5, 10, and 20 years with optional reinvestment

### Workflow: Owner (Optimizing Income)

The dividend dashboard provides:

1. **Income overview** -- Shares, average buy price, DY at buy, current DY, YoC, annual income per holding
2. **Growth tracking** -- Year-over-year income changes and projected timeline to income goals
3. **Reinvestment advisor** -- Ranks watchlist stocks for next capital deployment
4. **Dividend calendar** -- Upcoming ex-dates and estimated income

### Monthly Addition in Dividend Mode

Capital is deployed monthly via **DCA** (dollar-cost averaging) into the stock with the best yield in its buy zone. The app ranks all watchlist stocks -- including existing holdings eligible for average-up -- by attractiveness each month.

### Average Up

A five-step decision framework determines whether averaging up is warranted:

1. Is the stock still fundamentally strong?
2. Current DY meets the minimum for the risk profile?
3. DGR > 5%? (More justified. DGR < 0% means do not add)
4. Is there a better stock to buy instead?
5. Would the position exceed max weight after buying?

## Holdings Management

### Adding a Holding

Enter the ticker, buy price (per share), number of lots (1 lot = 100 shares), and buy date. The app records the transaction with fee calculations based on your brokerage's fee configuration.

### Average Down

When a stock drops below your buy price, the app provides guided average-down support:

- Current loss percentage and valuation status
- Projected new average price if averaging down
- New position weight versus risk profile limits
- Fundamental health check

The app will **not** suggest averaging down on a stock that is not undervalued. The conviction checklist for average-down includes: "I would buy this even if I did not already own it."

### Transaction History

View all buy transactions for each holding, including date, price, lots, and fees.

### Stock Comparison

Compare multiple stocks side-by-side on key metrics to make informed allocation decisions.

## Conviction Checklists

Before acting on any suggestion, the app presents a conviction checklist:

1. **Auto-checks** are pre-filled from data (e.g., "ROE still above 15%", "DER below 1.0")
2. **Manual checks** require you to verify information the app cannot access (e.g., "I checked for recent negative news")
3. All checks must pass before the suggestion unlocks
4. The suggestion includes specific lot counts, costs, fees, and portfolio impact

Checklist strictness varies by risk profile: Conservative requires 10+ items (all must pass), Moderate 8+, and Aggressive 6+.

## Screener

The stock screener filters and ranks stocks within your selected universe. Filter by valuation metrics, dividend characteristics, and fundamental health to discover new opportunities.

## Monthly Payday Assistant

If you have configured a monthly addition amount and payday date:

1. On payday, the app shows a reminder with your planned addition amount
2. You **confirm** the actual amount (the app never assumes money was transferred)
3. Capital is split across portfolios based on your configured ratio
4. The app suggests how to deploy the capital based on current market conditions

See [Getting Started](getting-started.md) for initial configuration.
