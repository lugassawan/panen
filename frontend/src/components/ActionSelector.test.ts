import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ActionSelector from "./ActionSelector.svelte";

const mockAvailableActions = vi.fn();

vi.mock("../../wailsjs/go/backend/App", () => ({
  AvailableActions: (...args: unknown[]) => mockAvailableActions(...args),
}));

describe("ActionSelector", () => {
  const defaultProps = {
    portfolioId: "p1",
    ticker: "BBCA",
    onselect: vi.fn(),
  };

  beforeEach(() => {
    mockAvailableActions.mockReset();
    defaultProps.onselect = vi.fn();
  });

  it("shows loading state initially", () => {
    mockAvailableActions.mockReturnValue(new Promise(() => {}));
    render(ActionSelector, { props: defaultProps });

    expect(screen.getByText(/loading actions/i)).toBeInTheDocument();
  });

  it("renders action chips when AvailableActions returns data", async () => {
    mockAvailableActions.mockResolvedValue(["BUY", "HOLD"]);
    render(ActionSelector, { props: defaultProps });

    expect(await screen.findByText("Buy")).toBeInTheDocument();
    expect(screen.getByText("Hold")).toBeInTheDocument();
    expect(screen.getAllByTestId("action-chip")).toHaveLength(2);
  });

  it("auto-selects first action on load", async () => {
    mockAvailableActions.mockResolvedValue(["AVERAGE_DOWN", "HOLD"]);
    render(ActionSelector, { props: defaultProps });

    await screen.findByText("Average Down");
    expect(defaultProps.onselect).toHaveBeenCalledWith("AVERAGE_DOWN");
  });

  it("calls onselect when chip is clicked", async () => {
    mockAvailableActions.mockResolvedValue(["BUY", "HOLD", "SELL_EXIT"]);
    render(ActionSelector, { props: defaultProps });
    const user = userEvent.setup();

    await screen.findByText("Buy");
    await user.click(screen.getByText("Hold"));

    expect(defaultProps.onselect).toHaveBeenCalledWith("HOLD");
  });

  it("highlights selected chip", async () => {
    mockAvailableActions.mockResolvedValue(["BUY", "HOLD"]);
    render(ActionSelector, { props: defaultProps });
    const user = userEvent.setup();

    await screen.findByText("Buy");
    const holdChip = screen.getByText("Hold");
    await user.click(holdChip);

    expect(holdChip.getAttribute("aria-pressed")).toBe("true");
    expect(screen.getByText("Buy").getAttribute("aria-pressed")).toBe("false");
  });

  it("shows error on failure", async () => {
    mockAvailableActions.mockRejectedValue(new Error("network error"));
    render(ActionSelector, { props: defaultProps });

    expect(await screen.findByRole("alert")).toHaveTextContent(/network error/);
  });

  it("passes portfolioId and ticker to AvailableActions", async () => {
    mockAvailableActions.mockResolvedValue(["BUY"]);
    render(ActionSelector, { props: defaultProps });

    await screen.findByText("Buy");
    expect(mockAvailableActions).toHaveBeenCalledWith("p1", "BBCA");
  });
});
