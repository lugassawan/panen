import { render, screen, waitFor } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";
import DividendRankingPanel from "./DividendRankingPanel.svelte";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetDividendRanking: vi.fn(() =>
    Promise.resolve([
      {
        ticker: "TLKM",
        indicator: "AVERAGE_UP",
        dividendYield: 5.2,
        yieldOnCost: 7.8,
        payoutRatio: 60,
        positionPct: 15,
        score: 42.5,
        isHolding: true,
      },
      {
        ticker: "BBCA",
        indicator: "BUY_ZONE",
        dividendYield: 3.5,
        yieldOnCost: 0,
        payoutRatio: 45,
        positionPct: 0,
        score: 38.2,
        isHolding: false,
      },
    ]),
  ),
}));

describe("DividendRankingPanel", () => {
  it("renders the ranking table with items", async () => {
    render(DividendRankingPanel, {
      props: { portfolioId: "p1" },
    });

    await waitFor(() => {
      expect(screen.getByText("TLKM")).toBeInTheDocument();
    });

    expect(screen.getByTestId("dividend-ranking-panel")).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
  });

  it("shows indicator badges", async () => {
    render(DividendRankingPanel, {
      props: { portfolioId: "p1" },
    });

    await waitFor(() => {
      expect(screen.getByText("Average Up")).toBeInTheDocument();
    });

    expect(screen.getByText("Buy Zone")).toBeInTheDocument();
  });

  it("marks non-holding items as watchlist", async () => {
    render(DividendRankingPanel, {
      props: { portfolioId: "p1" },
    });

    await waitFor(() => {
      expect(screen.getByText("(watchlist)")).toBeInTheDocument();
    });
  });

  it("shows DCA Ranking header", async () => {
    render(DividendRankingPanel, {
      props: { portfolioId: "p1" },
    });

    await waitFor(() => {
      expect(screen.getByText("DCA Ranking")).toBeInTheDocument();
    });
  });
});
