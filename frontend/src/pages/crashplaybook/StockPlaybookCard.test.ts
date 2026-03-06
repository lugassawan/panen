import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";
import type { StockPlaybookResponse } from "../../lib/types";
import StockPlaybookCard from "./StockPlaybookCard.svelte";

function makeStock(overrides: Partial<StockPlaybookResponse> = {}): StockPlaybookResponse {
  return {
    ticker: "BBCA",
    currentPrice: 8500,
    entryPrice: 7500,
    levels: [
      { level: "NORMAL_DIP", triggerPrice: 7500, deployPct: 30 },
      { level: "CRASH", triggerPrice: 6250, deployPct: 40 },
      { level: "EXTREME", triggerPrice: 5250, deployPct: 30 },
    ],
    ...overrides,
  };
}

describe("StockPlaybookCard", () => {
  it("renders ticker and current price", () => {
    render(StockPlaybookCard, { props: { stock: makeStock(), onDiagnostic: vi.fn() } });
    expect(screen.getByText("BBCA")).toBeTruthy();
    expect(document.body.textContent).toContain("8,500");
  });

  it("renders 3 response levels", () => {
    render(StockPlaybookCard, { props: { stock: makeStock(), onDiagnostic: vi.fn() } });
    expect(screen.getByText("Normal Dip")).toBeTruthy();
    expect(screen.getByText("Crash")).toBeTruthy();
    expect(screen.getByText("Extreme")).toBeTruthy();
  });

  it("shows Level Hit badge when active level is set", () => {
    render(StockPlaybookCard, {
      props: { stock: makeStock({ activeLevel: "NORMAL_DIP" }), onDiagnostic: vi.fn() },
    });
    expect(screen.getByText("Level Hit")).toBeTruthy();
  });

  it("shows Run Diagnostic button when active level is set", () => {
    render(StockPlaybookCard, {
      props: { stock: makeStock({ activeLevel: "CRASH" }), onDiagnostic: vi.fn() },
    });
    expect(screen.getByText("Run Diagnostic")).toBeTruthy();
  });

  it("hides Run Diagnostic button when no active level", () => {
    render(StockPlaybookCard, { props: { stock: makeStock(), onDiagnostic: vi.fn() } });
    expect(screen.queryByText("Run Diagnostic")).toBeNull();
  });
});
