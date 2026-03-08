import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { HoldingDetailResponse } from "../../lib/types";
import HoldingsTable from "./HoldingsTable.svelte";

const holdings: HoldingDetailResponse[] = [
  {
    id: "h1",
    ticker: "BBCA",
    avgBuyPrice: 8000,
    lots: 10,
    currentPrice: 9000,
    verdict: "UNDERVALUED",
  },
  {
    id: "h2",
    ticker: "BBRI",
    avgBuyPrice: 5000,
    lots: 5,
    currentPrice: 4500,
  },
];

describe("HoldingsTable", () => {
  it("renders all holdings", () => {
    render(HoldingsTable, {
      props: { holdings, onChecklist: vi.fn(), onRemove: vi.fn() },
    });
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("BBRI")).toBeInTheDocument();
  });

  it("shows positive P/L with profit styling", () => {
    render(HoldingsTable, {
      props: { holdings, onChecklist: vi.fn(), onRemove: vi.fn() },
    });
    const plCell = screen.getByTestId("pl-BBCA");
    expect(plCell.textContent).toContain("+");
    expect(plCell.className).toContain("text-profit");
  });

  it("shows negative P/L with loss styling", () => {
    render(HoldingsTable, {
      props: { holdings, onChecklist: vi.fn(), onRemove: vi.fn() },
    });
    const plCell = screen.getByTestId("pl-BBRI");
    expect(plCell.className).toContain("text-loss");
  });

  it("renders verdict badge for holdings with verdict", () => {
    render(HoldingsTable, {
      props: { holdings, onChecklist: vi.fn(), onRemove: vi.fn() },
    });
    expect(screen.getByText("Undervalued")).toBeInTheDocument();
  });

  it("calls onChecklist when Checklist button clicked", async () => {
    const user = userEvent.setup();
    const onChecklist = vi.fn();
    render(HoldingsTable, {
      props: { holdings, onChecklist, onRemove: vi.fn() },
    });

    const buttons = screen.getAllByRole("button", { name: /Checklist/i });
    await user.click(buttons[0]);
    expect(onChecklist).toHaveBeenCalledWith("BBCA");
  });

  it("renders table with aria-label", () => {
    render(HoldingsTable, {
      props: { holdings, onChecklist: vi.fn(), onRemove: vi.fn() },
    });
    expect(screen.getByRole("table", { name: "Holdings" })).toBeInTheDocument();
  });

  it("calls onRemove when trash button clicked", async () => {
    const user = userEvent.setup();
    const onRemove = vi.fn();
    render(HoldingsTable, {
      props: { holdings, onChecklist: vi.fn(), onRemove },
    });

    const removeButtons = screen.getAllByRole("button", {
      name: /Remove BBCA/i,
    });
    await user.click(removeButtons[0]);
    expect(onRemove).toHaveBeenCalledWith("h1", "BBCA");
  });
});
