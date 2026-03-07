import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { StockValuationResponse } from "../../lib/types";
import ComparisonPage from "./ComparisonPage.svelte";

vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

const mockLookupStock = vi.fn();
const mockListWatchlists = vi.fn(() => Promise.resolve([]));
const mockGetWatchlistItems = vi.fn(() => Promise.resolve([]));

vi.mock("../../../wailsjs/go/backend/App", () => ({
  LookupStock: (...args: unknown[]) => mockLookupStock(...args),
  ListWatchlists: () => mockListWatchlists(),
  GetWatchlistItems: (...args: unknown[]) => mockGetWatchlistItems(...args),
  GetAlertCount: vi.fn(() => Promise.resolve(0)),
}));

function makeResponse(overrides: Partial<StockValuationResponse> = {}): StockValuationResponse {
  return {
    ticker: "BBCA",
    price: 9250,
    high52Week: 11000,
    low52Week: 7500,
    eps: 350,
    bvps: 2100,
    roe: 21.5,
    der: 5.2,
    pbv: 4.4,
    per: 26.4,
    dividendYield: 2.5,
    payoutRatio: 55.0,
    grahamNumber: 4073,
    marginOfSafety: 30.0,
    entryPrice: 2851,
    exitTarget: 6500,
    verdict: "UNDERVALUED",
    riskProfile: "MODERATE",
    fetchedAt: "2025-01-15T10:30:00Z",
    source: "Yahoo Finance",
    ...overrides,
  };
}

describe("ComparisonPage", () => {
  beforeEach(() => {
    mockLookupStock.mockReset();
    mockListWatchlists.mockReset().mockResolvedValue([]);
    mockGetWatchlistItems.mockReset().mockResolvedValue([]);
  });

  it("renders 2 input slots and Compare button", () => {
    render(ComparisonPage);
    const slots = screen.getAllByTestId("ticker-slot");
    expect(slots).toHaveLength(2);
    expect(screen.getByRole("button", { name: /Compare$/i })).toBeInTheDocument();
  });

  it("shows empty state before comparison", () => {
    render(ComparisonPage);
    expect(screen.getByText("Compare Stocks")).toBeInTheDocument();
    expect(screen.getByText(/Enter at least 2 tickers/)).toBeInTheDocument();
  });

  it("adds ticker slot up to 4 and removes down to 2", async () => {
    const user = userEvent.setup();
    render(ComparisonPage);

    expect(screen.getAllByTestId("ticker-slot")).toHaveLength(2);

    await user.click(screen.getByText("Add Ticker"));
    expect(screen.getAllByTestId("ticker-slot")).toHaveLength(3);

    await user.click(screen.getByText("Add Ticker"));
    expect(screen.getAllByTestId("ticker-slot")).toHaveLength(4);

    expect(screen.queryByText("Add Ticker")).not.toBeInTheDocument();

    const removeButtons = screen.getAllByLabelText("Remove");
    await user.click(removeButtons[0]);
    expect(screen.getAllByTestId("ticker-slot")).toHaveLength(3);

    await user.click(screen.getAllByLabelText("Remove")[0]);
    expect(screen.getAllByTestId("ticker-slot")).toHaveLength(2);
  });

  it("calls LookupStock for each filled ticker on submit", async () => {
    const user = userEvent.setup();
    mockLookupStock
      .mockResolvedValueOnce(makeResponse({ ticker: "BBCA" }))
      .mockResolvedValueOnce(makeResponse({ ticker: "BMRI", price: 5100 }));
    render(ComparisonPage);

    const inputs = screen.getAllByRole("textbox");
    await user.type(inputs[0], "BBCA");
    await user.type(inputs[1], "BMRI");
    await user.click(screen.getByRole("button", { name: /Compare$/i }));

    expect(mockLookupStock).toHaveBeenCalledTimes(2);
    expect(mockLookupStock).toHaveBeenCalledWith("BBCA", "MODERATE");
    expect(mockLookupStock).toHaveBeenCalledWith("BMRI", "MODERATE");
  });

  it("displays comparison table with results", async () => {
    const user = userEvent.setup();
    mockLookupStock
      .mockResolvedValueOnce(makeResponse({ ticker: "BBCA" }))
      .mockResolvedValueOnce(makeResponse({ ticker: "BMRI", price: 5100 }));
    render(ComparisonPage);

    const inputs = screen.getAllByRole("textbox");
    await user.type(inputs[0], "BBCA");
    await user.type(inputs[1], "BMRI");
    await user.click(screen.getByRole("button", { name: /Compare$/i }));

    expect(await screen.findByTestId("comparison-table")).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("BMRI")).toBeInTheDocument();
  });

  it("highlights best value per metric row", async () => {
    const user = userEvent.setup();
    mockLookupStock
      .mockResolvedValueOnce(makeResponse({ ticker: "BBCA", grahamNumber: 4073 }))
      .mockResolvedValueOnce(makeResponse({ ticker: "BMRI", grahamNumber: 3200 }));
    render(ComparisonPage);

    const inputs = screen.getAllByRole("textbox");
    await user.type(inputs[0], "BBCA");
    await user.type(inputs[1], "BMRI");
    await user.click(screen.getByRole("button", { name: /Compare$/i }));

    await screen.findByTestId("comparison-table");
    const table = screen.getByTestId("comparison-table");
    const profitCells = table.querySelectorAll(".text-profit");
    expect(profitCells.length).toBeGreaterThan(0);
  });

  it("handles partial failures", async () => {
    const user = userEvent.setup();
    mockLookupStock
      .mockResolvedValueOnce(makeResponse({ ticker: "BBCA" }))
      .mockRejectedValueOnce(new Error("ticker not found"));
    render(ComparisonPage);

    const inputs = screen.getAllByRole("textbox");
    await user.type(inputs[0], "BBCA");
    await user.type(inputs[1], "XXXX");
    await user.click(screen.getByRole("button", { name: /Compare$/i }));

    expect(await screen.findByText(/ticker not found/)).toBeInTheDocument();
    expect(screen.getByTestId("comparison-table")).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText(/1 of 2 loaded/)).toBeInTheDocument();
  });

  it("requires minimum 2 filled tickers to compare", async () => {
    const user = userEvent.setup();
    render(ComparisonPage);

    const inputs = screen.getAllByRole("textbox");
    await user.type(inputs[0], "BBCA");

    const button = screen.getByRole("button", { name: /Compare$/i });
    expect(button).toBeDisabled();
  });

  it("auto-uppercases ticker codes", async () => {
    const user = userEvent.setup();
    mockLookupStock
      .mockResolvedValueOnce(makeResponse({ ticker: "BBCA" }))
      .mockResolvedValueOnce(makeResponse({ ticker: "BMRI" }));
    render(ComparisonPage);

    const inputs = screen.getAllByRole("textbox");
    await user.type(inputs[0], "bbca");
    await user.type(inputs[1], "bmri");
    await user.click(screen.getByRole("button", { name: /Compare$/i }));

    await screen.findByTestId("comparison-table");
    expect(mockLookupStock).toHaveBeenCalledWith("BBCA", "MODERATE");
    expect(mockLookupStock).toHaveBeenCalledWith("BMRI", "MODERATE");
  });
});
