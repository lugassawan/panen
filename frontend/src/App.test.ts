import { render, screen, within } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import App from "./App.svelte";

vi.mock("../wailsjs/go/backend/App", () => ({
  LookupStock: vi.fn(() => Promise.resolve({})),
  ListBrokerageAccounts: vi.fn(() => Promise.resolve([])),
  ListBrokerConfigs: vi.fn(() => Promise.resolve([])),
  ListPortfolios: vi.fn(() => Promise.resolve([])),
  GetPortfolio: vi.fn(() => Promise.resolve({ portfolio: {}, holdings: [] })),
  CreateBrokerageAccount: vi.fn(() => Promise.resolve({})),
  UpdateBrokerageAccount: vi.fn(() => Promise.resolve({})),
  DeleteBrokerageAccount: vi.fn(() => Promise.resolve()),
  CreatePortfolio: vi.fn(() => Promise.resolve({})),
  AddHolding: vi.fn(() => Promise.resolve({})),
  ListWatchlists: vi.fn(() => Promise.resolve([])),
  ListIndexNames: vi.fn(() => Promise.resolve([])),
  ListWatchlistSectors: vi.fn(() => Promise.resolve([])),
  CreateWatchlist: vi.fn(() => Promise.resolve({})),
  DeleteWatchlist: vi.fn(() => Promise.resolve()),
  AddToWatchlist: vi.fn(() => Promise.resolve()),
  RemoveFromWatchlist: vi.fn(() => Promise.resolve()),
  GetWatchlistItems: vi.fn(() => Promise.resolve([])),
  GetPresetItems: vi.fn(() => Promise.resolve([])),
  TriggerRefresh: vi.fn(() => Promise.resolve()),
  GetRefreshStatus: vi.fn(() => Promise.resolve({ state: "idle", lastRefresh: "" })),
  GetRefreshSettings: vi.fn(() =>
    Promise.resolve({
      autoRefreshEnabled: true,
      intervalMinutes: 720,
      lastRefreshedAt: "",
    }),
  ),
  UpdateRefreshSettings: vi.fn(() => Promise.resolve()),
}));

vi.mock("../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

vi.mock("./lib/stores/sync.svelte", () => ({
  sync: {
    state: "idle",
    isSyncing: false,
    lastRefresh: "",
    currentTicker: null,
    progress: null,
    progressPercent: 0,
    lastSummary: null,
    hasError: false,
    errorMessage: null,
  },
}));

vi.mock("./lib/stores/theme.svelte", () => ({
  theme: {
    current: "light",
    preference: "system",
    isDark: false,
    set: vi.fn(),
    toggle: vi.fn(),
  },
}));

describe("App navigation", () => {
  it("renders sidebar with 5 nav items", () => {
    render(App);
    const nav = screen.getByRole("navigation", { name: /main/i });
    const buttons = within(nav).getAllByRole("button");
    expect(buttons).toHaveLength(5);
    expect(buttons[0]).toHaveTextContent("Stock Lookup");
    expect(buttons[1]).toHaveTextContent("Watchlist");
    expect(buttons[2]).toHaveTextContent("Portfolio");
    expect(buttons[3]).toHaveTextContent("Brokerage");
    expect(buttons[4]).toHaveTextContent("Settings");
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

  it("switches to Brokerage page when clicking Brokerage nav", async () => {
    const user = userEvent.setup();
    render(App);

    const nav = screen.getByRole("navigation", { name: /main/i });
    await user.click(within(nav).getByText("Brokerage"));

    expect(await screen.findByText("Brokerage Accounts")).toBeInTheDocument();
    expect(screen.queryByLabelText("Stock ticker")).not.toBeInTheDocument();
  });

  it("switches to Settings page when clicking Settings nav", async () => {
    const user = userEvent.setup();
    render(App);

    const nav = screen.getByRole("navigation", { name: /main/i });
    await user.click(within(nav).getByText("Settings"));

    expect(screen.getByLabelText("Language")).toBeInTheDocument();
    expect(screen.getByText("Theme")).toBeInTheDocument();
    expect(screen.getByText("Language selection coming in a future update")).toBeInTheDocument();
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
    const [lookupBtn, , portfolioBtn] = buttons;

    expect(lookupBtn).toHaveAttribute("aria-current", "page");
    expect(portfolioBtn).not.toHaveAttribute("aria-current");

    await user.click(portfolioBtn);

    expect(portfolioBtn).toHaveAttribute("aria-current", "page");
    expect(lookupBtn).not.toHaveAttribute("aria-current");
  });
});
