import { describe, expect, it } from "vitest";
import {
  calcPL,
  calcPLAbsolute,
  currentValue,
  holdingWeights,
  overallPL,
  sectorWeights,
  totalInvested,
} from "./portfolio";
import type { HoldingDetailResponse } from "./types";

function holding(
  overrides: Partial<HoldingDetailResponse> & Pick<HoldingDetailResponse, "avgBuyPrice" | "lots">,
): HoldingDetailResponse {
  return {
    id: "h1",
    ticker: "BBCA",
    ...overrides,
  };
}

describe("calcPL", () => {
  it("returns null when currentPrice is undefined", () => {
    expect(calcPL(undefined, 1000)).toBeNull();
  });

  it("calculates positive P&L", () => {
    expect(calcPL(1200, 1000)).toBeCloseTo(20);
  });

  it("calculates negative P&L", () => {
    expect(calcPL(800, 1000)).toBeCloseTo(-20);
  });

  it("returns 0 when prices are equal", () => {
    expect(calcPL(1000, 1000)).toBe(0);
  });
});

describe("totalInvested", () => {
  it("returns 0 for empty holdings", () => {
    expect(totalInvested([])).toBe(0);
  });

  it("sums avgBuyPrice * lots * 100", () => {
    const holdings = [
      holding({ avgBuyPrice: 1000, lots: 10 }),
      holding({ avgBuyPrice: 2000, lots: 5 }),
    ];
    // 1000*10*100 + 2000*5*100 = 1_000_000 + 1_000_000 = 2_000_000
    expect(totalInvested(holdings)).toBe(2_000_000);
  });
});

describe("currentValue", () => {
  it("uses currentPrice when available", () => {
    const holdings = [holding({ avgBuyPrice: 1000, lots: 10, currentPrice: 1200 })];
    // 1200*10*100 = 1_200_000
    expect(currentValue(holdings)).toBe(1_200_000);
  });

  it("falls back to avgBuyPrice when currentPrice is missing", () => {
    const holdings = [holding({ avgBuyPrice: 1000, lots: 10 })];
    expect(currentValue(holdings)).toBe(1_000_000);
  });
});

describe("overallPL", () => {
  it("returns 0 for empty holdings", () => {
    expect(overallPL([])).toBe(0);
  });

  it("calculates overall percentage", () => {
    const holdings = [holding({ avgBuyPrice: 1000, lots: 10, currentPrice: 1200 })];
    // invested: 1_000_000, current: 1_200_000
    // PL: (1_200_000 - 1_000_000) / 1_000_000 * 100 = 20
    expect(overallPL(holdings)).toBeCloseTo(20);
  });

  it("handles negative overall PL", () => {
    const holdings = [holding({ avgBuyPrice: 1000, lots: 10, currentPrice: 800 })];
    expect(overallPL(holdings)).toBeCloseTo(-20);
  });
});

describe("calcPLAbsolute", () => {
  it("returns null when currentPrice is undefined", () => {
    expect(calcPLAbsolute(holding({ avgBuyPrice: 1000, lots: 10 }))).toBeNull();
  });

  it("calculates positive absolute P/L", () => {
    const h = holding({ avgBuyPrice: 1000, lots: 10, currentPrice: 1200 });
    // (1200 - 1000) * 10 * 100 = 200_000
    expect(calcPLAbsolute(h)).toBe(200_000);
  });

  it("calculates negative absolute P/L", () => {
    const h = holding({ avgBuyPrice: 1000, lots: 10, currentPrice: 800 });
    // (800 - 1000) * 10 * 100 = -200_000
    expect(calcPLAbsolute(h)).toBe(-200_000);
  });
});

describe("holdingWeights", () => {
  it("returns empty for empty holdings", () => {
    expect(holdingWeights([])).toEqual([]);
  });

  it("calculates weight percentages", () => {
    const holdings = [
      holding({ ticker: "BBCA", avgBuyPrice: 1000, lots: 10, currentPrice: 1000 }),
      holding({ ticker: "TLKM", avgBuyPrice: 500, lots: 10, currentPrice: 500 }),
    ];
    // BBCA value: 1000*10*100 = 1_000_000
    // TLKM value: 500*10*100 = 500_000
    // total = 1_500_000
    const weights = holdingWeights(holdings);
    expect(weights).toHaveLength(2);
    expect(weights[0].ticker).toBe("BBCA");
    expect(weights[0].value).toBe(1_000_000);
    expect(weights[0].pct).toBeCloseTo(66.67, 1);
    expect(weights[1].ticker).toBe("TLKM");
    expect(weights[1].pct).toBeCloseTo(33.33, 1);
  });

  it("uses avgBuyPrice when currentPrice is missing", () => {
    const holdings = [holding({ ticker: "BBCA", avgBuyPrice: 1000, lots: 10 })];
    const weights = holdingWeights(holdings);
    expect(weights[0].value).toBe(1_000_000);
    expect(weights[0].pct).toBe(100);
  });
});

describe("sectorWeights", () => {
  it("returns empty for empty holdings", () => {
    expect(sectorWeights([], {})).toEqual([]);
  });

  it("groups by sector", () => {
    const holdings = [
      holding({ ticker: "BBCA", avgBuyPrice: 1000, lots: 10, currentPrice: 1000 }),
      holding({ ticker: "BBRI", avgBuyPrice: 500, lots: 10, currentPrice: 500 }),
      holding({ ticker: "TLKM", avgBuyPrice: 500, lots: 10, currentPrice: 500 }),
    ];
    const sectorMap = { BBCA: "Financials", BBRI: "Financials", TLKM: "Telco" };
    // Financials: 1_000_000 + 500_000 = 1_500_000
    // Telco: 500_000
    // total = 2_000_000
    const weights = sectorWeights(holdings, sectorMap);
    const financials = weights.find((w) => w.sector === "Financials");
    const telco = weights.find((w) => w.sector === "Telco");
    expect(financials).toBeDefined();
    expect(financials?.value).toBe(1_500_000);
    expect(financials?.pct).toBe(75);
    expect(telco).toBeDefined();
    expect(telco?.pct).toBe(25);
  });

  it("assigns Unknown for missing sector", () => {
    const holdings = [holding({ ticker: "XXXX", avgBuyPrice: 1000, lots: 1, currentPrice: 1000 })];
    const weights = sectorWeights(holdings, {});
    expect(weights[0].sector).toBe("Unknown");
  });
});
