import { describe, expect, it, vi } from "vitest";

const mockGetActiveAlerts = vi.fn(() => Promise.resolve([]));
const mockGetAlertsByTicker = vi.fn(() => Promise.resolve([]));
const mockAcknowledgeAlert = vi.fn(() => Promise.resolve());
const mockGetAlertCount = vi.fn(() => Promise.resolve(0));

vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetActiveAlerts: (...args: unknown[]) => mockGetActiveAlerts(...args),
  GetAlertsByTicker: (...args: unknown[]) => mockGetAlertsByTicker(...args),
  AcknowledgeAlert: (...args: unknown[]) => mockAcknowledgeAlert(...args),
  GetAlertCount: (...args: unknown[]) => mockGetAlertCount(...args),
}));

import { alerts } from "./alerts.svelte";

describe("alerts store", () => {
  it("starts with zero active count", () => {
    expect(alerts.activeCount).toBe(0);
  });

  it("starts with empty active alerts", () => {
    expect(alerts.activeAlerts).toEqual([]);
  });

  it("starts not loading", () => {
    expect(alerts.loading).toBe(false);
  });

  it("loadCount updates activeCount", async () => {
    mockGetAlertCount.mockResolvedValueOnce(5);
    await alerts.loadCount();
    expect(alerts.activeCount).toBe(5);
  });

  it("loadActiveAlerts populates alerts and count", async () => {
    const mockAlerts = [
      { id: "1", ticker: "BBCA", metric: "roe", severity: "CRITICAL", status: "ACTIVE" },
      { id: "2", ticker: "BMRI", metric: "der", severity: "WARNING", status: "ACTIVE" },
    ];
    mockGetActiveAlerts.mockResolvedValueOnce(mockAlerts);
    await alerts.loadActiveAlerts();
    expect(alerts.activeAlerts).toEqual(mockAlerts);
    expect(alerts.activeCount).toBe(2);
  });

  it("loadAlertsByTicker returns alerts for a ticker", async () => {
    const tickerAlerts = [
      { id: "1", ticker: "BBCA", metric: "roe", severity: "CRITICAL", status: "ACTIVE" },
    ];
    mockGetAlertsByTicker.mockResolvedValueOnce(tickerAlerts);
    const result = await alerts.loadAlertsByTicker("BBCA");
    expect(result).toEqual(tickerAlerts);
    expect(mockGetAlertsByTicker).toHaveBeenCalledWith("BBCA");
  });

  it("loadAlertsByTicker returns empty array on error", async () => {
    mockGetAlertsByTicker.mockRejectedValueOnce(new Error("fail"));
    const result = await alerts.loadAlertsByTicker("FAIL");
    expect(result).toEqual([]);
  });

  it("acknowledgeAlert calls backend and reloads", async () => {
    mockGetActiveAlerts.mockResolvedValueOnce([]);
    await alerts.acknowledgeAlert("alert-1");
    expect(mockAcknowledgeAlert).toHaveBeenCalledWith("alert-1");
    expect(mockGetActiveAlerts).toHaveBeenCalled();
  });
});
