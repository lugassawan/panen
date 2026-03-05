import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { toastStore } from "../stores/toast.svelte";
import ToastContainerWrapper from "./__tests__/ToastContainerWrapper.svelte";

describe("ToastContainer", () => {
  beforeEach(() => {
    toastStore.clear();
  });

  it("renders nothing when no toasts", () => {
    render(ToastContainerWrapper);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it("renders toasts from store", () => {
    render(ToastContainerWrapper, {
      props: { toasts: [{ message: "Saved!", variant: "success" as const }] },
    });
    expect(screen.getByText("Saved!")).toBeInTheDocument();
  });

  it("renders multiple toasts", () => {
    render(ToastContainerWrapper, {
      props: {
        toasts: [
          { message: "First", variant: "info" as const },
          { message: "Second", variant: "error" as const },
        ],
      },
    });
    expect(screen.getByText("First")).toBeInTheDocument();
    expect(screen.getByText("Second")).toBeInTheDocument();
  });

  it("dismisses toast on click", async () => {
    const user = userEvent.setup();
    render(ToastContainerWrapper, {
      props: { toasts: [{ message: "Dismiss me", variant: "info" as const }] },
    });

    expect(screen.getByText("Dismiss me")).toBeInTheDocument();
    await user.click(screen.getByLabelText("Dismiss notification"));
    expect(screen.queryByText("Dismiss me")).not.toBeInTheDocument();
  });

  it("uses assertive aria-live for errors", () => {
    render(ToastContainerWrapper, {
      props: { toasts: [{ message: "Error!", variant: "error" as const }] },
    });
    const toast = screen.getByText("Error!").closest("[role]");
    expect(toast?.getAttribute("aria-live")).toBe("assertive");
  });

  it("uses polite aria-live for non-errors", () => {
    render(ToastContainerWrapper, {
      props: { toasts: [{ message: "Info", variant: "info" as const }] },
    });
    const toast = screen.getByText("Info").closest("[role]");
    expect(toast?.getAttribute("aria-live")).toBe("polite");
  });
});
