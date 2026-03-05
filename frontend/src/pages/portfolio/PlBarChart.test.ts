import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

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

import type { HoldingDetailResponse } from "../../lib/types";
import PlBarChart from "./PlBarChart.svelte";

const holdings: HoldingDetailResponse[] = [
  { id: "h1", ticker: "BBCA", avgBuyPrice: 8000, lots: 10, currentPrice: 9000 },
  { id: "h2", ticker: "TLKM", avgBuyPrice: 3000, lots: 5, currentPrice: 2700 },
];

describe("PlBarChart", () => {
  it("renders chart heading", () => {
    render(PlBarChart, { props: { holdings } });
    expect(screen.getByText("P/L by Holding")).toBeInTheDocument();
  });

  it("shows toggle buttons when holdings have prices", () => {
    render(PlBarChart, { props: { holdings } });
    expect(screen.getByRole("button", { name: "%" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Rp" })).toBeInTheDocument();
  });

  it("shows empty state when no holdings have current prices", () => {
    const noPrice: HoldingDetailResponse[] = [
      { id: "h1", ticker: "BBCA", avgBuyPrice: 8000, lots: 10 },
    ];
    render(PlBarChart, { props: { holdings: noPrice } });
    expect(screen.getByText("No P/L data")).toBeInTheDocument();
  });

  it("renders accessible data table", () => {
    render(PlBarChart, { props: { holdings } });
    expect(screen.getByRole("table")).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("TLKM")).toBeInTheDocument();
  });

  it("toggles to rupiah mode", async () => {
    const user = userEvent.setup();
    render(PlBarChart, { props: { holdings } });

    const rpButton = screen.getByRole("button", { name: "Rp" });
    await user.click(rpButton);
    expect(rpButton).toHaveAttribute("aria-pressed", "true");
  });
});
