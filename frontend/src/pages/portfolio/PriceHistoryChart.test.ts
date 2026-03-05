import { render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

const mockGetPriceHistory = vi.fn(() =>
  Promise.resolve([
    { date: "2025-01-02", open: 9000, high: 9200, low: 8900, close: 9100, volume: 100000 },
    { date: "2025-02-01", open: 9100, high: 9300, low: 9000, close: 9250, volume: 150000 },
  ]),
);

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetPriceHistory: (...args: unknown[]) => mockGetPriceHistory(...args),
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

vi.mock("chartjs-plugin-annotation", () => ({
  default: {},
}));

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
  valuationZoneColors: () => ({
    graham: "#9333ea",
    entry: "#16a34a",
    exit: "#dc2626",
    entryBand: "#16a34a18",
    exitBand: "#dc262618",
  }),
}));

import PriceHistoryChart from "./PriceHistoryChart.svelte";

describe("PriceHistoryChart", () => {
  it("calls GetPriceHistory with ticker when single ticker", async () => {
    render(PriceHistoryChart, {
      props: { tickers: ["BBCA"] },
    });

    await waitFor(() => {
      expect(mockGetPriceHistory).toHaveBeenCalledWith("BBCA");
    });
  });

  it("renders chart container after loading", async () => {
    render(PriceHistoryChart, {
      props: { tickers: ["BBCA"] },
    });

    await waitFor(() => {
      expect(screen.getByTestId("price-history-chart")).toBeInTheDocument();
    });
  });

  it("shows loading state initially", () => {
    render(PriceHistoryChart, {
      props: { tickers: ["BBCA"] },
    });

    expect(screen.getByText(/Loading price history/)).toBeInTheDocument();
  });

  it("shows time range pills after data loads", async () => {
    render(PriceHistoryChart, {
      props: { tickers: ["BBCA"] },
    });

    await waitFor(() => {
      expect(screen.getByRole("group", { name: "Time range" })).toBeInTheDocument();
      expect(screen.getByText("1M")).toBeInTheDocument();
      expect(screen.getByText("ALL")).toBeInTheDocument();
    });
  });

  it("shows ticker selector when multiple tickers", async () => {
    render(PriceHistoryChart, {
      props: { tickers: ["BBCA", "TLKM"] },
    });

    await waitFor(() => {
      expect(screen.getByLabelText("Select ticker")).toBeInTheDocument();
    });
  });

  it("shows error message when backend call fails", async () => {
    mockGetPriceHistory.mockImplementationOnce(() => Promise.reject(new Error("network error")));
    render(PriceHistoryChart, {
      props: { tickers: ["BBCA"] },
    });

    await waitFor(() => {
      expect(screen.getByText("network error")).toBeInTheDocument();
    });
  });

  it("shows empty state when no tickers provided", () => {
    render(PriceHistoryChart, {
      props: { tickers: [] },
    });

    expect(screen.getByText("Select a ticker")).toBeInTheDocument();
  });

  it("does not show valuation toggles without valuation props", async () => {
    render(PriceHistoryChart, {
      props: { tickers: ["BBCA"] },
    });

    await waitFor(() => {
      expect(screen.getByTestId("price-history-chart")).toBeInTheDocument();
    });

    expect(screen.queryByRole("group", { name: "Valuation zones" })).not.toBeInTheDocument();
  });

  it("shows valuation toggles when valuation data provided", async () => {
    render(PriceHistoryChart, {
      props: {
        tickers: ["BBCA"],
        valuations: {
          BBCA: { grahamNumber: 8000, entryPrice: 7500, exitTarget: 11000 },
        },
      },
    });

    await waitFor(() => {
      expect(screen.getByRole("group", { name: "Valuation zones" })).toBeInTheDocument();
      expect(screen.getByLabelText("Graham")).toBeInTheDocument();
      expect(screen.getByLabelText("Entry Price")).toBeInTheDocument();
      expect(screen.getByLabelText("Exit Target")).toBeInTheDocument();
    });
  });

  it("has all valuation toggles checked by default", async () => {
    render(PriceHistoryChart, {
      props: {
        tickers: ["BBCA"],
        valuations: {
          BBCA: { grahamNumber: 8000, entryPrice: 7500, exitTarget: 11000 },
        },
      },
    });

    await waitFor(() => {
      expect(screen.getByLabelText("Graham")).toBeChecked();
      expect(screen.getByLabelText("Entry Price")).toBeChecked();
      expect(screen.getByLabelText("Exit Target")).toBeChecked();
    });
  });

  it("does not show toggles when all valuation values are zero", async () => {
    render(PriceHistoryChart, {
      props: {
        tickers: ["BBCA"],
        valuations: {
          BBCA: { grahamNumber: 0, entryPrice: 0, exitTarget: 0 },
        },
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId("price-history-chart")).toBeInTheDocument();
    });

    expect(screen.queryByRole("group", { name: "Valuation zones" })).not.toBeInTheDocument();
  });

  it("unchecking toggle updates checkbox state", async () => {
    const user = userEvent.setup();
    render(PriceHistoryChart, {
      props: {
        tickers: ["BBCA"],
        valuations: {
          BBCA: { grahamNumber: 8000, entryPrice: 7500, exitTarget: 11000 },
        },
      },
    });

    await waitFor(() => {
      expect(screen.getByLabelText("Graham")).toBeChecked();
    });

    await user.click(screen.getByLabelText("Graham"));
    expect(screen.getByLabelText("Graham")).not.toBeChecked();
  });
});
