import { render, screen, within } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { ScreenerItemResponse } from "../../lib/types";
import ScreenerPage from "./ScreenerPage.svelte";

const mockRunScreen = vi.fn<() => Promise<ScreenerItemResponse[]>>();
const mockListIndices = vi.fn<() => Promise<string[]>>();
const mockListSectors = vi.fn<() => Promise<string[]>>();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  RunScreen: (...args: unknown[]) => mockRunScreen(...args),
  ListScreenerIndices: () => mockListIndices(),
  ListScreenerSectors: () => mockListSectors(),
}));

function setup() {
  mockListIndices.mockResolvedValue(["IDX30", "LQ45"]);
  mockListSectors.mockResolvedValue(["Banking", "Telecom"]);
  mockRunScreen.mockResolvedValue([]);
}

describe("ScreenerPage", () => {
  it("renders page header and initial state", () => {
    setup();
    render(ScreenerPage);
    expect(screen.getByText("Stock Screener")).toBeInTheDocument();
    expect(screen.getByText(/configure and run a screen/i)).toBeInTheDocument();
  });

  it("renders filter controls", () => {
    setup();
    render(ScreenerPage);
    expect(screen.getByLabelText("Universe")).toBeInTheDocument();
    expect(screen.getByText("Run Screen")).toBeInTheDocument();
  });

  it("renders risk profile buttons", () => {
    setup();
    render(ScreenerPage);
    const radioGroup = screen.getByRole("radiogroup", { name: /risk profile/i });
    expect(within(radioGroup).getByText("Conservative")).toBeInTheDocument();
    expect(within(radioGroup).getByText("Moderate")).toBeInTheDocument();
    expect(within(radioGroup).getByText("Aggressive")).toBeInTheDocument();
  });

  it("displays results table after screen", async () => {
    setup();
    mockRunScreen.mockResolvedValue([
      {
        ticker: "BBCA",
        sector: "Banking",
        price: 9000,
        roe: 20,
        der: 0.5,
        eps: 500,
        pbv: 3.0,
        per: 18,
        dividendYield: 2.5,
        grahamNumber: 5000,
        entryPrice: 4000,
        exitTarget: 10000,
        verdict: "UNDERVALUED",
        checks: [
          {
            key: "roe_above_min",
            label: "ROE above minimum",
            status: "PASS",
            value: 20,
            limit: 15,
          },
          {
            key: "der_below_max",
            label: "DER below maximum",
            status: "PASS",
            value: 0.5,
            limit: 0.8,
          },
        ],
        passed: true,
        score: 3.5,
        fetchedAt: "2026-01-01T00:00:00Z",
      },
      {
        ticker: "TLKM",
        sector: "Telecom",
        price: 3000,
        roe: 10,
        der: 0.6,
        checks: [],
        passed: false,
        score: 1.0,
      },
    ]);

    const user = userEvent.setup();
    render(ScreenerPage);

    // Wait for reference data to load (options appear asynchronously)
    const universeNameSelect = await screen.findByLabelText("Index");
    await screen.findByText("IDX30");
    await user.selectOptions(universeNameSelect, "IDX30");

    await user.click(screen.getByText("Run Screen"));

    expect(await screen.findByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("TLKM")).toBeInTheDocument();
    expect(screen.getByText("2 stocks screened")).toBeInTheDocument();
    expect(screen.getByText("1 pass")).toBeInTheDocument();
    expect(screen.getByText("1 fail")).toBeInTheDocument();
  });

  it("shows error state", async () => {
    setup();
    mockRunScreen.mockRejectedValue(new Error("test error"));

    const user = userEvent.setup();
    render(ScreenerPage);

    const universeNameSelect = await screen.findByLabelText("Index");
    await screen.findByText("IDX30");
    await user.selectOptions(universeNameSelect, "IDX30");

    await user.click(screen.getByText("Run Screen"));

    expect(await screen.findByText("test error")).toBeInTheDocument();
  });

  it("disables Run Screen when no universe selected", () => {
    setup();
    render(ScreenerPage);
    expect(screen.getByText("Run Screen")).toBeDisabled();
  });
});
