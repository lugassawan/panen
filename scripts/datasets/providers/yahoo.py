"""Yahoo Finance data fetcher for IDX stocks."""

import logging
import time
from typing import Any

import requests

logger = logging.getLogger(__name__)

_BASE_URL = "https://query1.finance.yahoo.com/v10/finance/quoteSummary"

_MODULES = [
    "incomeStatementHistory",
    "incomeStatementHistoryQuarterly",
    "balanceSheetHistory",
    "balanceSheetHistoryQuarterly",
    "defaultKeyStatistics",
    "financialData",
]

_MAX_RETRIES = 3
_RATE_LIMIT_SECS = 1.0

# Track last request time for rate limiting across calls.
_last_request_time: float = 0.0


def _to_yahoo_ticker(ticker: str) -> str:
    """Convert IDX ticker to Yahoo Finance format (append .JK suffix)."""
    return f"{ticker}.JK"


def _get_with_retries(url: str, params: dict[str, str]) -> dict[str, Any] | None:
    """GET request with exponential backoff retries and rate limiting."""
    global _last_request_time

    for attempt in range(_MAX_RETRIES):
        # Rate limiting: wait at least 1 second between requests.
        elapsed = time.monotonic() - _last_request_time
        if elapsed < _RATE_LIMIT_SECS:
            time.sleep(_RATE_LIMIT_SECS - elapsed)

        try:
            _last_request_time = time.monotonic()
            resp = requests.get(url, params=params, timeout=30)
            resp.raise_for_status()
            return resp.json()
        except requests.RequestException as exc:
            wait = 2**attempt
            logger.warning(
                "Yahoo Finance request failed (attempt %d/%d): %s — retrying in %ds",
                attempt + 1,
                _MAX_RETRIES,
                exc,
                wait,
            )
            time.sleep(wait)

    return None


def _safe_get(obj: Any, key: str) -> Any:
    """Safely extract a value from a Yahoo Finance data object.

    Yahoo wraps values in dicts like {"raw": 123, "fmt": "123"}.
    """
    if obj is None:
        return None
    val = obj.get(key)
    if isinstance(val, dict):
        return val.get("raw")
    return val


def _extract_income_statements(result: dict[str, Any]) -> list[dict[str, Any]]:
    """Extract annual and quarterly income statement data."""
    statements: list[dict[str, Any]] = []

    for module_key in ("incomeStatementHistory", "incomeStatementHistoryQuarterly"):
        module = result.get(module_key, {})
        period = "annual" if "Quarterly" not in module_key else "quarterly"

        for stmt in module.get("incomeStatementHistory", []):
            statements.append(
                {
                    "period": period,
                    "endDate": _safe_get(stmt, "endDate"),
                    "totalRevenue": _safe_get(stmt, "totalRevenue"),
                    "netIncome": _safe_get(stmt, "netIncome"),
                    "operatingIncome": _safe_get(stmt, "operatingIncome"),
                    "costOfRevenue": _safe_get(stmt, "costOfRevenue"),
                    "grossProfit": _safe_get(stmt, "grossProfit"),
                }
            )

    return statements


def _extract_balance_sheets(result: dict[str, Any]) -> list[dict[str, Any]]:
    """Extract annual and quarterly balance sheet data."""
    sheets: list[dict[str, Any]] = []

    for module_key in ("balanceSheetHistory", "balanceSheetHistoryQuarterly"):
        module = result.get(module_key, {})
        period = "annual" if "Quarterly" not in module_key else "quarterly"

        for stmt in module.get("balanceSheetStatements", []):
            sheets.append(
                {
                    "period": period,
                    "endDate": _safe_get(stmt, "endDate"),
                    "totalAssets": _safe_get(stmt, "totalAssets"),
                    "totalStockholderEquity": _safe_get(
                        stmt, "totalStockholderEquity"
                    ),
                    "totalDebt": _safe_get(stmt, "longTermDebt"),
                    "totalCurrentAssets": _safe_get(stmt, "totalCurrentAssets"),
                    "totalCurrentLiabilities": _safe_get(
                        stmt, "totalCurrentLiabilities"
                    ),
                }
            )

    return sheets


def _extract_key_statistics(result: dict[str, Any]) -> dict[str, Any]:
    """Extract key statistics (PER, PBV, etc.)."""
    stats = result.get("defaultKeyStatistics", {})
    return {
        "trailingPE": _safe_get(stats, "trailingPE") or _safe_get(stats, "trailingPe"),
        "forwardPE": _safe_get(stats, "forwardPE") or _safe_get(stats, "forwardPe"),
        "priceToBook": _safe_get(stats, "priceToBook"),
        "enterpriseValue": _safe_get(stats, "enterpriseValue"),
    }


def _extract_financial_data(result: dict[str, Any]) -> dict[str, Any]:
    """Extract financial data (ROE, ROA, ratios)."""
    fin = result.get("financialData", {})
    return {
        "returnOnEquity": _safe_get(fin, "returnOnEquity"),
        "returnOnAssets": _safe_get(fin, "returnOnAssets"),
        "currentRatio": _safe_get(fin, "currentRatio"),
        "debtToEquity": _safe_get(fin, "debtToEquity"),
        "revenueGrowth": _safe_get(fin, "revenueGrowth"),
        "earningsGrowth": _safe_get(fin, "earningsGrowth"),
        "currentPrice": _safe_get(fin, "currentPrice"),
    }


def fetch(ticker: str) -> dict[str, Any] | None:
    """Fetch financial data for an IDX ticker from Yahoo Finance.

    Returns a dict with income statements, balance sheets, key statistics,
    and financial data, or None on failure.
    """
    yahoo_ticker = _to_yahoo_ticker(ticker)
    params = {
        "modules": ",".join(_MODULES),
    }
    url = f"{_BASE_URL}/{yahoo_ticker}"

    data = _get_with_retries(url, params)
    if data is None:
        logger.error("Failed to fetch data for %s after %d retries", ticker, _MAX_RETRIES)
        return None

    try:
        result = data["quoteSummary"]["result"][0]
    except (KeyError, IndexError, TypeError):
        logger.error("Unexpected response structure for %s", ticker)
        return None

    return {
        "ticker": ticker,
        "incomeStatements": _extract_income_statements(result),
        "balanceSheets": _extract_balance_sheets(result),
        "keyStatistics": _extract_key_statistics(result),
        "financialData": _extract_financial_data(result),
    }
