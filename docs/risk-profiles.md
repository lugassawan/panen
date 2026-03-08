# Risk Profiles

Every portfolio in Panen has a risk profile that determines valuation thresholds, position sizing limits, checklist strictness, and exit strategy parameters. The risk profile is a **calculation parameter**, not just a label.

## Three Profiles

### Conservative

Best for: Investors who prioritize capital preservation, prefer blue-chip stocks, and want strict discipline.

- Focuses on the safest stocks (IDX30 blue chips)
- Requires the deepest discounts before buying (30-50% margin of safety)
- Uses Graham Number as the primary valuation method
- Limits individual stock weight to 10% of portfolio
- Suggests 8-15 stocks for diversification
- Tight trailing stop (8-10%) to protect gains
- Strictest conviction checklist (10+ items, all must pass)

### Moderate

Best for: Investors comfortable with a balanced approach, mixing blue chips and mid-caps.

- Broader stock universe (LQ45 mix of blue chips and mid-caps)
- Moderate discounts (15-30% margin of safety)
- Uses Graham Number and PBV bands for valuation
- Limits individual stock weight to 20%
- Suggests 5-10 stocks
- Moderate trailing stop (12-15%)
- Standard conviction checklist (8+ items)

### Aggressive

Best for: Experienced investors willing to accept higher risk for potentially higher returns.

- Widest universe including mid-caps (IDX80+)
- Minimal discounts (0-15% margin of safety)
- Uses PBV bands and forward PER for valuation
- Limits individual stock weight to 35%
- Concentrated portfolio (3-7 stocks)
- Wide trailing stop (18-25%) to ride momentum
- Streamlined conviction checklist (6+ items)

## Parameter Comparison

| Parameter | Conservative | Moderate | Aggressive |
|-----------|-------------|----------|------------|
| **Margin of Safety** | 30-50% | 15-30% | 0-15% |
| **Stock Quality Filter** | Blue chips (IDX30) | Mix (LQ45) | Mid-cap included (IDX80+) |
| **Valuation Methods** | Graham Number | Graham + PBV band | PBV band + forward PER |
| **Max Single Stock Weight** | 10% | 20% | 35% |
| **Suggested Portfolio Size** | 8-15 stocks | 5-10 stocks | 3-7 stocks |
| **Trailing Stop % (Value)** | 8-10% | 12-15% | 18-25% |
| **Checklist Strictness** | 10+ items, all pass | 8+ items | 6+ items |

### Dividend-Specific Parameters

| Parameter | Conservative | Moderate | Aggressive |
|-----------|-------------|----------|------------|
| **Min Dividend Yield** | 5% | 3% | 2% |
| **Max Payout Ratio** | 60% | 75% | 90% |
| **Min Dividend Consistency** | 5 years | 3 years | 1 year |

## How Risk Profile Affects the App

### Entry Decisions

The risk profile determines how cheap a stock needs to be before the app considers it a buy. A Conservative profile requires a 30-50% discount to intrinsic value, meaning fewer stocks qualify but the ones that do have a larger safety cushion.

### Position Sizing

When the app suggests how many lots to buy, it enforces the max single stock weight. If buying 5 lots of BBRI would push BBRI to 25% of a Conservative portfolio (max 10%), the app warns you or suggests buying fewer lots.

### Exit Signals

Conservative portfolios use tighter trailing stops (8-10%), exiting positions faster when prices decline. Aggressive portfolios allow wider swings (18-25%), staying in positions through normal volatility.

### Conviction Checklists

More items means more things to verify before acting. A Conservative checklist might include 12 checks covering valuation, fundamentals, sector health, macro conditions, and personal verification. An Aggressive checklist might have 6 focused checks.

### Crash Playbook

During market crashes, the risk profile determines how much capital the playbook suggests deploying at each level and how deep the discount needs to be before triggering a buy signal.

## Choosing a Risk Profile

### Start Conservative If

- You are new to IDX investing
- You prefer stability over growth
- You want the app to be strict about what qualifies as a good buy
- You have a lower risk tolerance

### Start Moderate If

- You have some investing experience
- You want a balance of opportunities and discipline
- You are comfortable with mid-cap stocks alongside blue chips

### Start Aggressive If

- You are experienced and understand the risks
- You want a concentrated portfolio of your highest-conviction picks
- You are comfortable with larger drawdowns
- You actively follow market and company news

## Changing Risk Profiles

You can change a portfolio's risk profile at any time. When you do:

- Valuation zones are recalculated for all holdings
- Trailing stop levels adjust to the new percentages
- Checklist requirements update
- Existing holdings are not automatically sold, but their action signals may change

For example, switching from Aggressive to Conservative might flag previously "fair value" stocks as "overvalued" because the higher margin of safety now requires deeper discounts.

## Mix and Match

Risk profiles are set **per portfolio**, not globally. You can have:

- An Aggressive Value portfolio at Ajaib (concentrated high-conviction picks)
- A Conservative Dividend portfolio at IPOT (stable blue-chip income)

This flexibility lets you match your risk tolerance to your investment thesis for each portfolio independently.
