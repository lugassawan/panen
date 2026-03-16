/**
 * Apps Script web app for reading/writing financial datasets in Google Sheets.
 *
 * Sheets: financials, sector_metrics, definitions, tickers
 *
 * doGet — public read endpoints
 * doPost — token-protected write endpoints
 */

// ─── Configuration ──────────────────────────────────────────────────────────

const SHEET_NAMES = ["financials", "sector_metrics", "definitions", "tickers"];

// Composite keys used for upsert deduplication per sheet.
const UPSERT_KEYS = {
  financials: ["ticker", "statementType", "period", "endDate"],
  sector_metrics: ["ticker", "metric", "year", "quarter"],
  tickers: ["ticker"],
};

// ─── Read Endpoints (doGet) ─────────────────────────────────────────────────

function doGet(e) {
  try {
    const action = (e.parameter.action || "").toLowerCase();
    const ticker = e.parameter.ticker || null;

    switch (action) {
      case "financials":
        return jsonResponse(getSheetData("financials", ticker));
      case "sectormetrics":
        return jsonResponse(getSheetData("sector_metrics", ticker));
      case "definitions":
        return jsonResponse(getSheetData("definitions", null));
      case "tickers":
        return jsonResponse(getSheetData("tickers", null));
      default:
        return jsonResponse({ error: "Unknown action: " + action }, 400);
    }
  } catch (err) {
    return jsonResponse({ error: err.message }, 500);
  }
}

// ─── Write Endpoints (doPost) ───────────────────────────────────────────────

function doPost(e) {
  try {
    if (!authorizeRequest(e)) {
      return jsonResponse({ error: "Unauthorized" }, 401);
    }

    const body = JSON.parse(e.postData.contents);
    const action = (body.action || "").toLowerCase();
    const sheetName = body.sheet || "";
    const rows = body.rows || [];

    if (action !== "upsert") {
      return jsonResponse({ error: "Unknown action: " + action }, 400);
    }

    if (SHEET_NAMES.indexOf(sheetName) === -1) {
      return jsonResponse({ error: "Invalid sheet: " + sheetName }, 400);
    }

    if (!Array.isArray(rows) || rows.length === 0) {
      return jsonResponse({ error: "No rows provided" }, 400);
    }

    const count = upsertRows(sheetName, rows);
    return jsonResponse({ status: "ok", upserted: count });
  } catch (err) {
    return jsonResponse({ error: err.message }, 500);
  }
}

// ─── Authorization ──────────────────────────────────────────────────────────

function authorizeRequest(e) {
  const props = PropertiesService.getScriptProperties();
  const expectedToken = props.getProperty("APPSCRIPT_TOKEN");
  if (!expectedToken) {
    return false;
  }

  const authHeader = e.parameter.token || "";
  if (authHeader === expectedToken) {
    return true;
  }

  // Also check Authorization header via custom parameter workaround.
  // Apps Script doPost doesn't expose headers directly, so the Python
  // client sends the token as a query parameter.
  return false;
}

// ─── Sheet Data Helpers ─────────────────────────────────────────────────────

function getSheetData(sheetName, ticker) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sheet = ss.getSheetByName(sheetName);
  if (!sheet) {
    return { version: 1, updatedAt: new Date().toISOString(), data: [] };
  }

  const data = sheet.getDataRange().getValues();
  if (data.length < 2) {
    return { version: 1, updatedAt: new Date().toISOString(), data: [] };
  }

  const headers = data[0];
  const tickerCol = headers.indexOf("ticker");
  const rows = [];

  for (let i = 1; i < data.length; i++) {
    if (ticker && tickerCol >= 0 && data[i][tickerCol] !== ticker) {
      continue;
    }

    const row = {};
    for (let j = 0; j < headers.length; j++) {
      row[headers[j]] = data[i][j];
    }
    rows.push(row);
  }

  return {
    version: 1,
    updatedAt: new Date().toISOString(),
    data: rows,
  };
}

function upsertRows(sheetName, rows) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  let sheet = ss.getSheetByName(sheetName);

  // Auto-create sheet with headers from first row if it doesn't exist.
  if (!sheet) {
    sheet = ss.insertSheet(sheetName);
    const headers = Object.keys(rows[0]);
    if (headers.indexOf("updatedAt") === -1) {
      headers.push("updatedAt");
    }
    sheet.appendRow(headers);
  }

  const data = sheet.getDataRange().getValues();
  const headers = data[0];
  const keys = UPSERT_KEYS[sheetName] || ["ticker"];
  const updatedAtCol = headers.indexOf("updatedAt");
  const now = new Date().toISOString();

  // Build index of existing rows by composite key.
  const existingIndex = {};
  for (let i = 1; i < data.length; i++) {
    const keyVal = keys.map(function (k) {
      const col = headers.indexOf(k);
      return col >= 0 ? String(data[i][col]) : "";
    }).join("|");
    existingIndex[keyVal] = i; // 0-based in data array, row i+1 in sheet
  }

  let upsertCount = 0;

  for (let r = 0; r < rows.length; r++) {
    const row = rows[r];
    const keyVal = keys.map(function (k) {
      return String(row[k] || "");
    }).join("|");

    // Build the row values in header order.
    const rowValues = headers.map(function (h) {
      if (h === "updatedAt") {
        return now;
      }
      return row[h] !== undefined ? row[h] : "";
    });

    if (existingIndex.hasOwnProperty(keyVal)) {
      // Update existing row.
      const sheetRow = existingIndex[keyVal] + 1; // 1-based sheet row
      sheet.getRange(sheetRow, 1, 1, rowValues.length).setValues([rowValues]);
    } else {
      // Append new row.
      sheet.appendRow(rowValues);
    }

    upsertCount++;
  }

  return upsertCount;
}

// ─── Utilities ──────────────────────────────────────────────────────────────

function jsonResponse(data, statusCode) {
  // Apps Script always returns 200 for doGet/doPost, but we include
  // the logical status in the response body for client-side handling.
  const output = ContentService.createTextOutput(JSON.stringify(data));
  output.setMimeType(ContentService.MimeType.JSON);
  return output;
}
