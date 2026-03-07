import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";
import type { DashboardOverviewResponse } from "../../lib/types";

vi.mock("../../lib/stores/theme.svelte", () => ({
  theme: {
    current: "light",
    preference: "system",
    isDark: false,
    set: vi.fn(),
    toggle: vi.fn(),
  },
}));

import DashboardPage from "./DashboardPage.svelte";

const emptyOverview: DashboardOverviewResponse = {
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
};

const readyOverview: DashboardOverviewResponse = {
  totalMarketValue: 16000000,
  totalCostBasis: 14000000,
  totalPlAmount: 2000000,
  totalPlPercent: 14.28,
  totalDividendIncome: 480000,
  portfolios: [
    {
      id: "p1",
      name: "Value",
      mode: "VALUE",
      marketValue: 16000000,
      costBasis: 14000000,
      plAmount: 2000000,
      plPercent: 14.28,
      weight: 100,
    },
  ],
  topGainers: [
    {
      ticker: "BBCA",
      portfolioId: "p1",
      portfolioName: "Value",
      marketValue: 9000000,
      costBasis: 8000000,
      plAmount: 1000000,
      plPercent: 12.5,
    },
  ],
  topLosers: [],
  portfolioAllocation: [{ label: "Value", value: 16000000, pct: 100 }],
  sectorAllocation: [{ label: "Banking", value: 16000000, pct: 100 }],
  recentTransactions: [],
};

const mockGetDashboardOverview = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetDashboardOverview: (...args: unknown[]) => mockGetDashboardOverview(...args),
}));

describe("DashboardPage", () => {
  it("shows empty state when no portfolios", async () => {
    mockGetDashboardOverview.mockResolvedValue(emptyOverview);

    render(DashboardPage, { props: { onNavigate: vi.fn() } });

    const emptyTitle = await screen.findByText("No portfolios yet");
    expect(emptyTitle).toBeTruthy();
  });

  it("shows dashboard data when portfolios exist", async () => {
    mockGetDashboardOverview.mockResolvedValue(readyOverview);

    render(DashboardPage, { props: { onNavigate: vi.fn() } });

    const title = await screen.findByText("Dashboard");
    expect(title).toBeTruthy();
  });
});
