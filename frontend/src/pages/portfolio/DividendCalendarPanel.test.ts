import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetDividendCalendar: vi.fn(() =>
    Promise.resolve([
      { ticker: "BBCA", exDate: "2026-09-15", amount: 50, isProjection: true, totalIncome: 50000 },
    ]),
  ),
}));

import DividendCalendarPanel from "./DividendCalendarPanel.svelte";

describe("DividendCalendarPanel", () => {
  it("renders the component", () => {
    render(DividendCalendarPanel, { props: { portfolioId: "p1" } });
    expect(screen.getByTestId("dividend-calendar-panel")).toBeInTheDocument();
  });

  it("shows loading state initially", () => {
    render(DividendCalendarPanel, { props: { portfolioId: "p1" } });
    expect(screen.getByText("Loading calendar...")).toBeInTheDocument();
  });
});
