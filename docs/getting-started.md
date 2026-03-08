# Getting Started

This guide walks you through launching Panen for the first time, setting up your brokerage accounts, creating portfolios, and adding your first stock.

## First Launch

When you open Panen for the first time, the app creates a local database at `~/.panen/data/default.db`. All your data lives on your device -- there are no accounts, no logins, and no data sent to any server.

You will see an empty dashboard prompting you to get started.

## Step 1: Create a Brokerage Account

A brokerage account in Panen represents your real-world brokerage relationship (e.g., Ajaib, IPOT, Stockbit Sekuritas). You need at least one to create portfolios.

1. Open **Settings** from the sidebar
2. Find the **Brokerage Accounts** section
3. Click **Add Brokerage Account**
4. Choose your broker:
   - **Broker picker** -- Select from the community-maintained list. Buy and sell fees are auto-filled from the latest data
   - **Manual entry** -- Enter a custom broker name and fee percentages
5. Review the fees and click **Save**

Fees are **required** -- there is no 0% default. Accurate fees ensure correct P/L calculations, break-even prices, and allocation suggestions throughout the app.

You can create multiple brokerage accounts if you use more than one broker. This mirrors the real-world constraint that each broker account is separate.

## Step 2: Create a Portfolio

Each brokerage account can have up to two portfolios: one **Value** portfolio and one **Dividend** portfolio.

1. Navigate to the **Portfolio** section in the sidebar
2. Click **Create Portfolio**
3. Select the brokerage account it belongs to
4. Choose the portfolio mode:
   - **Value Mode** -- Focus on capital growth via undervalued stocks
   - **Dividend Mode** -- Focus on passive income via dividend-paying stocks
5. Choose a **risk profile** (Conservative, Moderate, or Aggressive) -- see [Risk Profiles](risk-profiles.md) for details
6. Set your **initial capital** and optionally configure **monthly addition**
7. Select your **stock universe** (e.g., IDX30, LQ45, IDX80)
8. Click **Create**

If you are new to investing, start with a single Value portfolio using a Conservative or Moderate risk profile and the IDX30 universe. You can add a Dividend portfolio later when you are ready.

## Step 3: Look Up a Stock

Use the **Stock Lookup** page to search for any IDX stock by ticker (e.g., BBCA, BBRI, TLKM).

The stock detail page shows:

- **Current price** and 52-week range
- **Valuation metrics** -- PBV, PER, Graham Number, intrinsic value estimates
- **Valuation zone** -- Whether the stock is undervalued, fairly valued, or overvalued
- **Fundamentals** -- ROE, DER, EPS, BVPS
- **Dividend data** -- Yield, payout ratio, consistency, growth rate
- **Price history chart**

This is where you research before deciding to add a stock to your portfolio.

## Step 4: Add a Stock to Your Portfolio

Once you have found a stock you want to track:

1. Go to your portfolio's detail page
2. Use the **Add Holding** form
3. Enter the ticker, buy price, number of lots, and buy date
4. The app calculates fees and records the transaction

The stock now appears in your portfolio with its current valuation status, P/L, and suggested actions.

**Stock uniqueness rule:** A stock can only exist in one portfolio per brokerage account. If you try to add BBRI to your Dividend portfolio when it is already in your Value portfolio at the same broker, the app will offer to move it or suggest adding it through a different brokerage.

## Step 5: Explore Your Dashboard

The dashboard provides an overview of all your portfolios:

- Total portfolio value and overall P/L
- Per-portfolio summaries
- Stocks requiring attention (valuation changes, fundamental alerts)
- Upcoming dividend ex-dates (if you have a Dividend portfolio)

## Next Steps

- [Portfolio Management](portfolio-management.md) -- Learn about Value and Dividend mode workflows
- [Valuation Models](valuation-models.md) -- Understand how entry and exit prices are calculated
- [Risk Profiles](risk-profiles.md) -- Choose the right risk profile for your investment style
- [Crash Playbook](crash-playbook.md) -- Prepare for market downturns before they happen

## Settings Worth Configuring

- **Language** -- Switch between English and Bahasa Indonesia
- **Theme** -- Light, Dark, or System default
- **Monthly Addition** -- Set your payday date and monthly investment amount
- **Data Providers** -- View and manage data sources (Yahoo Finance, IDX)
- **Export/Import** -- Back up your data or migrate to another device

## Data and Privacy

- All data is stored locally in `~/.panen/`
- Daily backups are created automatically (7-day retention)
- No analytics, no telemetry, no tracking
- Export your data anytime for backup or migration
- Debug logs exclude all financial data
