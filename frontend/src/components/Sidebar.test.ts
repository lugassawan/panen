import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

vi.mock("../i18n", () => ({
  locale: { current: "en" },
  t: (key: string) => {
    const keys: Record<string, string> = {
      "nav.lookup": "Stock Lookup",
      "nav.watchlist": "Watchlist",
      "nav.screener": "Screener",
      "nav.portfolio": "Portfolio",
      "nav.payday": "Payday",
      "nav.crashPlaybook": "Crash Playbook",
      "nav.brokerage": "Brokerage",
      "nav.settings": "Settings",
      "nav.searchPages": "Search pages...",
      "nav.noResults": "No results found",
    };
    return keys[key] ?? key;
  },
}));

vi.mock("../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

vi.mock("../../wailsjs/go/backend/App", () => ({
  TriggerRefresh: vi.fn(),
}));

import Sidebar from "./Sidebar.svelte";

describe("Sidebar", () => {
  it("renders all navigation items", () => {
    render(Sidebar, {
      props: { currentPage: "portfolio", onNavigate: vi.fn() },
    });

    expect(screen.getByRole("button", { name: /Stock Lookup/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Watchlist/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Screener/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Portfolio/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Payday/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Crash Playbook/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Brokerage/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Settings/i })).toBeInTheDocument();
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
