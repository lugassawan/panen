import { render, screen, waitFor } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

const mockGetHoldingSectors = vi.fn(() =>
  Promise.resolve({ BBCA: "Financials", TLKM: "Communication Services" }),
);

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetHoldingSectors: (...args: unknown[]) => mockGetHoldingSectors(...args),
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
      props: { holdings, portfolioId: "p1", portfolioMode: "VALUE" },
    });

    await waitFor(() => {
      expect(mockGetHoldingSectors).toHaveBeenCalledWith(["BBCA", "TLKM"]);
    });
  });

  it("renders sub-components after loading", async () => {
    render(ChartsTab, {
      props: { holdings, portfolioId: "p1", portfolioMode: "VALUE" },
    });

    await waitFor(() => {
      expect(screen.getByTestId("pl-bar-chart")).toBeInTheDocument();
      expect(screen.getByTestId("composition-chart")).toBeInTheDocument();
      expect(screen.getByTestId("sector-warnings")).toBeInTheDocument();
    });
  });

  it("shows loading state initially", () => {
    render(ChartsTab, {
      props: { holdings, portfolioId: "p1", portfolioMode: "VALUE" },
    });

    expect(screen.getByText(/Loading chart data/)).toBeInTheDocument();
  });

  it("handles empty holdings without calling backend", async () => {
    mockGetHoldingSectors.mockClear();
    render(ChartsTab, {
      props: { holdings: [], portfolioId: "p1", portfolioMode: "VALUE" },
    });

    await waitFor(() => {
      expect(mockGetHoldingSectors).not.toHaveBeenCalled();
    });
  });
});
