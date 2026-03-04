import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ChecklistEvaluationResponse } from "../lib/types";
import ChecklistPanel from "./ChecklistPanel.svelte";

const mockEvaluateChecklist = vi.fn();
const mockToggleManualCheck = vi.fn();
const mockResetChecklist = vi.fn();

vi.mock("../../wailsjs/go/backend/App", () => ({
  EvaluateChecklist: (...args: unknown[]) => mockEvaluateChecklist(...args),
  ToggleManualCheck: (...args: unknown[]) => mockToggleManualCheck(...args),
  ResetChecklist: (...args: unknown[]) => mockResetChecklist(...args),
}));

function makeEvaluation(
  overrides: Partial<ChecklistEvaluationResponse> = {},
): ChecklistEvaluationResponse {
  return {
    action: "BUY",
    ticker: "BBCA",
    checks: [
      {
        key: "margin_of_safety",
        label: "Margin of Safety > 20%",
        type: "AUTO",
        status: "PASS",
        detail: "MoS: 35.2%",
      },
      {
        key: "der_check",
        label: "DER < 1.0",
        type: "AUTO",
        status: "FAIL",
        detail: "DER: 1.25",
      },
      {
        key: "manual_review",
        label: "Reviewed annual report",
        type: "MANUAL",
        status: "PENDING",
        detail: "",
      },
    ],
    allPassed: false,
    ...overrides,
  };
}

describe("ChecklistPanel", () => {
  const defaultProps = {
    portfolioId: "p1",
    ticker: "BBCA",
    action: "BUY" as const,
  };

  beforeEach(() => {
    mockEvaluateChecklist.mockReset();
    mockToggleManualCheck.mockReset();
    mockResetChecklist.mockReset();
  });

  it("shows loading state", () => {
    mockEvaluateChecklist.mockReturnValue(new Promise(() => {}));
    render(ChecklistPanel, { props: defaultProps });

    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.getByText(/evaluating checklist/i)).toBeInTheDocument();
  });

  it("shows auto-check items with pass/fail icons", async () => {
    mockEvaluateChecklist.mockResolvedValue(makeEvaluation());
    render(ChecklistPanel, { props: defaultProps });

    await screen.findByText("Margin of Safety > 20%");
    expect(screen.getByText("MoS: 35.2%")).toBeInTheDocument();
    expect(screen.getByText("DER: 1.25")).toBeInTheDocument();
    expect(screen.getAllByTestId("auto-check")).toHaveLength(2);
  });

  it("shows manual-check items with checkboxes", async () => {
    mockEvaluateChecklist.mockResolvedValue(makeEvaluation());
    render(ChecklistPanel, { props: defaultProps });

    await screen.findByText("Reviewed annual report");
    expect(screen.getAllByTestId("manual-check")).toHaveLength(1);
    expect(screen.getByRole("checkbox")).toBeInTheDocument();
  });

  it("calls ToggleManualCheck when checkbox toggled", async () => {
    mockEvaluateChecklist.mockResolvedValue(makeEvaluation());
    mockToggleManualCheck.mockResolvedValue(undefined);
    render(ChecklistPanel, { props: defaultProps });
    const user = userEvent.setup();

    await screen.findByText("Reviewed annual report");
    const checkbox = screen.getByRole("checkbox");
    await user.click(checkbox);

    expect(mockToggleManualCheck).toHaveBeenCalledWith("p1", "BBCA", "BUY", "manual_review", true);
  });

  it("shows check count when not all passed", async () => {
    mockEvaluateChecklist.mockResolvedValue(makeEvaluation());
    render(ChecklistPanel, { props: defaultProps });

    expect(await screen.findByText(/1 \/ 3 checks passed/)).toBeInTheDocument();
  });

  it("shows all checks passed message when allPassed is true", async () => {
    mockEvaluateChecklist.mockResolvedValue(
      makeEvaluation({
        allPassed: true,
        checks: [
          {
            key: "margin_of_safety",
            label: "Margin of Safety > 20%",
            type: "AUTO",
            status: "PASS",
            detail: "MoS: 35.2%",
          },
        ],
      }),
    );
    render(ChecklistPanel, { props: defaultProps });

    expect(await screen.findByText(/all checks passed/i)).toBeInTheDocument();
  });

  it("shows suggestion card when allPassed with suggestion", async () => {
    mockEvaluateChecklist.mockResolvedValue(
      makeEvaluation({
        allPassed: true,
        checks: [
          {
            key: "margin_of_safety",
            label: "Margin of Safety > 20%",
            type: "AUTO",
            status: "PASS",
            detail: "MoS: 35.2%",
          },
        ],
        suggestion: {
          action: "BUY",
          ticker: "BBCA",
          lots: 5,
          pricePerShare: 9250,
          grossCost: 4625000,
          fee: 6938,
          tax: 0,
          netCost: 4631938,
          newAvgBuyPrice: 9100,
          newPositionLots: 15,
          newPositionPct: 25.5,
          capitalGainPct: 0,
        },
      }),
    );
    render(ChecklistPanel, { props: defaultProps });

    expect(await screen.findByTestId("suggestion-card")).toBeInTheDocument();
    expect(screen.getByText(/Trade Suggestion: Buy/)).toBeInTheDocument();
  });

  it("does not show suggestion card when not all passed", async () => {
    mockEvaluateChecklist.mockResolvedValue(makeEvaluation());
    render(ChecklistPanel, { props: defaultProps });

    await screen.findByText(/1 \/ 3 checks passed/);
    expect(screen.queryByTestId("suggestion-card")).not.toBeInTheDocument();
  });

  it("calls ResetChecklist on reset button click", async () => {
    mockEvaluateChecklist.mockResolvedValue(makeEvaluation());
    mockResetChecklist.mockResolvedValue(undefined);
    render(ChecklistPanel, { props: defaultProps });
    const user = userEvent.setup();

    await screen.findByText(/checks passed/);
    await user.click(screen.getByRole("button", { name: /reset/i }));

    expect(mockResetChecklist).toHaveBeenCalledWith("p1", "BBCA", "BUY");
  });

  it("shows error on evaluation failure", async () => {
    mockEvaluateChecklist.mockRejectedValue(new Error("eval failed"));
    render(ChecklistPanel, { props: defaultProps });

    expect(await screen.findByRole("alert")).toHaveTextContent(/eval failed/);
  });
});
