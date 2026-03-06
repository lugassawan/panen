import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import type { SectorWeight } from "../../lib/types";
import SectorWarnings from "./SectorWarnings.svelte";

describe("SectorWarnings", () => {
  it("shows warning when sector exceeds 30%", () => {
    const weights: SectorWeight[] = [
      { sector: "Financials", value: 8_000_000, pct: 75 },
      { sector: "Telco", value: 2_000_000, pct: 25 },
    ];
    render(SectorWarnings, { props: { sectorWeights: weights } });
    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByText(/Financials/)).toBeInTheDocument();
    expect(screen.getByText(/75\.00%/)).toBeInTheDocument();
  });

  it("shows no warnings when all sectors under 30%", () => {
    const weights: SectorWeight[] = [
      { sector: "Financials", value: 3_000_000, pct: 25 },
      { sector: "Telco", value: 3_000_000, pct: 25 },
      { sector: "Consumer", value: 3_000_000, pct: 25 },
      { sector: "Energy", value: 3_000_000, pct: 25 },
    ];
    render(SectorWarnings, { props: { sectorWeights: weights } });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("shows multiple warnings for multiple concentrated sectors", () => {
    const weights: SectorWeight[] = [
      { sector: "Financials", value: 5_000_000, pct: 45 },
      { sector: "Telco", value: 4_000_000, pct: 35 },
      { sector: "Energy", value: 2_000_000, pct: 20 },
    ];
    render(SectorWarnings, { props: { sectorWeights: weights } });
    const alerts = screen.getAllByRole("alert");
    expect(alerts).toHaveLength(2);
    expect(screen.getByText(/Financials/)).toBeInTheDocument();
    expect(screen.getByText(/Telco/)).toBeInTheDocument();
  });

  it("shows no warnings for empty sector weights", () => {
    render(SectorWarnings, { props: { sectorWeights: [] } });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("does not warn at exactly 30%", () => {
    const weights: SectorWeight[] = [
      { sector: "Financials", value: 3_000_000, pct: 30 },
      { sector: "Telco", value: 7_000_000, pct: 70 },
    ];
    render(SectorWarnings, { props: { sectorWeights: weights } });
    const alerts = screen.getAllByRole("alert");
    expect(alerts).toHaveLength(1);
    expect(screen.getByText(/Telco/)).toBeInTheDocument();
    expect(screen.queryByText(/High concentration:.*Financials/)).not.toBeInTheDocument();
  });
});
