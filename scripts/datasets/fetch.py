"""Entry point for the financial data pipeline.

Orchestrates ticker discovery, Yahoo Finance data fetching, sector metric
extraction, and pushing results to Google Sheets via Apps Script.
"""

import logging
import sys
from typing import Any

import appscript
from metrics import banking, consumer, mining, telco
from providers import yahoo
from tickers import TickerInfo, get_active_tickers

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)

# Map sector names to their metric extractor modules.
_SECTOR_MODULES: dict[str, Any] = {
    "Banking": banking,
    "Consumer": consumer,
    "Telco": telco,
    "Mining": mining,
}


def _build_ticker_rows(tickers: dict[str, TickerInfo]) -> list[dict[str, Any]]:
    """Convert ticker info dict to rows suitable for the tickers sheet."""
    return [
        {
            "ticker": code,
            "name": info.name,
            "sector": info.sector,
            "board": info.board,
            "status": info.status,
        }
        for code, info in sorted(tickers.items())
    ]


def _flatten_financials(ticker: str, data: dict[str, Any]) -> list[dict[str, Any]]:
    """Flatten Yahoo Finance data into rows for the financials sheet."""
    rows: list[dict[str, Any]] = []

    # Key statistics and financial data as a single summary row.
    summary = {"ticker": ticker}
    summary.update(data.get("keyStatistics", {}))
    summary.update(data.get("financialData", {}))
    rows.append(summary)

    # Income statement rows.
    for stmt in data.get("incomeStatements", []):
        row = {"ticker": ticker, "statementType": "income"}
        row.update(stmt)
        rows.append(row)

    # Balance sheet rows.
    for sheet in data.get("balanceSheets", []):
        row = {"ticker": ticker, "statementType": "balanceSheet"}
        row.update(sheet)
        rows.append(row)

    return rows


def main() -> None:
    """Run the full data pipeline."""
    logger.info("Starting financial data pipeline")

    # 1. Get active tickers.
    tickers = get_active_tickers()
    if not tickers:
        logger.error("No tickers found — aborting")
        sys.exit(1)

    logger.info("Found %d active tickers", len(tickers))

    # Push ticker list to the tickers sheet.
    ticker_rows = _build_ticker_rows(tickers)
    appscript.post_data("tickers", ticker_rows)

    # 2. Fetch financials from Yahoo Finance.
    success = 0
    failed = 0
    skipped = 0
    all_financials: dict[str, dict[str, Any]] = {}

    for ticker in sorted(tickers):
        logger.info("Fetching %s...", ticker)
        data = yahoo.fetch(ticker)
        if data is None:
            failed += 1
            continue

        all_financials[ticker] = data
        flat_rows = _flatten_financials(ticker, data)
        if appscript.post_data("financials", flat_rows):
            success += 1
        else:
            failed += 1

    # 3. Compute derivable sector metrics.
    metrics_success = 0
    metrics_skipped = 0

    for ticker, info in sorted(tickers.items()):
        sector_module = _SECTOR_MODULES.get(info.sector)
        if sector_module is None:
            metrics_skipped += 1
            continue

        financials = all_financials.get(ticker)
        if financials is None:
            metrics_skipped += 1
            continue

        metrics = sector_module.extract(ticker, financials)
        if metrics:
            if appscript.post_data("sector_metrics", metrics):
                metrics_success += 1
            else:
                failed += 1
        else:
            metrics_skipped += 1

    # 4. Report summary.
    logger.info("=" * 50)
    logger.info("Pipeline complete")
    logger.info("  Financials — success: %d, failed: %d", success, failed)
    logger.info(
        "  Sector metrics — posted: %d, skipped: %d",
        metrics_success,
        metrics_skipped,
    )
    logger.info("  Total tickers: %d", len(tickers))
    logger.info("=" * 50)

    if failed > 0:
        sys.exit(1)


if __name__ == "__main__":
    main()
