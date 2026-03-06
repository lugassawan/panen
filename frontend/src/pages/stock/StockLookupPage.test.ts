import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { StockValuationResponse } from "../../lib/types";
import StockLookupPage from "./StockLookupPage.svelte";

vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

const mockLookupStock = vi.fn();
vi.mock("../../../wailsjs/go/backend/App", () => ({
  LookupStock: (...args: unknown[]) => mockLookupStock(...args),
  GetAlertCount: vi.fn(() => Promise.resolve(0)),
  GetActiveAlerts: vi.fn(() => Promise.resolve([])),
  GetAlertsByTicker: vi.fn(() => Promise.resolve([])),
  AcknowledgeAlert: vi.fn(() => Promise.resolve()),
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

describe("StockLookupPage", () => {
  beforeEach(() => {
    mockLookupStock.mockReset();
  });

  it("renders search form with input, selector, and button", () => {
    render(StockLookupPage);
    expect(screen.getByLabelText("Stock ticker")).toBeInTheDocument();
    expect(screen.getByLabelText("Risk profile")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Lookup/i })).toBeInTheDocument();
  });

  it("submits ticker and displays results", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockResolvedValueOnce(makeResponse());
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "bbca");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    expect(await screen.findByText(/Undervalued/)).toBeInTheDocument();
    expect(screen.getByText(/IDR\s*9,250/)).toBeInTheDocument();
    expect(screen.getByTestId("graham-number")).toHaveTextContent("IDR");
  });

  it("shows loading state during fetch", async () => {
    const user = userEvent.setup();
    let resolveFn!: (value: StockValuationResponse) => void;
    mockLookupStock.mockReturnValueOnce(
      new Promise<StockValuationResponse>((resolve) => {
        resolveFn = resolve;
      }),
    );
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "BBCA");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    expect(screen.getByText("Fetching valuation data...")).toBeInTheDocument();

    resolveFn(makeResponse());
    expect(await screen.findByText(/Undervalued/)).toBeInTheDocument();
  });

  it("shows error state on failure", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockRejectedValueOnce(new Error("ticker not found"));
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "XXXX");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    expect(await screen.findByText("ticker not found")).toBeInTheDocument();
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });

  it("does nothing when ticker is empty", async () => {
    const user = userEvent.setup();
    render(StockLookupPage);

    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    expect(mockLookupStock).not.toHaveBeenCalled();
  });

  it("displays accessible verdict text and icon", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockResolvedValueOnce(makeResponse());
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "BBCA");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    const verdictEl = await screen.findByText(/Undervalued/);
    expect(verdictEl).toBeInTheDocument();
    expect(verdictEl.closest("div")?.textContent).toContain("\u25B2");
  });

  it("displays 52-week range values", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockResolvedValueOnce(makeResponse());
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "BBCA");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    await screen.findByText(/Undervalued/);
    expect(screen.getByText(/52W Low/)).toBeInTheDocument();
    expect(screen.getByText(/52W High/)).toBeInTheDocument();
  });

  it("displays Graham number and entry zone", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockResolvedValueOnce(makeResponse());
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "BBCA");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    await screen.findByText(/Undervalued/);
    expect(screen.getByTestId("graham-number")).toBeInTheDocument();
    expect(screen.getByTestId("entry-price")).toBeInTheDocument();
  });

  it("renders PBV/PER bands when present", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockResolvedValueOnce(
      makeResponse({
        pbvBand: { min: 1.0, max: 3.0, avg: 2.0, median: 1.8 },
        perBand: { min: 10.0, max: 20.0, avg: 15.0, median: 14.0 },
      }),
    );
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "BBCA");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    await screen.findByText(/Undervalued/);
    expect(screen.getByTestId("pbv-band")).toBeInTheDocument();
    expect(screen.getByTestId("per-band")).toBeInTheDocument();
  });

  it("does not render band sections when data is missing", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockResolvedValueOnce(makeResponse());
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "BBCA");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    await screen.findByText(/Undervalued/);
    expect(screen.queryByTestId("pbv-band")).not.toBeInTheDocument();
    expect(screen.queryByTestId("per-band")).not.toBeInTheDocument();
  });

  it("passes correct risk profile value to API", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockResolvedValueOnce(makeResponse({ riskProfile: "AGGRESSIVE" }));
    render(StockLookupPage);

    await user.selectOptions(screen.getByLabelText("Risk profile"), "AGGRESSIVE");
    await user.type(screen.getByLabelText("Stock ticker"), "BBCA");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    await screen.findByText(/Undervalued/);
    expect(mockLookupStock).toHaveBeenCalledWith("BBCA", "AGGRESSIVE");
  });

  it("auto-uppercases ticker before API call", async () => {
    const user = userEvent.setup();
    mockLookupStock.mockResolvedValueOnce(makeResponse());
    render(StockLookupPage);

    await user.type(screen.getByLabelText("Stock ticker"), "bbca");
    await user.click(screen.getByRole("button", { name: /Lookup/i }));

    await screen.findByText(/Undervalued/);
    expect(mockLookupStock).toHaveBeenCalledWith("BBCA", "MODERATE");
  });
});
