import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { PortfolioResponse } from "../../lib/types";
import PortfolioList from "./PortfolioList.svelte";

const portfolios: PortfolioResponse[] = [
  {
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
];

describe("PortfolioList", () => {
  const defaultHandlers = {
    onView: vi.fn(),
    onEdit: vi.fn(),
    onDelete: vi.fn(),
    onCreate: vi.fn(),
  };

  it("renders portfolio name and mode badge", () => {
    render(PortfolioList, {
      props: { portfolios, ...defaultHandlers },
    });
    expect(screen.getByText("Growth")).toBeInTheDocument();
    expect(screen.getByTestId("mode-badge")).toHaveTextContent("Value");
  });

  it("renders New Portfolio button when less than 2 portfolios", () => {
    render(PortfolioList, {
      props: { portfolios, ...defaultHandlers },
    });
    expect(screen.getByRole("button", { name: /New Portfolio/i })).toBeInTheDocument();
  });

  it("hides New Portfolio button when 2 portfolios", () => {
    const twoPortfolios: PortfolioResponse[] = [
      ...portfolios,
      { ...portfolios[0], id: "p2", name: "Income", mode: "DIVIDEND" },
    ];
    render(PortfolioList, {
      props: { portfolios: twoPortfolios, ...defaultHandlers },
    });
    expect(screen.queryByRole("button", { name: /New Portfolio/i })).not.toBeInTheDocument();
  });

  it("calls onView when portfolio card clicked", async () => {
    const user = userEvent.setup();
    const onView = vi.fn();
    render(PortfolioList, {
      props: { portfolios, ...defaultHandlers, onView },
    });

    await user.click(screen.getByText("Growth"));
    expect(onView).toHaveBeenCalledWith(portfolios[0]);
  });

  it("calls onEdit when Edit button clicked", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    render(PortfolioList, {
      props: { portfolios, ...defaultHandlers, onEdit },
    });

    await user.click(screen.getByRole("button", { name: /Edit/i }));
    expect(onEdit).toHaveBeenCalledWith(portfolios[0]);
  });

  it("calls onDelete when Delete button clicked", async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    render(PortfolioList, {
      props: { portfolios, ...defaultHandlers, onDelete },
    });

    await user.click(screen.getByRole("button", { name: /Delete/i }));
    expect(onDelete).toHaveBeenCalledWith(portfolios[0]);
  });
});
