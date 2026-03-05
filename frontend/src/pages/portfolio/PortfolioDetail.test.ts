import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  AddHolding: vi.fn(),
  GetDividendRanking: vi.fn(),
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
});
