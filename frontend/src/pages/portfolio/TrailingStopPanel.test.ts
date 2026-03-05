import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import type { TrailingStopResponse } from "../../lib/types";
import TrailingStopPanel from "./TrailingStopPanel.svelte";

function makeTrailingStop(overrides: Partial<TrailingStopResponse> = {}): TrailingStopResponse {
  return {
    peakPrice: 10000,
    stopPercentage: 13.5,
    stopPrice: 8650,
    triggered: false,
    fundamentalExits: [
      { key: "roe_low", label: "ROE below 10%", detail: "ROE is 18.0%", triggered: false },
      { key: "der_high", label: "DER above 1.5x", detail: "DER is 0.50x", triggered: false },
      {
        key: "eps_negative",
        label: "EPS zero or negative",
        detail: "EPS is 500.00",
        triggered: false,
      },
    ],
    ...overrides,
  };
}

describe("TrailingStopPanel", () => {
  it("renders peak price, stop level, and stop percentage", () => {
    render(TrailingStopPanel, {
      props: { trailingStop: makeTrailingStop() },
    });

    expect(screen.getByText("Trailing Stop")).toBeInTheDocument();
    expect(screen.getByTestId("trailing-stop-panel")).toBeInTheDocument();
    expect(screen.getByText("-13.5%")).toBeInTheDocument();
  });

  it("shows triggered state", () => {
    render(TrailingStopPanel, {
      props: { trailingStop: makeTrailingStop({ triggered: true }) },
    });

    expect(screen.getByText("Trailing Stop Triggered")).toBeInTheDocument();
  });

  it("shows fundamental warnings when triggered", () => {
    const ts = makeTrailingStop({
      fundamentalExits: [
        { key: "roe_low", label: "ROE below 10%", detail: "ROE is 5.0%", triggered: true },
        { key: "der_high", label: "DER above 1.5x", detail: "DER is 0.50x", triggered: false },
        {
          key: "eps_negative",
          label: "EPS zero or negative",
          detail: "EPS is 500.00",
          triggered: false,
        },
      ],
    });

    render(TrailingStopPanel, { props: { trailingStop: ts } });

    expect(screen.getByText("Fundamental Warnings")).toBeInTheDocument();
    expect(screen.getByText("ROE below 10%")).toBeInTheDocument();
  });

  it("hides fundamental warnings section when none are triggered", () => {
    render(TrailingStopPanel, {
      props: { trailingStop: makeTrailingStop() },
    });

    expect(screen.queryByText("Fundamental Warnings")).not.toBeInTheDocument();
  });

  it("shows disclaimer note", () => {
    render(TrailingStopPanel, {
      props: { trailingStop: makeTrailingStop() },
    });

    expect(screen.getByText(/Set trailing stops manually/)).toBeInTheDocument();
  });
});
