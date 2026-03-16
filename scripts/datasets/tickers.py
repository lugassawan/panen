"""Dynamic ticker list from IDX API with fallback to static config."""

import json
import logging
from dataclasses import dataclass
from pathlib import Path

from providers.idx_listing import fetch_active_stocks

logger = logging.getLogger(__name__)

# Path to sectors.json relative to repo root.
_REPO_ROOT = Path(__file__).resolve().parent.parent.parent
_SECTORS_JSON = _REPO_ROOT / "configs" / "sectors.json"


@dataclass
class TickerInfo:
    """Metadata for a listed stock."""

    name: str
    sector: str
    board: str
    status: str


def get_active_tickers() -> dict[str, TickerInfo]:
    """Get active tickers, trying IDX API first then falling back to sectors.json.

    Returns a dict mapping ticker code to TickerInfo.
    """
    tickers = _fetch_from_idx()
    if tickers:
        return tickers

    logger.info("Falling back to static ticker list from %s", _SECTORS_JSON)
    return _load_from_sectors_json()


def _fetch_from_idx() -> dict[str, TickerInfo]:
    """Fetch active tickers from IDX API."""
    stocks = fetch_active_stocks()
    if not stocks:
        return {}

    return {
        stock.code: TickerInfo(
            name=stock.name,
            sector=stock.sector,
            board=stock.board,
            status=stock.status,
        )
        for stock in stocks
        if stock.code
    }


def _load_from_sectors_json() -> dict[str, TickerInfo]:
    """Load tickers from the static sectors.json config file."""
    try:
        data = json.loads(_SECTORS_JSON.read_text(encoding="utf-8"))
    except (FileNotFoundError, json.JSONDecodeError) as exc:
        logger.error("Failed to load %s: %s", _SECTORS_JSON, exc)
        return {}

    return {
        ticker: TickerInfo(
            name=ticker,
            sector=sector,
            board="",
            status="Active",
        )
        for ticker, sector in data.items()
    }
