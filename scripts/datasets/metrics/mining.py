"""Mining sector metric extractors: ASP, stripping ratio."""

from typing import Any


def extract(ticker: str, financials: dict[str, Any]) -> list[dict[str, Any]]:
    """Extract mining-specific metrics.

    ASP (Average Selling Price) and stripping ratio require company-specific
    operational data not available in standard financial statements.
    """
    return [
        {
            "ticker": ticker,
            "metric": "ASP",
            "value": None,
            "source": "manual",
            "year": None,
            "quarter": None,
        },
        {
            "ticker": ticker,
            "metric": "StrippingRatio",
            "value": None,
            "source": "manual",
            "year": None,
            "quarter": None,
        },
    ]
