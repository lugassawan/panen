import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { MonthlyPaydayResponse } from "../../lib/types";
import PaydayPage from "./PaydayPage.svelte";

const mockGetPaydayDay = vi.fn();
const mockSavePaydayDay = vi.fn();
const mockGetCurrentMonthStatus = vi.fn();
const mockConfirmPayday = vi.fn();
const mockDeferPayday = vi.fn();
const mockSkipPayday = vi.fn();
const mockGetCashFlowSummary = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetPaydayDay: (...args: unknown[]) => mockGetPaydayDay(...args),
  SavePaydayDay: (...args: unknown[]) => mockSavePaydayDay(...args),
  GetCurrentMonthStatus: (...args: unknown[]) => mockGetCurrentMonthStatus(...args),
  ConfirmPayday: (...args: unknown[]) => mockConfirmPayday(...args),
  DeferPayday: (...args: unknown[]) => mockDeferPayday(...args),
  SkipPayday: (...args: unknown[]) => mockSkipPayday(...args),
  GetCashFlowSummary: (...args: unknown[]) => mockGetCashFlowSummary(...args),
}));

function makeMonthlyStatus(overrides: Partial<MonthlyPaydayResponse> = {}): MonthlyPaydayResponse {
  return {
    month: "2026-03",
    paydayDay: 25,
    portfolios: [
      {
        portfolioId: "p1",
        portfolioName: "Growth",
        mode: "VALUE",
        expected: 5000000,
        actual: 0,
        status: "PENDING",
      },
    ],
    totalExpected: 5000000,
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("PaydayPage", () => {
  it("shows setup when payday day is 0", async () => {
    mockGetPaydayDay.mockResolvedValue(0);

    render(PaydayPage);

    const heading = await screen.findByText("Set Your Payday");
    expect(heading).toBeTruthy();
  });

  it("shows dashboard when payday is configured", async () => {
    mockGetPaydayDay.mockResolvedValue(25);
    mockGetCurrentMonthStatus.mockResolvedValue(makeMonthlyStatus());

    render(PaydayPage);

    const portfolioName = await screen.findByText("Growth");
    expect(portfolioName).toBeTruthy();

    // Verify the expected amount appears somewhere on the page
    const container = document.body;
    expect(container.textContent).toContain("5,000,000");
  });

  it("shows PENDING status badge", async () => {
    mockGetPaydayDay.mockResolvedValue(25);
    mockGetCurrentMonthStatus.mockResolvedValue(makeMonthlyStatus());

    render(PaydayPage);

    const badge = await screen.findByText("PENDING");
    expect(badge).toBeTruthy();
  });

  it("shows action buttons for PENDING portfolio", async () => {
    mockGetPaydayDay.mockResolvedValue(25);
    mockGetCurrentMonthStatus.mockResolvedValue(makeMonthlyStatus());

    render(PaydayPage);

    const confirmBtn = await screen.findByText("Confirm");
    expect(confirmBtn).toBeTruthy();

    const deferBtn = screen.getByText("Defer");
    expect(deferBtn).toBeTruthy();

    const skipBtn = screen.getByText("Skip");
    expect(skipBtn).toBeTruthy();
  });

  it("does not show action buttons for CONFIRMED portfolio", async () => {
    mockGetPaydayDay.mockResolvedValue(25);
    mockGetCurrentMonthStatus.mockResolvedValue(
      makeMonthlyStatus({
        portfolios: [
          {
            portfolioId: "p1",
            portfolioName: "Growth",
            mode: "VALUE",
            expected: 5000000,
            actual: 5000000,
            status: "CONFIRMED",
          },
        ],
      }),
    );

    render(PaydayPage);

    await screen.findByText("CONFIRMED");
    expect(screen.queryByText("Confirm")).toBeNull();
    expect(screen.queryByText("Defer")).toBeNull();
    expect(screen.queryByText("Skip")).toBeNull();
  });

  it("shows error state", async () => {
    mockGetPaydayDay.mockRejectedValue(new Error("Network error"));

    render(PaydayPage);

    const errorMsg = await screen.findByText("Network error");
    expect(errorMsg).toBeTruthy();
  });

  it("calls SkipPayday when Skip is clicked", async () => {
    mockGetPaydayDay.mockResolvedValue(25);
    mockGetCurrentMonthStatus.mockResolvedValue(makeMonthlyStatus());
    mockSkipPayday.mockResolvedValue(undefined);

    render(PaydayPage);

    const skipBtn = await screen.findByText("Skip");
    const user = userEvent.setup();
    await user.click(skipBtn);

    expect(mockSkipPayday).toHaveBeenCalledWith("p1");
  });

  it("shows War Chest label for VALUE mode", async () => {
    mockGetPaydayDay.mockResolvedValue(25);
    mockGetCurrentMonthStatus.mockResolvedValue(makeMonthlyStatus());

    render(PaydayPage);

    const label = await screen.findByText("War Chest");
    expect(label).toBeTruthy();
  });

  it("shows DCA Fund label for DIVIDEND mode", async () => {
    mockGetPaydayDay.mockResolvedValue(25);
    mockGetCurrentMonthStatus.mockResolvedValue(
      makeMonthlyStatus({
        portfolios: [
          {
            portfolioId: "p1",
            portfolioName: "Income",
            mode: "DIVIDEND",
            expected: 3000000,
            actual: 0,
            status: "PENDING",
          },
        ],
      }),
    );

    render(PaydayPage);

    const label = await screen.findByText("DCA Fund");
    expect(label).toBeTruthy();
  });
});
