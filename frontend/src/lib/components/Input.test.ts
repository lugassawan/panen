import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import InputWrapper from "./__tests__/InputWrapper.svelte";

describe("Input", () => {
  it("renders with default type text", () => {
    render(InputWrapper, { props: { "aria-label": "test" } });
    const input = screen.getByRole("textbox");
    expect(input).toBeInTheDocument();
  });

  it("applies base styling classes", () => {
    render(InputWrapper, { props: { "aria-label": "test" } });
    const input = screen.getByRole("textbox");
    expect(input.className).toContain("rounded");
    expect(input.className).toContain("border-border-default");
    expect(input.className).toContain("bg-bg-elevated");
  });

  it("sets placeholder", () => {
    render(InputWrapper, {
      props: { placeholder: "Enter value", "aria-label": "test" },
    });
    expect(screen.getByPlaceholderText("Enter value")).toBeInTheDocument();
  });

  it("sets id attribute", () => {
    render(InputWrapper, { props: { id: "my-input", "aria-label": "test" } });
    expect(document.getElementById("my-input")).toBeInTheDocument();
  });

  it("is disabled when disabled prop is true", () => {
    render(InputWrapper, {
      props: { disabled: true, "aria-label": "test" },
    });
    expect(screen.getByRole("textbox")).toBeDisabled();
  });

  it("calls oninput handler", async () => {
    const handler = vi.fn();
    const user = userEvent.setup();
    render(InputWrapper, {
      props: { oninput: handler, "aria-label": "test" },
    });
    await user.type(screen.getByRole("textbox"), "a");
    expect(handler).toHaveBeenCalled();
  });

  it("appends extra class", () => {
    render(InputWrapper, {
      props: { class: "font-mono", "aria-label": "test" },
    });
    const input = screen.getByRole("textbox");
    expect(input.className).toContain("font-mono");
  });

  it("sets aria-label", () => {
    render(InputWrapper, { props: { "aria-label": "Stock ticker" } });
    expect(screen.getByLabelText("Stock ticker")).toBeInTheDocument();
  });
});
