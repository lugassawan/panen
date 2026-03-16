"""Apps Script client for reading from and writing to Google Sheets."""

import logging
import os
import time
from typing import Any

import requests

logger = logging.getLogger(__name__)

_MAX_RETRIES = 3
_TIMEOUT_SECS = 30


def _get_config() -> tuple[str, str]:
    """Read Apps Script URL and token from environment variables."""
    url = os.environ.get("APPSCRIPT_URL", "")
    token = os.environ.get("APPSCRIPT_TOKEN", "")
    if not url:
        raise EnvironmentError("APPSCRIPT_URL environment variable is not set")
    if not token:
        raise EnvironmentError("APPSCRIPT_TOKEN environment variable is not set")
    return url, token


def _request_with_retries(
    method: str,
    url: str,
    headers: dict[str, str],
    **kwargs: Any,
) -> dict[str, Any] | None:
    """Make an HTTP request with exponential backoff retries."""
    for attempt in range(_MAX_RETRIES):
        try:
            resp = requests.request(
                method, url, headers=headers, timeout=_TIMEOUT_SECS, **kwargs
            )
            resp.raise_for_status()
            return resp.json()
        except requests.RequestException as exc:
            wait = 2**attempt
            logger.warning(
                "Apps Script request failed (attempt %d/%d): %s — retrying in %ds",
                attempt + 1,
                _MAX_RETRIES,
                exc,
                wait,
            )
            if attempt < _MAX_RETRIES - 1:
                time.sleep(wait)

    logger.error("Apps Script request failed after %d retries", _MAX_RETRIES)
    return None


def post_data(sheet: str, rows: list[dict[str, Any]]) -> bool:
    """POST data rows to an Apps Script upsert endpoint.

    Returns True on success, False on failure.
    """
    url, token = _get_config()
    headers = {
        "Content-Type": "application/json",
    }
    payload = {
        "action": "upsert",
        "sheet": sheet,
        "rows": rows,
    }

    # Apps Script doPost cannot read HTTP headers, so pass token as query param.
    post_url = f"{url}?token={token}"
    result = _request_with_retries("POST", post_url, headers=headers, json=payload)
    if result is None:
        return False

    if result.get("status") != "ok":
        logger.error("Apps Script upsert error: %s", result.get("error", "unknown"))
        return False

    logger.info("Posted %d rows to sheet '%s'", len(rows), sheet)
    return True


def get_data(action: str, ticker: str | None = None) -> list[dict[str, Any]]:
    """GET data from an Apps Script read endpoint.

    Returns a list of row dicts, or an empty list on failure.
    """
    url, token = _get_config()
    headers: dict[str, str] = {}
    params: dict[str, str] = {"action": action, "token": token}
    if ticker:
        params["ticker"] = ticker

    result = _request_with_retries("GET", url, headers=headers, params=params)
    if result is None:
        return []

    return result.get("data", [])
