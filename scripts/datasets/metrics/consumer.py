"""Consumer sector metric extractors: SSSG, GPM."""

from datetime import datetime, timezone
from typing import Any


def extract(ticker: str, financials: dict[str, Any]) -> list[dict[str, Any]]:
    """Extract consumer-specific metrics.

    GPM (Gross Profit Margin) is derivable from income statement data.
    SSSG (Same-Store Sales Growth) requires company-specific disclosures.
    """
    metrics: list[dict[str, Any]] = []

    # SSSG is always manual — not available in standard financials.
    metrics.append(
        {
            "ticker": ticker,
            "metric": "SSSG",
            "value": None,
            "source": "manual",
            "year": None,
            "quarter": None,
        }
    )

    # GPM: derive from income statement gross profit / revenue.
    for stmt in financials.get("incomeStatements", []):
        revenue = stmt.get("totalRevenue")
        gross_profit = stmt.get("grossProfit")
        if revenue and gross_profit and revenue != 0:
            gpm = gross_profit / revenue

            year, quarter = _parse_period(stmt.get("endDate"))
            metrics.append(
                {
                    "ticker": ticker,
                    "metric": "GPM",
                    "value": round(gpm, 4),
                    "source": "auto",
                    "year": year,
                    "quarter": quarter,
                }
            )

    return metrics


def _parse_period(end_date: int | None) -> tuple[int | None, int | None]:
    """Convert Unix timestamp to (year, quarter)."""
    if end_date is None:
        return None, None
    dt = datetime.fromtimestamp(end_date, tz=timezone.utc)
    quarter = (dt.month - 1) // 3 + 1
    return dt.year, quarter
