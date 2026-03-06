import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { CrashCapitalResponse, DeploymentPlanResponse } from "../../lib/types";
import CrashCapitalPanel from "./CrashCapitalPanel.svelte";

function makeCapital(): CrashCapitalResponse {
  return { portfolioId: "p1", amount: 10000000, deployed: 0 };
}

function makePlan(): DeploymentPlanResponse {
  return {
    total: 10000000,
    deployed: 0,
    remaining: 10000000,
    levels: [
      { level: "NORMAL_DIP", pct: 30, amount: 3000000 },
      { level: "CRASH", pct: 40, amount: 4000000 },
      { level: "EXTREME", pct: 30, amount: 3000000 },
    ],
  };
}

describe("CrashCapitalPanel", () => {
  it("renders crash capital heading", () => {
    render(CrashCapitalPanel, {
      props: { capital: makeCapital(), plan: makePlan(), onSave: vi.fn(), onOpenSettings: vi.fn() },
    });
    expect(screen.getByText("Crash Capital")).toBeTruthy();
  });

  it("shows deployment breakdown", () => {
    render(CrashCapitalPanel, {
      props: { capital: makeCapital(), plan: makePlan(), onSave: vi.fn(), onOpenSettings: vi.fn() },
    });
    expect(document.body.textContent).toContain("3,000,000");
    expect(document.body.textContent).toContain("4,000,000");
  });

  it("shows Deployment Settings link", () => {
    render(CrashCapitalPanel, {
      props: { capital: makeCapital(), plan: makePlan(), onSave: vi.fn(), onOpenSettings: vi.fn() },
    });
    expect(screen.getByText("Deployment Settings")).toBeTruthy();
  });

  it("calls onOpenSettings when link is clicked", async () => {
    const onOpenSettings = vi.fn();
    render(CrashCapitalPanel, {
      props: { capital: makeCapital(), plan: makePlan(), onSave: vi.fn(), onOpenSettings },
    });
    const user = userEvent.setup();
    await user.click(screen.getByText("Deployment Settings"));
    expect(onOpenSettings).toHaveBeenCalled();
  });
});
