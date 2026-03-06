import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetDividendIncomeSummary: vi.fn(() =>
    Promise.resolve({
      totalAnnualIncome: 0,
      perStock: [],
      monthlyBreakdown: [],
    }),
  ),
  GetDGR: vi.fn(() => Promise.resolve([])),
  GetYoCProgression: vi.fn(() => Promise.resolve([])),
  GetDividendCalendar: vi.fn(() => Promise.resolve([])),
}));

vi.mock("chart.js", () => {
  class MockChart {
    static register = vi.fn();
    destroy = vi.fn();
  }
  return {
    Chart: MockChart,
    BarController: {},
    BarElement: {},
    CategoryScale: {},
    Filler: {},
    LinearScale: {},
    LineController: {},
    LineElement: {},
    PointElement: {},
    Tooltip: {},
  };
});

vi.mock("../../lib/chartColors.svelte", () => ({
  chartColors: () => ({
    profit: "#1b7d4e",
    loss: "#c4342d",
    textPrimary: "#1a1a1a",
    textSecondary: "#4b5060",
    textMuted: "#9ca3af",
    borderDefault: "#e0dbd2",
    bgElevated: "#ffffff",
  }),
  defaultChartOptions: () => ({
    responsive: true,
    maintainAspectRatio: false,
    animation: { duration: 200 },
    plugins: {},
    scales: { x: {}, y: {} },
  }),
}));

import DividendChartsTab from "./DividendChartsTab.svelte";

describe("DividendChartsTab", () => {
  const holdings = [{ id: "h1", ticker: "BBCA", avgBuyPrice: 8000, lots: 10 }];

  it("renders all sub-components", () => {
    render(DividendChartsTab, { props: { portfolioId: "p1", holdings } });
    expect(screen.getByTestId("dividend-charts-tab")).toBeInTheDocument();
    expect(screen.getByTestId("dividend-income-chart")).toBeInTheDocument();
    expect(screen.getByTestId("dgr-chart")).toBeInTheDocument();
    expect(screen.getByTestId("yoc-progression-chart")).toBeInTheDocument();
    expect(screen.getByTestId("dividend-calendar-panel")).toBeInTheDocument();
  });
});
