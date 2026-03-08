import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import ConfirmDialogTest from "./ConfirmDialog.test.svelte";

describe("ConfirmDialog", () => {
  it("renders title and body", () => {
    render(ConfirmDialogTest, {
      props: { title: "Delete Item", body: "Are you sure?", onConfirm: vi.fn(), onCancel: vi.fn() },
    });

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Delete Item")).toBeInTheDocument();
    expect(screen.getByText("Are you sure?")).toBeInTheDocument();
  });

  it("calls onConfirm when confirm button clicked", async () => {
    const onConfirm = vi.fn();
    render(ConfirmDialogTest, {
      props: { title: "Delete", body: "Sure?", onConfirm, onCancel: vi.fn() },
    });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /confirm/i }));

    expect(onConfirm).toHaveBeenCalled();
  });

  it("calls onCancel when cancel button clicked", async () => {
    const onCancel = vi.fn();
    render(ConfirmDialogTest, {
      props: { title: "Delete", body: "Sure?", onConfirm: vi.fn(), onCancel },
    });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(onCancel).toHaveBeenCalled();
  });

  it("renders custom confirm label", () => {
    render(ConfirmDialogTest, {
      props: {
        title: "Remove",
        body: "Body",
        confirmLabel: "Remove Now",
        onConfirm: vi.fn(),
        onCancel: vi.fn(),
      },
    });

    expect(screen.getByRole("button", { name: /remove now/i })).toBeInTheDocument();
  });
});
