import { render, screen, waitFor } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

const mockGetHoldingSectors = vi.fn(() =>
  Promise.resolve({ BBCA: "Financials", TLKM: "Communication Services" }),
);

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetHoldingSectors: (...args: unknown[]) => mockGetHoldingSectors(...args),
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
    ArcElement: {},
    DoughnutController: {},
    Legend: {},
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
  accentPalette: (n: number) => Array.from({ length: n }, () => "#1b6b4a"),
}));

import type { HoldingDetailResponse } from "../../lib/types";
import ChartsTab from "./ChartsTab.svelte";

const holdings: HoldingDetailResponse[] = [
  { id: "h1", ticker: "BBCA", avgBuyPrice: 8000, lots: 10, currentPrice: 9000 },
  { id: "h2", ticker: "TLKM", avgBuyPrice: 3000, lots: 5, currentPrice: 3200 },
];

describe("ChartsTab", () => {
  it("calls GetHoldingSectors with tickers", async () => {
    render(ChartsTab, {
      props: { holdings, portfolioMode: "VALUE" },
    });

    await waitFor(() => {
      expect(mockGetHoldingSectors).toHaveBeenCalledWith(["BBCA", "TLKM"]);
    });
  });

  it("renders sub-components after loading", async () => {
    render(ChartsTab, {
      props: { holdings, portfolioMode: "VALUE" },
    });

    await waitFor(() => {
      expect(screen.getByTestId("pl-bar-chart")).toBeInTheDocument();
      expect(screen.getByTestId("composition-chart")).toBeInTheDocument();
      expect(screen.getByTestId("sector-warnings")).toBeInTheDocument();
    });
  });

  it("shows loading state initially", () => {
    render(ChartsTab, {
      props: { holdings, portfolioMode: "VALUE" },
    });

    expect(screen.getByText(/Loading chart data/)).toBeInTheDocument();
  });

  it("handles empty holdings without calling backend", async () => {
    mockGetHoldingSectors.mockClear();
    render(ChartsTab, {
      props: { holdings: [], portfolioMode: "VALUE" },
    });

    await waitFor(() => {
      expect(mockGetHoldingSectors).not.toHaveBeenCalled();
    });
  });

  it("shows error message when backend call fails", async () => {
    mockGetHoldingSectors.mockImplementationOnce(() => Promise.reject(new Error("network error")));
    render(ChartsTab, {
      props: { holdings, portfolioMode: "VALUE" },
    });

    await waitFor(() => {
      expect(screen.getByText("network error")).toBeInTheDocument();
    });
  });
});
