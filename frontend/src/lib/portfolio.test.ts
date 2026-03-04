import { describe, expect, it } from "vitest";
import { calcPL, currentValue, overallPL, totalInvested } from "./portfolio";
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
