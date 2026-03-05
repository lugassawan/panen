import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { DeploymentSettingsResponse } from "../../lib/types";
import DeploymentSettings from "./DeploymentSettings.svelte";

function makeSettings(): DeploymentSettingsResponse {
  return { normal: 30, crash: 40, extreme: 30 };
}

describe("DeploymentSettings", () => {
  it("renders heading and description", () => {
    render(DeploymentSettings, {
      props: { settings: makeSettings(), onSave: vi.fn(), onClose: vi.fn() },
    });
    expect(screen.getByText("Deployment Settings")).toBeTruthy();
    expect(screen.getByText(/Must sum to 100%/)).toBeTruthy();
  });

  it("shows total as valid when sum is 100", () => {
    render(DeploymentSettings, {
      props: { settings: makeSettings(), onSave: vi.fn(), onClose: vi.fn() },
    });
    expect(document.body.textContent).toContain("Total: 100%");
  });

  it("disables save when sum is not 100", async () => {
    const settings = { normal: 20, crash: 40, extreme: 30 };
    render(DeploymentSettings, {
      props: { settings, onSave: vi.fn(), onClose: vi.fn() },
    });
    const saveButton = screen.getByText("Save");
    expect(saveButton).toHaveProperty("disabled", true);
  });

  it("calls onClose when Cancel is clicked", async () => {
    const onClose = vi.fn();
    render(DeploymentSettings, {
      props: { settings: makeSettings(), onSave: vi.fn(), onClose },
    });
    const user = userEvent.setup();
    await user.click(screen.getByText("Cancel"));
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onSave with values when Save is clicked", async () => {
    const onSave = vi.fn();
    render(DeploymentSettings, {
      props: { settings: makeSettings(), onSave, onClose: vi.fn() },
    });
    const user = userEvent.setup();
    await user.click(screen.getByText("Save"));
    expect(onSave).toHaveBeenCalledWith(30, 40, 30);
  });
});
