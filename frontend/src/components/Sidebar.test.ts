import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

vi.mock("../i18n", () => ({
  locale: { current: "en" },
  t: (key: string) => {
    const keys: Record<string, string> = {
      "nav.dashboard": "Dashboard",
      "nav.lookup": "Stock Lookup",
      "nav.watchlist": "Watchlist",
      "nav.screener": "Screener",
      "nav.portfolio": "Portfolio",
      "nav.payday": "Payday",
      "nav.crashPlaybook": "Crash Playbook",
      "nav.alerts": "Alerts",
      "nav.brokerage": "Brokerage",
      "nav.comparison": "Compare",
      "nav.settings": "Settings",
      "nav.noResults": "No results found",
      "nav.group.overview": "Overview",
      "nav.group.research": "Research",
      "nav.group.portfolio": "Portfolio",
      "nav.group.account": "Account",
    };
    return keys[key] ?? key;
  },
}));

vi.mock("../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

vi.mock("../../wailsjs/go/backend/App", () => ({
  TriggerRefresh: vi.fn(),
  GetAlertCount: vi.fn(() => Promise.resolve(0)),
  GetActiveAlerts: vi.fn(() => Promise.resolve([])),
  GetAlertsByTicker: vi.fn(() => Promise.resolve([])),
  AcknowledgeAlert: vi.fn(() => Promise.resolve()),
}));

import Sidebar from "./Sidebar.svelte";

describe("Sidebar", () => {
  it("renders all navigation items", () => {
    render(Sidebar, {
      props: { currentPage: "portfolio", onNavigate: vi.fn() },
    });

    expect(screen.getByRole("button", { name: /Dashboard/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Stock Lookup/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Watchlist/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Screener/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Portfolio/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Payday/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Crash Playbook/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Alerts/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Brokerage/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Settings/i })).toBeInTheDocument();
  });

  it("renders group headers", () => {
    render(Sidebar, {
      props: { currentPage: "dashboard", onNavigate: vi.fn() },
    });

    expect(screen.getByText("Overview")).toBeInTheDocument();
    expect(screen.getByText("Research")).toBeInTheDocument();
    // "Portfolio" appears as both a group header and a nav button — verify at least 2
    const portfolioElements = screen.getAllByText("Portfolio");
    expect(portfolioElements.length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText("Account")).toBeInTheDocument();
  });

  it("marks current page with aria-current", () => {
    render(Sidebar, {
      props: { currentPage: "portfolio", onNavigate: vi.fn() },
    });

    const portfolioBtn = screen.getByRole("button", {
      name: /Portfolio/i,
    });
    expect(portfolioBtn).toHaveAttribute("aria-current", "page");

    const watchlistBtn = screen.getByRole("button", {
      name: /Watchlist/i,
    });
    expect(watchlistBtn).not.toHaveAttribute("aria-current");
  });

  it("calls onNavigate with page when clicked", async () => {
    const user = userEvent.setup();
    const onNavigate = vi.fn();
    render(Sidebar, { props: { currentPage: "portfolio", onNavigate } });

    await user.click(screen.getByRole("button", { name: /Watchlist/i }));
    expect(onNavigate).toHaveBeenCalledWith("watchlist");
  });

  it("renders Panen logo text", () => {
    render(Sidebar, {
      props: { currentPage: "portfolio", onNavigate: vi.fn() },
    });
    expect(screen.getByText("Panen")).toBeInTheDocument();
  });
});
