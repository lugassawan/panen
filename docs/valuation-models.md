# Valuation Models

Panen uses multiple valuation methods to estimate fair value and generate entry/exit signals. The methods used depend on the portfolio's risk profile.

## Graham Number

The Graham Number is an intrinsic value estimate based on Benjamin Graham's formula:

```
Graham Number = sqrt(22.5 * EPS * BVPS)
```

Where:
- **EPS** = Earnings Per Share (trailing twelve months)
- **BVPS** = Book Value Per Share

The constant 22.5 comes from Graham's criteria that a stock should have a PER no higher than 15 and a PBV no higher than 1.5 (15 x 1.5 = 22.5).

### Interpretation

- **Price below Graham Number** -- The stock may be undervalued
- **Price above Graham Number** -- The stock may be overvalued
- The Graham Number works best for profitable, asset-heavy companies

### Limitations

- Not suitable for companies with negative earnings or book value
- Does not account for growth expectations
- Works best for stable, established companies (blue chips)

## PBV Bands (Price-to-Book Value)

PBV band analysis compares the current PBV ratio against its historical 5-year range to determine relative cheapness.

### How It Works

1. The app collects 5 years of PBV data for the stock
2. It calculates statistical bands: minimum, -1 standard deviation, mean, +1 standard deviation, maximum
3. The current PBV is plotted against these bands

### Valuation Zones

| Zone | PBV Position | Interpretation |
|------|-------------|----------------|
| Undervalued | Below -1 SD | Historically cheap |
| Fair value | Between -1 SD and +1 SD | Normal trading range |
| Overvalued | Above +1 SD | Historically expensive |

### Entry and Exit Levels

- **Entry target** -- PBV at -1 SD (or lower, depending on margin of safety)
- **Exit target** -- PBV at +1 SD or above the mean, depending on risk profile

## PER Bands (Price-to-Earnings Ratio)

PER band analysis works the same way as PBV bands but uses the Price-to-Earnings Ratio.

### How It Works

1. Collect 5 years of PER data
2. Calculate statistical bands (min, -1 SD, mean, +1 SD, max)
3. Plot current PER against the bands

### When to Use PER vs. PBV

| Scenario | Preferred Method |
|----------|-----------------|
| Asset-heavy companies (banks, property) | PBV bands |
| Earnings-driven companies (consumer, tech) | PER bands |
| Stable blue chips | Both or Graham Number |
| High-growth companies | Forward PER (Aggressive risk profile) |

## Margin of Safety

Margin of safety is a buffer applied below the calculated fair value to account for uncertainty. It reduces the entry price target to provide a safety cushion.

```
Entry Price = Fair Value * (1 - Margin of Safety %)
```

### Margin of Safety by Risk Profile

| Risk Profile | Margin of Safety | Effect |
|-------------|-----------------|--------|
| Conservative | 30-50% | Only buy at a deep discount. Fewer opportunities but lower risk |
| Moderate | 15-30% | Balanced approach. Reasonable discount with more opportunities |
| Aggressive | 0-15% | Minimal discount required. More opportunities but higher risk |

### Example

If a stock's Graham Number is Rp 5,000:

| Risk Profile | Margin of Safety | Entry Target |
|-------------|-----------------|--------------|
| Conservative (40%) | 40% | Rp 3,000 |
| Moderate (25%) | 25% | Rp 3,750 |
| Aggressive (10%) | 10% | Rp 4,500 |

## Valuation Zones

The app assigns each stock a valuation zone based on all applicable methods:

| Zone | Meaning | Color |
|------|---------|-------|
| Undervalued | Price below fair value with margin of safety | Green |
| Fair Value | Price near calculated fair value | Yellow |
| Overvalued | Price above fair value | Red |

The specific thresholds depend on the risk profile and the valuation methods enabled.

## Trailing Stop (Value Mode Only)

The trailing stop is a **suggestion**, not an automated trading tool. It tracks a percentage below the highest price since purchase and updates on each data refresh.

### How It Works

1. The app records the highest price since you bought the stock
2. The trailing stop level = peak price * (1 - trailing stop %)
3. If the current price falls below the trailing stop, the app flags it

### Trailing Stop by Risk Profile

| Risk Profile | Trailing Stop % | Behavior |
|-------------|----------------|----------|
| Conservative | 8-10% | Tight stop, exits early on declines |
| Moderate | 12-15% | Balanced approach |
| Aggressive | 18-25% | Wide stop, tolerates larger swings |

### Display Example

```
BBRI -- Exit Strategy
├── Fixed exit:     Rp 6,200 (upper PBV band)
├── Trailing stop:  Rp 5,220 (10% below peak of Rp 5,800)
└── Fundamental:    Sell if ROE drops below 15%
```

The trailing stop is shown in Value Mode only. Dividend Mode uses fundamental deterioration and dividend cuts as exit signals instead.

## Which Methods Apply Per Risk Profile

| Risk Profile | Valuation Methods |
|-------------|-------------------|
| Conservative | Graham Number (primary) |
| Moderate | Graham Number + PBV bands |
| Aggressive | PBV bands + forward PER |

See [Risk Profiles](risk-profiles.md) for the full parameter comparison.

## Fundamental Change Detection

The app monitors quarterly financial data and flags significant changes:

| Severity | Example | Effect |
|----------|---------|--------|
| Minor | ROE dropped 5% | Note in stock detail |
| Warning | DER crossed above 1.0 | Yellow alert, checklist updated |
| Critical | ROE dropped 30%+, earnings collapsed | Red alert, exit target recalculated |

Critical changes cause conviction checklist auto-checks to fail, preventing you from acting on stale suggestions. This protects against buying or holding based on outdated fundamentals.
