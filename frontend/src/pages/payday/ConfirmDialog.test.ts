import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import ConfirmDialog from "./ConfirmDialog.svelte";

describe("ConfirmDialog", () => {
  const defaultProps = {
    expected: 1000000,
    portfolioName: "Test Portfolio",
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
  };

  it("renders heading and portfolio name", () => {
    render(ConfirmDialog, { props: defaultProps });
    expect(screen.getByText("Confirm Payday")).toBeInTheDocument();
    expect(screen.getByText(/Test Portfolio/)).toBeInTheDocument();
  });

  it("pre-fills amount with expected value", () => {
    render(ConfirmDialog, { props: defaultProps });
    const input = screen.getByLabelText("Payday amount") as HTMLInputElement;
    expect(input.value).toBe("1000000");
  });

  it("calls onConfirm with amount on confirm click", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(ConfirmDialog, {
      props: { ...defaultProps, onConfirm },
    });

    await user.click(screen.getByRole("button", { name: "Confirm" }));
    expect(onConfirm).toHaveBeenCalledWith(1000000);
  });

  it("calls onCancel on cancel click", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(ConfirmDialog, {
      props: { ...defaultProps, onCancel },
    });

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalledOnce();
  });
});
