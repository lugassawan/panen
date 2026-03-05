import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import type { DividendMetricsResponse } from "../../lib/types";
import DividendMetricsPanel from "./DividendMetricsPanel.svelte";

function makeMetrics(overrides: Partial<DividendMetricsResponse> = {}): DividendMetricsResponse {
  return {
    indicator: "AVERAGE_UP",
    annualDPS: 200,
    yieldOnCost: 6.67,
    projectedYoC: 5.5,
    portfolioYield: 4.2,
    ...overrides,
  };
}

describe("DividendMetricsPanel", () => {
  it("renders the panel with indicator badge", () => {
    render(DividendMetricsPanel, {
      props: { ticker: "BBCA", dividendMetrics: makeMetrics() },
    });

    expect(screen.getByTestId("dividend-metrics-panel")).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("Average Up")).toBeInTheDocument();
  });

  it("displays YoC and projected YoC values", () => {
    render(DividendMetricsPanel, {
      props: {
        ticker: "TLKM",
        dividendMetrics: makeMetrics({ yieldOnCost: 8.5, projectedYoC: 7.2 }),
      },
    });

    expect(screen.getByText("Yield on Cost")).toBeInTheDocument();
    expect(screen.getByText("Projected YoC")).toBeInTheDocument();
  });

  it("displays portfolio yield", () => {
    render(DividendMetricsPanel, {
      props: { ticker: "ASII", dividendMetrics: makeMetrics({ portfolioYield: 5.1 }) },
    });

    expect(screen.getByText("Portfolio Yield")).toBeInTheDocument();
  });

  it("shows Buy Zone indicator", () => {
    render(DividendMetricsPanel, {
      props: { ticker: "HMSP", dividendMetrics: makeMetrics({ indicator: "BUY_ZONE" }) },
    });

    expect(screen.getByText("Buy Zone")).toBeInTheDocument();
  });

  it("shows Overvalued indicator", () => {
    render(DividendMetricsPanel, {
      props: { ticker: "UNVR", dividendMetrics: makeMetrics({ indicator: "OVERVALUED" }) },
    });

    expect(screen.getByText("Overvalued")).toBeInTheDocument();
  });
});
