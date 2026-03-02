import { render, screen, within } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import App from "./App.svelte";

vi.mock("../wailsjs/go/backend/App", () => ({
  LookupStock: vi.fn(() => Promise.resolve({})),
  ListBrokerageAccounts: vi.fn(() => Promise.resolve([])),
  ListPortfolios: vi.fn(() => Promise.resolve([])),
  GetPortfolio: vi.fn(() => Promise.resolve({ portfolio: {}, holdings: [] })),
  CreateBrokerageAccount: vi.fn(() => Promise.resolve({})),
  CreatePortfolio: vi.fn(() => Promise.resolve({})),
  AddHolding: vi.fn(() => Promise.resolve({})),
}));

describe("App navigation", () => {
  it("renders sidebar with 3 nav items", () => {
    render(App);
    const nav = screen.getByRole("navigation", { name: /main/i });
    const buttons = within(nav).getAllByRole("button");
    expect(buttons).toHaveLength(3);
    expect(buttons[0]).toHaveTextContent("Stock Lookup");
    expect(buttons[1]).toHaveTextContent("Portfolio");
    expect(buttons[2]).toHaveTextContent("Settings");
  });

  it("starts on Stock Lookup page by default", () => {
    render(App);
    expect(screen.getByLabelText("Stock ticker")).toBeInTheDocument();
    expect(screen.getByLabelText("Risk profile")).toBeInTheDocument();
  });

  it("shows Stock Lookup as active nav item by default", () => {
    render(App);
    const nav = screen.getByRole("navigation", { name: /main/i });
    const buttons = within(nav).getAllByRole("button");
    expect(buttons[0]).toHaveAttribute("aria-current", "page");
  });

  it("switches to Portfolio page when clicking Portfolio nav", async () => {
    const user = userEvent.setup();
    render(App);

    const nav = screen.getByRole("navigation", { name: /main/i });
    await user.click(within(nav).getByText("Portfolio"));

    expect(await screen.findByText("Set Up Your Brokerage")).toBeInTheDocument();
    expect(screen.queryByLabelText("Stock ticker")).not.toBeInTheDocument();
  });

  it("switches to Settings page when clicking Settings nav", async () => {
    const user = userEvent.setup();
    render(App);

    const nav = screen.getByRole("navigation", { name: /main/i });
    await user.click(within(nav).getByText("Settings"));

    expect(screen.getByLabelText("Language")).toBeInTheDocument();
    expect(screen.getByLabelText("Theme")).toBeInTheDocument();
    expect(screen.getByText("Coming in a future update")).toBeInTheDocument();
  });

  it("returns to Stock Lookup when clicking Lookup nav from another page", async () => {
    const user = userEvent.setup();
    render(App);

    const nav = screen.getByRole("navigation", { name: /main/i });
    await user.click(within(nav).getByText("Settings"));
    expect(screen.queryByLabelText("Stock ticker")).not.toBeInTheDocument();

    await user.click(within(nav).getByText("Stock Lookup"));
    expect(screen.getByLabelText("Stock ticker")).toBeInTheDocument();
  });

  it("updates active nav styling on page switch", async () => {
    const user = userEvent.setup();
    render(App);

    const nav = screen.getByRole("navigation", { name: /main/i });
    const buttons = within(nav).getAllByRole("button");
    const [lookupBtn, portfolioBtn] = buttons;

    expect(lookupBtn).toHaveAttribute("aria-current", "page");
    expect(portfolioBtn).not.toHaveAttribute("aria-current");

    await user.click(portfolioBtn);

    expect(portfolioBtn).toHaveAttribute("aria-current", "page");
    expect(lookupBtn).not.toHaveAttribute("aria-current");
  });
});
