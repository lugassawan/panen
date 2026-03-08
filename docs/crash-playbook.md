# Crash Playbook

The crash playbook is Panen's approach to market downturns: **prepare during calm markets, execute during crashes**. Because the app scrapes data on intervals (every 3-12 hours), it is always hours behind the real market. The strategy is therefore proactive, not reactive.

## Philosophy

The app never says "buy now" during a crash. It says: "Your pre-planned level was hit. Here is your checklist. Verify in your broker. Decide."

This approach works because:

1. **Decisions made in calm conditions are better** -- Panic selling and FOMO buying are the biggest destroyers of returns
2. **The app cannot react in real time** -- Data is delayed by hours, so split-second calls are impossible
3. **Pre-commitment builds conviction** -- When you have already decided what to do at each price level, you act with confidence instead of fear

## Market Conditions

The app monitors the Jakarta Composite Index (IHSG) to classify market conditions:

| Condition | IHSG Change | Scrape Frequency | App Behavior |
|-----------|-------------|-----------------|--------------|
| Normal | Business as usual | Every 12-24 hours | Standard operation |
| Elevated | -5% to -10% from peak | Every 6 hours | "Market dropped. Your playbook is ready." |
| Correction | -10% to -20% from peak | Every 3-4 hours | "Correction detected. Review your playbook." |
| Crash | > -20% from peak | Every 3-4 hours | "Crash detected. X stocks entered buy zones." |
| Recovery | Up > 10% from crash bottom | Returns to normal | Track recovery progress |

## Building a Playbook

Build your crash playbook **before** a crash happens -- ideally when markets are calm and you can think clearly.

### Step 1: Select Watchlist Stocks

Choose the stocks you would want to buy if they became significantly cheaper. These should be fundamentally strong companies you have already researched.

### Step 2: Define Entry Levels

For each stock, set pre-calculated entry levels at increasing discounts. The app helps calculate these based on your valuation models and risk profile.

Example for BBRI:

```
BBRI -- Pre-Planned Crash Response
├── Level 1: Rp 4,300 (normal dip)      → Deploy 30% of crash capital
├── Level 2: Rp 3,800 (crash territory) → Deploy 40% of crash capital
├── Level 3: Rp 3,200 (extreme)         → Deploy 30% of crash capital
```

### Step 3: Allocate Crash Capital

Decide how much capital you are willing to deploy during a crash. This can come from your war chest (Value Mode) or be additional capital you set aside.

```
Pre-committed crash capital: Rp 5,000,000
├── BBRI: Rp 2,500,000 (50%)
├── BBCA: Rp 1,500,000 (30%)
└── TLKM: Rp 1,000,000 (20%)
```

### Step 4: Set Lot Sizes

The app converts your capital allocation into specific lot counts at each price level, accounting for broker fees.

## During a Crash

When a market correction or crash is detected, the app:

1. **Highlights which pre-planned levels were hit** -- Shows which of your watchlist stocks have dropped to (or below) your entry targets
2. **Runs fundamental health checks** -- Verifies that the stock's fundamentals have not deteriorated (a cheap stock with collapsing earnings is a trap, not an opportunity)
3. **Updates the conviction checklist** -- Auto-checks are refreshed with latest data; you still complete manual checks
4. **Shows the playbook action** -- "Your Level 2 for BBRI was hit. Checklist ready. Deploy 4 lots per your plan."

### Crash Diagnostics

The app helps distinguish genuine opportunities from falling knives by checking:

- Did the broad market crash, or just this stock?
- Is there company-specific bad news?
- Are fundamentals still healthy (ROE, DER, earnings)?
- Is the price below your pre-calculated entry target?

If the stock dropped due to market-wide panic but fundamentals are intact, the opportunity is stronger. If the stock dropped due to fundamental deterioration, the app flags this and the conviction checklist's auto-checks will fail.

## After a Crash

Once you have made purchases during a crash:

1. The app records your actual buy transactions
2. Your portfolio updates with new average prices
3. The playbook resets for future events
4. Recovery tracking shows how your crash purchases perform

## Playbook and Risk Profiles

Your risk profile influences the playbook:

| Aspect | Conservative | Moderate | Aggressive |
|--------|-------------|----------|------------|
| Entry level depth | Deeper discounts required | Balanced | Shallower entries |
| Capital deployment | Gradual (30/40/30 split) | Balanced | Can be front-loaded |
| Fundamental strictness | All checks must pass | Standard | Fewer required checks |
| Stock quality | Blue chips only | Mix | Includes mid-caps |

## Best Practices

1. **Build playbooks when calm** -- Do not create a playbook while markets are falling. The whole point is pre-commitment
2. **Review quarterly** -- Update entry levels after each earnings season as fundamentals change
3. **Do not over-commit** -- Keep enough cash for living expenses and emergency funds outside the playbook
4. **Verify in your broker** -- The app suggests; you execute in your actual brokerage platform
5. **Trust the process** -- If you built the playbook with sound analysis, trust it during the emotional chaos of a crash
6. **Document your reasoning** -- Use the manual checklist items to record why you believe the thesis holds

## Example Scenario

**Before crash (January, IHSG at 7,500):**

You research BBRI and determine Graham Number is Rp 5,200. With a 25% margin of safety (Moderate profile), your entry target is Rp 3,900. You set up three levels in your playbook.

**During crash (March, IHSG drops to 6,000, -20%):**

- BBRI drops to Rp 4,100 -- Level 1 not yet hit (target was Rp 4,300)
- BBRI drops to Rp 3,800 -- Level 2 hit
- App checks fundamentals: ROE still 18%, DER still 0.8, earnings stable
- Conviction checklist: 8/8 auto-checks pass
- You complete manual checks: no fraud news, sector outlook stable
- App suggests: "Deploy 40% crash capital. Buy 5 lots at Rp 3,800. Cost with fees: Rp 1,912,800."
- You verify in your broker and execute the trade
- You confirm the purchase in the app

**After crash (June, IHSG recovers to 7,200):**

BBRI recovers to Rp 5,000. Your crash purchase shows +31.6% unrealized gain. The playbook resets, and you can build a new one for the next opportunity.
