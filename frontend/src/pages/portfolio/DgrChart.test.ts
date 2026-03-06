import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetDGR: vi.fn(() =>
    Promise.resolve([
      { year: 2022, dps: 100, growthPct: 0 },
      { year: 2023, dps: 120, growthPct: 20 },
    ]),
  ),
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
    LinearScale: {},
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

import DgrChart from "./DgrChart.svelte";

describe("DgrChart", () => {
  it("renders the component", () => {
    render(DgrChart, { props: { tickers: ["BBCA"] } });
    expect(screen.getByTestId("dgr-chart")).toBeInTheDocument();
  });

  it("shows single ticker name", () => {
    render(DgrChart, { props: { tickers: ["BBCA"] } });
    expect(screen.getByText("BBCA")).toBeInTheDocument();
  });
});
