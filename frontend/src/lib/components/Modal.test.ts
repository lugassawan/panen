import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import ModalTest from "./Modal.test.svelte";

describe("Modal", () => {
  it("renders dialog with role and aria-modal", () => {
    render(ModalTest, { props: { onClose: vi.fn() } });

    const dialog = screen.getByRole("dialog");
    expect(dialog).toBeInTheDocument();
    expect(dialog).toHaveAttribute("aria-modal", "true");
  });

  it("renders title in heading when provided", () => {
    render(ModalTest, { props: { title: "Test Title", onClose: vi.fn() } });

    expect(screen.getByText("Test Title")).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toHaveAttribute("aria-labelledby", "modal-title");
  });

  it("uses aria-label when no title provided", () => {
    render(ModalTest, { props: { "aria-label": "Custom label", onClose: vi.fn() } });

    const dialog = screen.getByRole("dialog");
    expect(dialog).toHaveAttribute("aria-label", "Custom label");
    expect(dialog).not.toHaveAttribute("aria-labelledby");
  });

  it("calls onClose on Escape key", async () => {
    const onClose = vi.fn();
    render(ModalTest, { props: { onClose } });
    const user = userEvent.setup();

    await user.keyboard("{Escape}");

    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose on backdrop click", async () => {
    const onClose = vi.fn();
    render(ModalTest, { props: { onClose } });
    const user = userEvent.setup();

    const backdrop = screen.getByRole("presentation");
    await user.click(backdrop);

    expect(onClose).toHaveBeenCalled();
  });

  it("renders children content", () => {
    render(ModalTest, { props: { onClose: vi.fn() } });

    expect(screen.getByText("Dialog body content")).toBeInTheDocument();
  });

  it("renders footer snippet when provided", () => {
    render(ModalTest, { props: { onClose: vi.fn(), hasFooter: true } });

    expect(screen.getByText("Footer action")).toBeInTheDocument();
  });

  it("does not render when open is false", () => {
    render(ModalTest, { props: { open: false, onClose: vi.fn() } });

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("wraps focus from last to first on Tab", async () => {
    render(ModalTest, { props: { onClose: vi.fn() } });
    const user = userEvent.setup();

    screen.getByText("Last").focus();
    await user.tab();

    expect(screen.getByText("First")).toHaveFocus();
  });

  it("wraps focus from first to last on Shift+Tab", async () => {
    render(ModalTest, { props: { onClose: vi.fn() } });
    const user = userEvent.setup();

    screen.getByText("First").focus();
    await user.tab({ shift: true });

    expect(screen.getByText("Last")).toHaveFocus();
  });
});
