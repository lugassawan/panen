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
    ArcElement: {},
    DoughnutController: {},
    Legend: {},
    Tooltip: {},
  };
});

vi.mock("../../lib/chartColors.svelte", () => ({
  accentPalette: (n: number) => Array.from({ length: n }, () => "#1b6b4a"),
  defaultChartOptions: () => ({
    responsive: true,
    maintainAspectRatio: false,
    animation: { duration: 200 },
    plugins: { legend: { labels: {} }, tooltip: {} },
    scales: { x: {}, y: {} },
  }),
}));

import type { HoldingWeight, SectorWeight } from "../../lib/types";
import CompositionChart from "./CompositionChart.svelte";

const holdingWeights: HoldingWeight[] = [
  { ticker: "BBCA", value: 9_000_000, pct: 64.29 },
  { ticker: "TLKM", value: 5_000_000, pct: 35.71 },
];

const sectorWeights: SectorWeight[] = [
  { sector: "Financials", value: 9_000_000, pct: 64.29 },
  { sector: "Communication Services", value: 5_000_000, pct: 35.71 },
];

describe("CompositionChart", () => {
  const defaultProps = {
    holdingWeights,
    sectorWeights,
    portfolioMode: "VALUE" as const,
  };

  it("renders composition heading", () => {
    render(CompositionChart, { props: defaultProps });
    expect(screen.getByText("Composition")).toBeInTheDocument();
  });

  it("shows toggle buttons", () => {
    render(CompositionChart, { props: defaultProps });
    expect(screen.getByRole("button", { name: "By Holding" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "By Sector" })).toBeInTheDocument();
  });

  it("shows By Holding active by default", () => {
    render(CompositionChart, { props: defaultProps });
    const btn = screen.getByRole("button", { name: "By Holding" });
    expect(btn).toHaveAttribute("aria-pressed", "true");
  });

  it("toggles to sector view", async () => {
    const user = userEvent.setup();
    render(CompositionChart, { props: defaultProps });

    await user.click(screen.getByRole("button", { name: "By Sector" }));
    const sectorBtn = screen.getByRole("button", { name: "By Sector" });
    expect(sectorBtn).toHaveAttribute("aria-pressed", "true");
  });

  it("shows empty state when no holdings", () => {
    render(CompositionChart, {
      props: { holdingWeights: [], sectorWeights: [], portfolioMode: "VALUE" },
    });
    expect(screen.getByText("No composition data")).toBeInTheDocument();
  });

  it("renders accessible data table", () => {
    render(CompositionChart, { props: defaultProps });
    expect(screen.getByRole("table")).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("TLKM")).toBeInTheDocument();
  });
});
