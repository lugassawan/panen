import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SellHoldingModal from "./SellHoldingModal.svelte";

const mockSellHolding = vi.fn();
vi.mock("../../../wailsjs/go/backend/App", () => ({
  SellHolding: (...args: unknown[]) => mockSellHolding(...args),
}));

describe("SellHoldingModal", () => {
  const defaultProps = {
    portfolioId: "p1",
    holdingId: "h1",
    ticker: "BBCA",
    maxLots: 10,
    avgBuyPrice: 8500,
    onSold: vi.fn(),
    onClose: vi.fn(),
  };

  beforeEach(() => {
    mockSellHolding.mockReset();
    defaultProps.onSold.mockReset();
    defaultProps.onClose.mockReset();
  });

  it("renders form fields and holding info", () => {
    render(SellHoldingModal, { props: defaultProps });

    expect(screen.getByLabelText(/sell price/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/lots/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/date/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sell/i })).toBeInTheDocument();
  });

  it("submits valid form data", async () => {
    mockSellHolding.mockResolvedValueOnce({
      id: "tx1",
      holdingId: "h1",
      price: 9000,
      lots: 5,
      fee: 100,
      tax: 50,
      realizedGain: 24850,
    });

    render(SellHoldingModal, { props: defaultProps });
    const user = userEvent.setup();

    await user.clear(screen.getByLabelText(/sell price/i));
    await user.type(screen.getByLabelText(/sell price/i), "9000");
    await user.clear(screen.getByLabelText(/lots/i));
    await user.type(screen.getByLabelText(/lots/i), "5");

    await user.click(screen.getByRole("button", { name: /^sell$/i }));

    const today = new Date().toISOString().split("T")[0];
    expect(mockSellHolding).toHaveBeenCalledWith("p1", "h1", 9000, 5, today);
    expect(defaultProps.onSold).toHaveBeenCalled();
  });

  it("shows error when sell price is zero", async () => {
    render(SellHoldingModal, { props: defaultProps });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /^sell$/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/sell price must be greater than 0/i);
    expect(mockSellHolding).not.toHaveBeenCalled();
  });

  it("shows error when lots exceeds max", async () => {
    render(SellHoldingModal, { props: defaultProps });
    const user = userEvent.setup();

    await user.clear(screen.getByLabelText(/sell price/i));
    await user.type(screen.getByLabelText(/sell price/i), "9000");
    await user.clear(screen.getByLabelText(/lots/i));
    await user.type(screen.getByLabelText(/lots/i), "11");

    await user.click(screen.getByRole("button", { name: /^sell$/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/lots must be between 1 and 10/i);
    expect(mockSellHolding).not.toHaveBeenCalled();
  });

  it("shows error on API failure", async () => {
    mockSellHolding.mockRejectedValueOnce(new Error("insufficient lots"));

    render(SellHoldingModal, { props: defaultProps });
    const user = userEvent.setup();

    await user.clear(screen.getByLabelText(/sell price/i));
    await user.type(screen.getByLabelText(/sell price/i), "9000");
    await user.clear(screen.getByLabelText(/lots/i));
    await user.type(screen.getByLabelText(/lots/i), "5");

    await user.click(screen.getByRole("button", { name: /^sell$/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/insufficient lots/i);
  });
});
