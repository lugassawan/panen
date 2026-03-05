import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { DiagnosticResponse } from "../../lib/types";
import FallingKnifeDialog from "./FallingKnifeDialog.svelte";

function makeDiagnostic(overrides: Partial<DiagnosticResponse> = {}): DiagnosticResponse {
  return {
    marketCrashed: true,
    companyBadNews: null,
    fundamentalsOK: null,
    belowEntry: true,
    signal: "INCONCLUSIVE",
    ...overrides,
  };
}

describe("FallingKnifeDialog", () => {
  it("renders ticker and signal badge", () => {
    render(FallingKnifeDialog, {
      props: {
        ticker: "BBCA",
        diagnostic: makeDiagnostic(),
        onUpdate: vi.fn(),
        onClose: vi.fn(),
      },
    });
    expect(screen.getByText("BBCA")).toBeTruthy();
    expect(screen.getByText("Inconclusive")).toBeTruthy();
  });

  it("shows auto-detected market crash status", () => {
    render(FallingKnifeDialog, {
      props: {
        ticker: "BBCA",
        diagnostic: makeDiagnostic({ marketCrashed: true }),
        onUpdate: vi.fn(),
        onClose: vi.fn(),
      },
    });
    const yesElements = screen.getAllByText("Yes");
    expect(yesElements.length).toBeGreaterThanOrEqual(1);
  });

  it("shows Opportunity badge when signal is OPPORTUNITY", () => {
    render(FallingKnifeDialog, {
      props: {
        ticker: "BBCA",
        diagnostic: makeDiagnostic({ signal: "OPPORTUNITY" }),
        onUpdate: vi.fn(),
        onClose: vi.fn(),
      },
    });
    expect(screen.getByText("Opportunity")).toBeTruthy();
  });

  it("shows Falling Knife badge when signal is FALLING_KNIFE", () => {
    render(FallingKnifeDialog, {
      props: {
        ticker: "BBCA",
        diagnostic: makeDiagnostic({ signal: "FALLING_KNIFE" }),
        onUpdate: vi.fn(),
        onClose: vi.fn(),
      },
    });
    expect(screen.getByText("Falling Knife")).toBeTruthy();
  });

  it("calls onClose when Close button is clicked", async () => {
    const onClose = vi.fn();
    render(FallingKnifeDialog, {
      props: {
        ticker: "BBCA",
        diagnostic: makeDiagnostic(),
        onUpdate: vi.fn(),
        onClose,
      },
    });
    const user = userEvent.setup();
    await user.click(screen.getByText("Close"));
    expect(onClose).toHaveBeenCalled();
  });
});
