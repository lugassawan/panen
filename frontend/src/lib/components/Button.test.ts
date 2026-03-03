import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import ButtonWrapper from "./__tests__/ButtonWrapper.svelte";

describe("Button", () => {
  it("renders children text", () => {
    render(ButtonWrapper, { props: { text: "Click me" } });
    expect(screen.getByRole("button")).toHaveTextContent("Click me");
  });

  it("applies primary variant by default", () => {
    render(ButtonWrapper, { props: { text: "Test" } });
    const btn = screen.getByRole("button");
    expect(btn.className).toContain("bg-green-700");
  });

  it("applies danger variant", () => {
    render(ButtonWrapper, {
      props: { variant: "danger", text: "Delete" },
    });
    const btn = screen.getByRole("button");
    expect(btn.className).toContain("bg-negative");
  });

  it("applies size classes", () => {
    render(ButtonWrapper, {
      props: { size: "lg", text: "Large" },
    });
    const btn = screen.getByRole("button");
    expect(btn.className).toContain("px-6");
    expect(btn.className).toContain("text-base");
  });

  it("is disabled when disabled prop is true", () => {
    render(ButtonWrapper, {
      props: { disabled: true, text: "Test" },
    });
    expect(screen.getByRole("button")).toBeDisabled();
  });

  it("is disabled when loading", () => {
    render(ButtonWrapper, {
      props: { loading: true, text: "Test" },
    });
    expect(screen.getByRole("button")).toBeDisabled();
  });

  it("shows spinner when loading", () => {
    render(ButtonWrapper, {
      props: { loading: true, text: "Test" },
    });
    const btn = screen.getByRole("button");
    const svg = btn.querySelector("svg");
    expect(svg).not.toBeNull();
    expect(svg?.classList.contains("animate-spin")).toBe(true);
  });

  it("calls onclick handler", async () => {
    const handler = vi.fn();
    const user = userEvent.setup();
    render(ButtonWrapper, {
      props: { onclick: handler, text: "Click" },
    });
    await user.click(screen.getByRole("button"));
    expect(handler).toHaveBeenCalledOnce();
  });

  it("does not call onclick when disabled", async () => {
    const handler = vi.fn();
    const user = userEvent.setup();
    render(ButtonWrapper, {
      props: { onclick: handler, disabled: true, text: "Click" },
    });
    await user.click(screen.getByRole("button"));
    expect(handler).not.toHaveBeenCalled();
  });
});
