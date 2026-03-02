import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import AddHoldingForm from "./AddHoldingForm.svelte";

const mockAddHolding = vi.fn();
vi.mock("../../wailsjs/go/backend/App", () => ({
  AddHolding: (...args: unknown[]) => mockAddHolding(...args),
}));

describe("AddHoldingForm", () => {
  const defaultProps = {
    portfolioId: "p1",
    onAdded: vi.fn(),
  };

  beforeEach(() => {
    mockAddHolding.mockReset();
    defaultProps.onAdded.mockReset();
  });

  it("renders all form fields", () => {
    render(AddHoldingForm, { props: defaultProps });

    expect(screen.getByLabelText(/ticker/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/buy price/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/lots/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/date/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /add holding/i })).toBeInTheDocument();
  });

  it("defaults date to today", () => {
    render(AddHoldingForm, { props: defaultProps });

    const dateInput = screen.getByLabelText(/date/i) as HTMLInputElement;
    const today = new Date().toISOString().split("T")[0];
    expect(dateInput.value).toBe(today);
  });

  it("submits valid form data with uppercased ticker", async () => {
    mockAddHolding.mockResolvedValueOnce({
      id: "h1",
      ticker: "BBCA",
      avgBuyPrice: 8500,
      lots: 10,
    });

    render(AddHoldingForm, { props: defaultProps });
    const user = userEvent.setup();

    await user.type(screen.getByLabelText(/ticker/i), "bbca");
    await user.clear(screen.getByLabelText(/buy price/i));
    await user.type(screen.getByLabelText(/buy price/i), "8500");
    await user.clear(screen.getByLabelText(/lots/i));
    await user.type(screen.getByLabelText(/lots/i), "10");
    await user.click(screen.getByRole("button", { name: /add holding/i }));

    const today = new Date().toISOString().split("T")[0];
    expect(mockAddHolding).toHaveBeenCalledWith("p1", "BBCA", 8500, 10, today);
    expect(defaultProps.onAdded).toHaveBeenCalled();
  });

  it("shows error when ticker is empty", async () => {
    render(AddHoldingForm, { props: defaultProps });
    const user = userEvent.setup();

    await user.clear(screen.getByLabelText(/buy price/i));
    await user.type(screen.getByLabelText(/buy price/i), "8500");
    await user.clear(screen.getByLabelText(/lots/i));
    await user.type(screen.getByLabelText(/lots/i), "10");
    await user.click(screen.getByRole("button", { name: /add holding/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/ticker is required/i);
    expect(mockAddHolding).not.toHaveBeenCalled();
  });

  it("shows error on API failure", async () => {
    mockAddHolding.mockRejectedValueOnce(new Error("duplicate ticker"));

    render(AddHoldingForm, { props: defaultProps });
    const user = userEvent.setup();

    await user.type(screen.getByLabelText(/ticker/i), "BBCA");
    await user.clear(screen.getByLabelText(/buy price/i));
    await user.type(screen.getByLabelText(/buy price/i), "8500");
    await user.clear(screen.getByLabelText(/lots/i));
    await user.type(screen.getByLabelText(/lots/i), "10");
    await user.click(screen.getByRole("button", { name: /add holding/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/duplicate ticker/i);
  });
});
