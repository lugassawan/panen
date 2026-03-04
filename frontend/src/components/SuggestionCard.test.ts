import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import type { SuggestionResponse } from "../lib/types";
import SuggestionCard from "./SuggestionCard.svelte";

function makeBuySuggestion(overrides: Partial<SuggestionResponse> = {}): SuggestionResponse {
  return {
    action: "BUY",
    ticker: "BBCA",
    lots: 5,
    pricePerShare: 9250,
    grossCost: 4625000,
    fee: 6938,
    tax: 0,
    netCost: 4631938,
    newAvgBuyPrice: 9100,
    newPositionLots: 15,
    newPositionPct: 25.5,
    capitalGainPct: 0,
    ...overrides,
  };
}

function makeSellSuggestion(overrides: Partial<SuggestionResponse> = {}): SuggestionResponse {
  return {
    action: "SELL_EXIT",
    ticker: "BBRI",
    lots: 10,
    pricePerShare: 5500,
    grossCost: 5500000,
    fee: 8250,
    tax: 5500,
    netCost: 5486250,
    newAvgBuyPrice: 0,
    newPositionLots: 0,
    newPositionPct: 0,
    capitalGainPct: 15.3,
    ...overrides,
  };
}

describe("SuggestionCard", () => {
  it("renders buy suggestion with cost details", () => {
    render(SuggestionCard, { props: { suggestion: makeBuySuggestion() } });

    expect(screen.getByTestId("suggestion-card")).toBeInTheDocument();
    expect(screen.getByText(/Trade Suggestion: Buy/)).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
    expect(screen.getByText("Gross Cost")).toBeInTheDocument();
    expect(screen.getByText("Net Cost")).toBeInTheDocument();
    expect(screen.getByText("New Avg Price")).toBeInTheDocument();
    expect(screen.getByText(/15 lots/)).toBeInTheDocument();
  });

  it("renders sell suggestion with capital gain", () => {
    render(SuggestionCard, {
      props: { suggestion: makeSellSuggestion() },
    });

    expect(screen.getByText(/Trade Suggestion: Sell \(Exit\)/)).toBeInTheDocument();
    expect(screen.getByText("Gross Proceeds")).toBeInTheDocument();
    expect(screen.getByText("Net Proceeds")).toBeInTheDocument();
    expect(screen.getByText("Capital Gain")).toBeInTheDocument();
    expect(screen.getByText(/15,30%/)).toBeInTheDocument();
  });

  it("shows tax only when > 0", () => {
    const { rerender } = render(SuggestionCard, {
      props: { suggestion: makeBuySuggestion({ tax: 0 }) },
    });

    expect(screen.queryByText("Tax")).not.toBeInTheDocument();

    rerender({ suggestion: makeSellSuggestion({ tax: 5500 }) });
    expect(screen.getByText("Tax")).toBeInTheDocument();
  });

  it("uses font-mono for financial numbers", () => {
    render(SuggestionCard, { props: { suggestion: makeBuySuggestion() } });

    const monoElements = screen.getByTestId("suggestion-card").querySelectorAll(".font-mono");
    expect(monoElements.length).toBeGreaterThan(0);
  });

  it("shows positive capital gain with text-profit", () => {
    render(SuggestionCard, {
      props: { suggestion: makeSellSuggestion({ capitalGainPct: 15.3 }) },
    });

    const gainEl = screen.getByText(/15,30%/);
    expect(gainEl.className).toContain("text-profit");
  });

  it("shows negative capital gain with text-loss", () => {
    render(SuggestionCard, {
      props: { suggestion: makeSellSuggestion({ capitalGainPct: -8.5 }) },
    });

    const gainEl = screen.getByText(/8,50%/);
    expect(gainEl.className).toContain("text-loss");
  });
});
