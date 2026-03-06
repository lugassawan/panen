import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetYoCProgression: vi.fn(() =>
    Promise.resolve([
      { date: "2023-03-15", yoc: 2.5 },
      { date: "2023-09-15", yoc: 5.5 },
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
  accentPalette: (n: number) => Array.from({ length: n }, () => "#d4a12a"),
}));

import YocProgressionChart from "./YocProgressionChart.svelte";

describe("YocProgressionChart", () => {
  it("renders the component", () => {
    render(YocProgressionChart, { props: { portfolioId: "p1", tickers: ["BBCA"] } });
    expect(screen.getByTestId("yoc-progression-chart")).toBeInTheDocument();
  });

  it("shows single ticker name", () => {
    render(YocProgressionChart, { props: { portfolioId: "p1", tickers: ["BBCA"] } });
    expect(screen.getByText("BBCA")).toBeInTheDocument();
  });
});
