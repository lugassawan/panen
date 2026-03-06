import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetDividendIncomeSummary: vi.fn(() =>
    Promise.resolve({
      totalAnnualIncome: 500000,
      perStock: [{ ticker: "BBCA", annualIncome: 500000, dividendYield: 3.0, lots: 10 }],
      monthlyBreakdown: [
        { month: 3, amount: 250000 },
        { month: 9, amount: 250000 },
      ],
    }),
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
  accentPalette: (n: number) => Array.from({ length: n }, () => "#d4a12a"),
}));

import DividendIncomeChart from "./DividendIncomeChart.svelte";

describe("DividendIncomeChart", () => {
  it("renders the component", () => {
    render(DividendIncomeChart, { props: { portfolioId: "p1" } });
    expect(screen.getByTestId("dividend-income-chart")).toBeInTheDocument();
  });

  it("shows loading state initially", () => {
    render(DividendIncomeChart, { props: { portfolioId: "p1" } });
    expect(screen.getByText("Loading income data...")).toBeInTheDocument();
  });
});
