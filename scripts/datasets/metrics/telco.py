"""Telco sector metric extractors: ARPU."""

from typing import Any


def extract(ticker: str, financials: dict[str, Any]) -> list[dict[str, Any]]:
    """Extract telco-specific metrics.

    ARPU (Average Revenue Per User) requires operator-specific disclosures
    and is not derivable from standard financial statements.
    """
    return [
        {
            "ticker": ticker,
            "metric": "ARPU",
            "value": None,
            "source": "manual",
            "year": None,
            "quarter": None,
        }
    ]
