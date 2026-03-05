import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import type { MarketStatusResponse } from "../../lib/types";
import MarketStatusBanner from "./MarketStatusBanner.svelte";

function makeMarket(overrides: Partial<MarketStatusResponse> = {}): MarketStatusResponse {
  return {
    condition: "NORMAL",
    ihsgPrice: 7200,
    ihsgPeak: 7500,
    drawdownPct: -4.0,
    fetchedAt: "2026-03-05T10:00:00Z",
    ...overrides,
  };
}

describe("MarketStatusBanner", () => {
  it("renders IHSG price", () => {
    render(MarketStatusBanner, { props: { market: makeMarket() } });
    expect(document.body.textContent).toContain("7.200");
  });

  it("shows Normal badge for normal condition", () => {
    render(MarketStatusBanner, { props: { market: makeMarket() } });
    expect(screen.getByText("Normal", { exact: true })).toBeTruthy();
  });

  it("shows Crash badge for crash condition", () => {
    render(MarketStatusBanner, {
      props: { market: makeMarket({ condition: "CRASH", drawdownPct: -25 }) },
    });
    expect(screen.getByText("Crash", { exact: true })).toBeTruthy();
  });

  it("shows drawdown percentage", () => {
    render(MarketStatusBanner, { props: { market: makeMarket() } });
    expect(document.body.textContent).toContain("-4,00%");
  });

  it("shows peak price", () => {
    render(MarketStatusBanner, { props: { market: makeMarket() } });
    expect(document.body.textContent).toContain("7.500");
  });
});
