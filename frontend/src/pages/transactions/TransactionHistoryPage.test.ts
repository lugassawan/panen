import { render, screen } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { PortfolioResponse, TransactionListResponse } from "../../lib/types";
import TransactionHistoryPage from "./TransactionHistoryPage.svelte";

const mockListTransactions = vi.fn();
const mockListPortfolios = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  ListTransactions: (...args: unknown[]) => mockListTransactions(...args),
  ListPortfolios: (...args: unknown[]) => mockListPortfolios(...args),
}));

function makeResponse(overrides: Partial<TransactionListResponse> = {}): TransactionListResponse {
  return {
    items: [],
    summary: {
      totalBuyAmount: 0,
      totalSellAmount: 0,
      totalDividendAmount: 0,
      totalFees: 0,
      transactionCount: 0,
    },
    ...overrides,
  };
}

function makePortfolios(): PortfolioResponse[] {
  return [
    {
      id: "p1",
      brokerageAcctId: "b1",
      name: "Growth",
      mode: "VALUE",
      riskProfile: "MODERATE",
      capital: 0,
      monthlyAddition: 0,
      maxStocks: 0,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
    },
  ];
}

beforeEach(() => {
  vi.clearAllMocks();
  mockListPortfolios.mockResolvedValue(makePortfolios());
});

describe("TransactionHistoryPage", () => {
  it("shows loading then empty state when no transactions", async () => {
    mockListTransactions.mockResolvedValue(makeResponse());

    render(TransactionHistoryPage);

    const emptyTitle = await screen.findByText("No transactions yet");
    expect(emptyTitle).toBeTruthy();
  });

  it("shows summary cards with formatted amounts", async () => {
    mockListTransactions.mockResolvedValue(
      makeResponse({
        summary: {
          totalBuyAmount: 5000000,
          totalSellAmount: 1000000,
          totalDividendAmount: 250000,
          totalFees: 15000,
          transactionCount: 3,
        },
        items: [
          {
            id: "1",
            type: "BUY",
            date: "2026-01-15",
            ticker: "BBCA",
            portfolioId: "p1",
            portfolioName: "Growth",
            lots: 5,
            price: 8500,
            fee: 6375,
            tax: 0,
            total: 4256375,
          },
        ],
      }),
    );

    render(TransactionHistoryPage);

    const container = document.body;
    await screen.findByText("BBCA");

    expect(container.textContent).toContain("5,000,000");
    expect(container.textContent).toContain("250,000");
    expect(container.textContent).toContain("15,000");
  });

  it("shows transaction rows in table", async () => {
    mockListTransactions.mockResolvedValue(
      makeResponse({
        items: [
          {
            id: "1",
            type: "BUY",
            date: "2026-01-15",
            ticker: "BBCA",
            portfolioId: "p1",
            portfolioName: "Growth",
            lots: 5,
            price: 8500,
            fee: 6375,
            tax: 0,
            total: 4256375,
          },
          {
            id: "2",
            type: "SELL",
            date: "2026-02-10",
            ticker: "BBRI",
            portfolioId: "p1",
            portfolioName: "Growth",
            lots: 3,
            price: 4500,
            fee: 3375,
            tax: 1350,
            total: 1345275,
          },
        ],
        summary: {
          totalBuyAmount: 4256375,
          totalSellAmount: 1345275,
          totalDividendAmount: 0,
          totalFees: 11100,
          transactionCount: 2,
        },
      }),
    );

    render(TransactionHistoryPage);

    const bbca = await screen.findByText("BBCA");
    expect(bbca).toBeTruthy();

    const bbri = screen.getByText("BBRI");
    expect(bbri).toBeTruthy();

    // "Buy" and "Sell" appear both as filter buttons and table badges.
    // Use getAllByText to verify the badge instances exist alongside the filters.
    const buyMatches = screen.getAllByText("Buy");
    expect(buyMatches.length).toBeGreaterThanOrEqual(2);

    const sellMatches = screen.getAllByText("Sell");
    expect(sellMatches.length).toBeGreaterThanOrEqual(2);
  });
});
