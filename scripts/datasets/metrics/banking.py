"""Banking sector metric extractors: CAR, NIM, NPL, LDR."""

from typing import Any


def extract(ticker: str, financials: dict[str, Any]) -> list[dict[str, Any]]:
    """Extract banking-specific metrics.

    CAR, NIM, NPL, and LDR are not derivable from standard financial
    statements — they require regulatory filings. All are marked as manual.
    """
    metrics: list[dict[str, Any]] = []

    for metric_name in ("CAR", "NIM", "NPL", "LDR"):
        metrics.append(
            {
                "ticker": ticker,
                "metric": metric_name,
                "value": None,
                "source": "manual",
                "year": None,
                "quarter": None,
            }
        )

    return metrics
