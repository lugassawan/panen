import { render, screen, waitFor } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

const mockGetCashFlowSummary = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetCashFlowSummary: (...args: unknown[]) => mockGetCashFlowSummary(...args),
}));

import CashFlowTable from "./CashFlowTable.svelte";

describe("CashFlowTable", () => {
  it("shows loading state initially", () => {
    mockGetCashFlowSummary.mockReturnValue(new Promise(() => {}));
    render(CashFlowTable, { props: { portfolioId: "p1" } });
    expect(screen.getByText("Loading cash flows...")).toBeInTheDocument();
  });

  it("renders summary data on success", async () => {
    mockGetCashFlowSummary.mockResolvedValueOnce({
      totalInflow: 5000000,
      totalDeployed: 3000000,
      balance: 2000000,
      items: [
        {
          id: "cf1",
          portfolioId: "p1",
          type: "MONTHLY",
          amount: 1000000,
          date: "2025-06-25",
          note: "June payday",
          createdAt: "2025-06-25T00:00:00Z",
        },
      ],
    });

    render(CashFlowTable, { props: { portfolioId: "p1" } });

    await waitFor(() => {
      expect(screen.getByText("Total Inflow")).toBeInTheDocument();
    });
    expect(screen.getByText("Balance")).toBeInTheDocument();
    expect(screen.getByText("MONTHLY")).toBeInTheDocument();
    expect(screen.getByText("June payday")).toBeInTheDocument();
  });

  it("shows error on failure", async () => {
    mockGetCashFlowSummary.mockRejectedValueOnce(new Error("network error"));

    render(CashFlowTable, { props: { portfolioId: "p1" } });

    await waitFor(() => {
      expect(screen.getByText("network error")).toBeInTheDocument();
    });
  });

  it("shows empty state when no items", async () => {
    mockGetCashFlowSummary.mockResolvedValueOnce({
      totalInflow: 0,
      totalDeployed: 0,
      balance: 0,
      items: [],
    });

    render(CashFlowTable, { props: { portfolioId: "p1" } });

    await waitFor(() => {
      expect(screen.getByText("No cash flow records yet.")).toBeInTheDocument();
    });
  });
});
