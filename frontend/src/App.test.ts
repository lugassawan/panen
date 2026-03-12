import { render, screen, within } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

vi.mock("./i18n", () => ({
  locale: { current: "en" },
  t: (key: string, params?: Record<string, string | number>) => {
    const keys: Record<string, string> = {
      "nav.dashboard": "Dashboard",
      "nav.lookup": "Stock Lookup",
      "nav.watchlist": "Watchlist",
      "nav.screener": "Screener",
      "nav.comparison": "Compare",
      "nav.portfolio": "Portfolio",
      "nav.payday": "Payday",
      "nav.crashPlaybook": "Crash Playbook",
      "nav.transactions": "Transactions",
      "nav.alerts": "Alerts",
      "nav.brokerage": "Brokerage",
      "nav.settings": "Settings",
      "nav.search": "Search",
      "nav.searchPages": "Search pages...",
      "nav.noResults": "No results found",
      "comparison.title": "Stock Comparison",
      "comparison.subtitle": "Compare 2-4 stocks side by side",
      "comparison.emptyTitle": "Compare Stocks",
      "comparison.emptyDescription":
        "Enter at least 2 tickers and click Compare to see side-by-side metrics.",
      "settings.title": "Settings",
      "settings.language": "Language",
      "settings.english": "English",
      "settings.indonesian": "Bahasa Indonesia",
      "settings.theme": "Theme",
      "settings.dataRefresh": "Data Refresh",
      "settings.autoRefresh": "Auto Refresh",
      "settings.autoRefreshTooltip":
        "Automatically refresh stock data in the background at the configured interval",
      "settings.refreshInterval": "Refresh Interval",
      "settings.refreshNow": "Refresh Now",
      "settings.syncing": "Syncing...",
      "settings.about": "About",
      "settings.version": "Version",
      "settings.checkForUpdates": "Check for Updates",
      "settings.settingsSaved": "Settings saved",
      "screener.title": "Stock Screener",
      "screener.subtitle": "Screen stocks against fundamental criteria by risk profile",
      "screener.configurePrompt": "Configure and run a screen",
      "screener.configureDescription":
        'Select a universe, choose a risk profile, and click "Run Screen" to discover stocks.',
      "screener.universe": "Universe",
      "screener.index": "Index",
      "screener.selectIndex": "Select index...",
      "screener.riskProfile": "Risk Profile",
      "screener.conservative": "Conservative",
      "screener.moderate": "Moderate",
      "screener.aggressive": "Aggressive",
      "screener.conservativeThresholds": "ROE > 15%, DER < 0.8",
      "screener.moderateThresholds": "ROE > 12%, DER < 1.0",
      "screener.aggressiveThresholds": "ROE > 8%, DER < 1.5",
      "screener.runScreen": "Run Screen",
      "screener.sectorFilter": "Sector Filter",
      "screener.allSectors": "All sectors",
      "screener.sector": "Sector",
      "screener.custom": "Custom",
      "portfolio.title": "Portfolios",
      "portfolio.loading": "Loading portfolio...",
      "portfolio.setupBrokerage": "Set Up Your Brokerage",
      "portfolio.newPortfolio": "New Portfolio",
      "portfolio.createPortfolio": "Create Your Portfolio",
      "brokerage.title": "Brokerage Accounts",
      "brokerage.addAccount": "Add Account",
      "brokerage.loading": "Loading accounts...",
      "brokerage.noAccounts": "No brokerage accounts yet",
      "brokerage.noAccountsDesc":
        "Add your first brokerage account to start tracking fees and managing portfolios.",
      "dashboard.title": "Dashboard",
      "dashboard.emptyTitle": "No portfolios yet",
      "dashboard.empty":
        "Set up a brokerage account and create your first portfolio to see your dashboard.",
      "dashboard.goToBrokerage": "Set Up Brokerage",
      "common.retry": "Retry",
      "common.loading": "Loading...",
    };
    let value = keys[key] ?? key;
    if (params) {
      value = value.replace(/\{(\w+)\}/g, (_, name) => String(params[name] ?? `{${name}}`));
    }
    return value;
  },
}));

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
  RunScreen: vi.fn(() => Promise.resolve([])),
  ListScreenerIndices: vi.fn(() => Promise.resolve([])),
  ListScreenerSectors: vi.fn(() => Promise.resolve([])),
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
  GetHoldingSectors: vi.fn(() => Promise.resolve({})),
  GetAlertCount: vi.fn(() => Promise.resolve(0)),
  GetActiveAlerts: vi.fn(() => Promise.resolve([])),
  GetAlertsByTicker: vi.fn(() => Promise.resolve([])),
  AcknowledgeAlert: vi.fn(() => Promise.resolve()),
  ListTransactions: vi.fn(() => Promise.resolve({ items: [], summary: {} })),
  GetDashboardOverview: vi.fn(() =>
    Promise.resolve({
      totalMarketValue: 0,
      totalCostBasis: 0,
      totalPlAmount: 0,
      totalPlPercent: 0,
      totalDividendIncome: 0,
      portfolios: [],
      topGainers: [],
      topLosers: [],
      portfolioAllocation: [],
      sectorAllocation: [],
      recentTransactions: [],
    }),
  ),
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
  it("renders sidebar with 13 nav items including search hint", () => {
    render(App);
    const nav = screen.getByRole("navigation", { name: /main/i });
    const buttons = within(nav).getAllByRole("button");
    expect(buttons).toHaveLength(13);
    expect(buttons[0]).toHaveTextContent("Dashboard");
    expect(buttons[1]).toHaveTextContent("Stock Lookup");
    expect(buttons[2]).toHaveTextContent("Watchlist");
    expect(buttons[3]).toHaveTextContent("Screener");
    expect(buttons[4]).toHaveTextContent("Compare");
    expect(buttons[5]).toHaveTextContent("Portfolio");
    expect(buttons[6]).toHaveTextContent("Payday");
    expect(buttons[7]).toHaveTextContent("Crash Playbook");
    expect(buttons[8]).toHaveTextContent("Transactions");
    expect(buttons[9]).toHaveTextContent("Alerts");
    expect(buttons[10]).toHaveTextContent("Brokerage");
    expect(buttons[11]).toHaveTextContent("Search");
    expect(buttons[12]).toHaveTextContent("Settings");
  });

  it("starts on Dashboard page by default", async () => {
    render(App);
    expect(await screen.findByText("No portfolios yet")).toBeInTheDocument();
  });

  it("shows Dashboard as active nav item by default", () => {
    render(App);
    const nav = screen.getByRole("navigation", { name: /main/i });
    const buttons = within(nav).getAllByRole("button");
    expect(buttons[0]).toHaveAttribute("aria-current", "page");
  });

  it("switches to Screener page when clicking Screener nav", async () => {
    const user = userEvent.setup();
    render(App);

    const nav = screen.getByRole("navigation", { name: /main/i });
    await user.click(within(nav).getByText("Screener"));

    expect(await screen.findByText("Stock Screener")).toBeInTheDocument();
    expect(screen.queryByLabelText("Stock ticker")).not.toBeInTheDocument();
  });

  it("switches to Compare page when clicking Compare nav", async () => {
    const user = userEvent.setup();
    render(App);

    const nav = screen.getByRole("navigation", { name: /main/i });
    await user.click(within(nav).getByText("Compare"));

    expect(await screen.findByText("Stock Comparison")).toBeInTheDocument();
    expect(screen.queryByLabelText("Stock ticker")).not.toBeInTheDocument();
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
    const [dashboardBtn, , , , , portfolioBtn] = buttons;

    expect(dashboardBtn).toHaveAttribute("aria-current", "page");
    expect(portfolioBtn).not.toHaveAttribute("aria-current");

    await user.click(portfolioBtn);

    expect(portfolioBtn).toHaveAttribute("aria-current", "page");
    expect(dashboardBtn).not.toHaveAttribute("aria-current");
  });
});
