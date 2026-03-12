import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  AddHolding: vi.fn(),
  GetDividendRanking: vi.fn(),
  GetHoldingSectors: vi.fn(() => Promise.resolve({ BBCA: "Financials" })),
  GetPriceHistory: vi.fn(() => Promise.resolve([])),
  GetDividendIncomeSummary: vi.fn(() =>
    Promise.resolve({ totalAnnualIncome: 0, perStock: [], monthlyBreakdown: [] }),
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
    LinearScale: {},
    Tooltip: {},
    ArcElement: {},
    DoughnutController: {},
    Filler: {},
    Legend: {},
    LineController: {},
    LineElement: {},
    PointElement: {},
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

import type { PortfolioDetailResponse } from "../../lib/types";
import PortfolioDetail from "./PortfolioDetail.svelte";

const detail: PortfolioDetailResponse = {
  portfolio: {
    id: "p1",
    brokerageAcctId: "b1",
    name: "Growth",
    mode: "VALUE",
    riskProfile: "MODERATE",
    capital: 50000000,
    monthlyAddition: 0,
    maxStocks: 5,
    createdAt: "",
    updatedAt: "",
  },
  holdings: [
    {
      id: "h1",
      ticker: "BBCA",
      avgBuyPrice: 8000,
      lots: 10,
      currentPrice: 9000,
      verdict: "UNDERVALUED",
    },
  ],
};

describe("PortfolioDetail", () => {
  const defaultProps = {
    detail,
    onBack: vi.fn(),
    onChecklist: vi.fn(),
    onHoldingAdded: vi.fn(),
    onSell: vi.fn(),
    onRemove: vi.fn(),
    onClearAll: vi.fn(),
  };

  it("renders portfolio name and mode badge", () => {
    render(PortfolioDetail, { props: defaultProps });
    expect(screen.getByText("Growth")).toBeInTheDocument();
    expect(screen.getByText("Value")).toBeInTheDocument();
  });

  it("renders summary bar with totals", () => {
    render(PortfolioDetail, { props: defaultProps });
    expect(screen.getByTestId("total-invested")).toBeInTheDocument();
    expect(screen.getByTestId("current-value")).toBeInTheDocument();
    expect(screen.getByTestId("overall-pl")).toBeInTheDocument();
  });

  it("renders holdings table", () => {
    render(PortfolioDetail, { props: defaultProps });
    expect(screen.getByText("BBCA")).toBeInTheDocument();
  });

  it("calls onBack when back button clicked", async () => {
    const user = userEvent.setup();
    const onBack = vi.fn();
    render(PortfolioDetail, {
      props: { ...defaultProps, onBack },
    });

    await user.click(screen.getByRole("button", { name: /Back to list/i }));
    expect(onBack).toHaveBeenCalledOnce();
  });

  it("shows portfolio yield for DIVIDEND mode", () => {
    const dividendDetail: PortfolioDetailResponse = {
      portfolio: { ...detail.portfolio, mode: "DIVIDEND" },
      holdings: [
        {
          ...detail.holdings[0],
          dividendMetrics: {
            indicator: "STRONG",
            annualDPS: 200,
            yieldOnCost: 2.5,
            projectedYoC: 3.0,
            portfolioYield: 2.8,
          },
        },
      ],
    };
    render(PortfolioDetail, {
      props: { ...defaultProps, detail: dividendDetail },
    });
    expect(screen.getByTestId("portfolio-yield")).toBeInTheDocument();
  });

  it("renders Add Holding section", () => {
    render(PortfolioDetail, { props: defaultProps });
    expect(screen.getByRole("heading", { name: "Add Holding" })).toBeInTheDocument();
  });

  it("renders tab bar with Holdings and Charts tabs", () => {
    render(PortfolioDetail, { props: defaultProps });
    const tablist = screen.getByRole("tablist");
    expect(tablist).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /holdings/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /charts/i })).toBeInTheDocument();
  });

  it("shows holdings tab by default", () => {
    render(PortfolioDetail, { props: defaultProps });
    const holdingsTab = screen.getByRole("tab", { name: /holdings/i });
    expect(holdingsTab).toHaveAttribute("aria-selected", "true");
    expect(screen.getByText("BBCA")).toBeInTheDocument();
  });

  it("switches to charts tab on click", async () => {
    const user = userEvent.setup();
    render(PortfolioDetail, { props: defaultProps });

    await user.click(screen.getByRole("tab", { name: /charts/i }));

    const chartsTab = screen.getByRole("tab", { name: /charts/i });
    expect(chartsTab).toHaveAttribute("aria-selected", "true");
    expect(screen.queryByRole("heading", { name: "Add Holding" })).not.toBeInTheDocument();
  });

  it("switches tabs via keyboard ArrowRight/ArrowLeft", async () => {
    const user = userEvent.setup();
    render(PortfolioDetail, { props: defaultProps });

    const holdingsTab = screen.getByRole("tab", { name: /holdings/i });
    holdingsTab.focus();
    await user.keyboard("{ArrowRight}");

    expect(screen.getByRole("tab", { name: /charts/i })).toHaveAttribute("aria-selected", "true");

    await user.keyboard("{ArrowLeft}");
    expect(screen.getByRole("tab", { name: /holdings/i })).toHaveAttribute("aria-selected", "true");
  });

  it("inactive tab has tabindex -1", () => {
    render(PortfolioDetail, { props: defaultProps });
    expect(screen.getByRole("tab", { name: /holdings/i })).toHaveAttribute("tabindex", "0");
    expect(screen.getByRole("tab", { name: /charts/i })).toHaveAttribute("tabindex", "-1");
  });

  it("shows Dividends tab for DIVIDEND mode portfolios", () => {
    const dividendDetail = {
      portfolio: { ...detail.portfolio, mode: "DIVIDEND" as const },
      holdings: detail.holdings,
    };
    render(PortfolioDetail, {
      props: { ...defaultProps, detail: dividendDetail },
    });
    expect(screen.getByRole("tab", { name: /dividends/i })).toBeInTheDocument();
  });

  it("does not show Dividends tab for VALUE mode portfolios", () => {
    render(PortfolioDetail, { props: defaultProps });
    expect(screen.queryByRole("tab", { name: /dividends/i })).not.toBeInTheDocument();
  });
});
