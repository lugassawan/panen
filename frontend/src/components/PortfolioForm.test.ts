import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PortfolioForm from "./PortfolioForm.svelte";

const mockCreatePortfolio = vi.fn();
vi.mock("../../wailsjs/go/backend/App", () => ({
  CreatePortfolio: (...args: unknown[]) => mockCreatePortfolio(...args),
}));

describe("PortfolioForm", () => {
  const defaultProps = {
    brokerageAcctId: "b1",
    onCreated: vi.fn(),
  };

  beforeEach(() => {
    mockCreatePortfolio.mockReset();
    defaultProps.onCreated.mockReset();
  });

  it("renders all form fields with defaults", () => {
    render(PortfolioForm, { props: defaultProps });

    expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/conservative/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/moderate/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/aggressive/i)).toBeInTheDocument();
    expect(screen.getByLabelText("Capital")).toBeInTheDocument();
    expect(screen.getByLabelText(/monthly addition/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/max stocks/i)).toBeInTheDocument();
  });

  it("defaults risk profile to MODERATE", () => {
    render(PortfolioForm, { props: defaultProps });

    const moderate = screen.getByLabelText(/moderate/i) as HTMLInputElement;
    expect(moderate.checked).toBe(true);
  });

  it("shows risk profile descriptions", () => {
    render(PortfolioForm, { props: defaultProps });

    expect(screen.getByText(/stricter margin of safety/i)).toBeInTheDocument();
    expect(screen.getByText(/balanced approach/i)).toBeInTheDocument();
    expect(screen.getByText(/lower margin of safety threshold/i)).toBeInTheDocument();
  });

  it("submits valid form data with VALUE mode", async () => {
    mockCreatePortfolio.mockResolvedValueOnce({
      id: "p1",
      name: "My Portfolio",
    });

    render(PortfolioForm, { props: defaultProps });
    const user = userEvent.setup();

    await user.type(screen.getByLabelText(/portfolio name/i), "My Portfolio");
    await user.clear(screen.getByLabelText("Capital"));
    await user.type(screen.getByLabelText("Capital"), "10000000");
    await user.clear(screen.getByLabelText(/monthly addition/i));
    await user.type(screen.getByLabelText(/monthly addition/i), "1000000");
    await user.clear(screen.getByLabelText(/max stocks/i));
    await user.type(screen.getByLabelText(/max stocks/i), "10");
    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(mockCreatePortfolio).toHaveBeenCalledWith(
      "b1",
      "My Portfolio",
      "VALUE",
      "MODERATE",
      10000000,
      1000000,
      10,
    );
    expect(defaultProps.onCreated).toHaveBeenCalled();
  });

  it("shows error when name is empty", async () => {
    render(PortfolioForm, { props: defaultProps });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/portfolio name is required/i);
    expect(mockCreatePortfolio).not.toHaveBeenCalled();
  });

  it("shows error on API failure", async () => {
    mockCreatePortfolio.mockRejectedValueOnce(new Error("db error"));

    render(PortfolioForm, { props: defaultProps });
    const user = userEvent.setup();

    await user.type(screen.getByLabelText(/portfolio name/i), "My Portfolio");
    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/db error/i);
  });
});
