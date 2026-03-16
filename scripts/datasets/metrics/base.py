"""Base metric definitions and sector-to-metric mapping."""

from dataclasses import dataclass


@dataclass
class MetricDefinition:
    """Defines a sector-specific financial metric."""

    name: str
    source: str  # "auto" (derivable from financials) or "manual" (requires manual input)
    description: str


# Mapping of sector name to its specific metrics.
SECTOR_METRICS: dict[str, list[MetricDefinition]] = {
    "Banking": [
        MetricDefinition("CAR", "manual", "Capital Adequacy Ratio"),
        MetricDefinition("NIM", "manual", "Net Interest Margin"),
        MetricDefinition("NPL", "manual", "Non-Performing Loan ratio"),
        MetricDefinition("LDR", "manual", "Loan-to-Deposit Ratio"),
    ],
    "Consumer": [
        MetricDefinition("SSSG", "manual", "Same-Store Sales Growth"),
        MetricDefinition("GPM", "auto", "Gross Profit Margin (gross profit / revenue)"),
    ],
    "Telco": [
        MetricDefinition("ARPU", "manual", "Average Revenue Per User"),
    ],
    "Mining": [
        MetricDefinition("ASP", "manual", "Average Selling Price"),
        MetricDefinition("StrippingRatio", "manual", "Stripping ratio (overburden / ore)"),
    ],
}


def get_sector_metrics(sector: str) -> list[MetricDefinition]:
    """Return metric definitions for a sector, or an empty list if none defined."""
    return SECTOR_METRICS.get(sector, [])
