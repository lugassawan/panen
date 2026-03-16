"""IDX listing API client for discovering active tickers."""

import logging
from dataclasses import dataclass

import requests

logger = logging.getLogger(__name__)

IDX_STOCK_DATA_URL = "https://www.idx.co.id/primary/StockData/GetStockData"

# IDX API requires a browser-like user agent to avoid 403 responses.
_HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
        "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
    ),
    "Accept": "application/json",
    "Referer": "https://www.idx.co.id/en/market-data/stocks/",
}


@dataclass
class IDXStock:
    """Represents a stock listed on IDX."""

    code: str
    name: str
    sector: str
    board: str
    status: str


def fetch_active_stocks() -> list[IDXStock]:
    """Fetch all active (non-suspended) stocks from IDX API.

    Returns an empty list and logs a warning on API failure.
    """
    try:
        params = {
            "start": 0,
            "length": 9999,
        }
        resp = requests.get(
            IDX_STOCK_DATA_URL,
            params=params,
            headers=_HEADERS,
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()

        stocks: list[IDXStock] = []
        for item in data.get("data", []):
            status = item.get("Status", "")
            if status == "Suspend":
                continue
            stocks.append(
                IDXStock(
                    code=item.get("Code", ""),
                    name=item.get("Name", ""),
                    sector=item.get("SectorName", ""),
                    board=item.get("ListingBoard", ""),
                    status=status,
                )
            )
        logger.info("Fetched %d active stocks from IDX", len(stocks))
        return stocks

    except (requests.RequestException, ValueError, KeyError) as exc:
        logger.warning("IDX API fetch failed: %s", exc)
        return []
