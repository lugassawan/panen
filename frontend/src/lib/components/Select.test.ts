import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import SelectWrapper from "./__tests__/SelectWrapper.svelte";

describe("Select", () => {
  it("renders a combobox with options", () => {
    render(SelectWrapper, { props: { "aria-label": "test" } });
    const select = screen.getByRole("combobox");
    expect(select).toBeInTheDocument();
    expect(select.querySelectorAll("option")).toHaveLength(3);
  });

  it("applies base styling classes", () => {
    render(SelectWrapper, { props: { "aria-label": "test" } });
    const select = screen.getByRole("combobox");
    expect(select.className).toContain("appearance-none");
    expect(select.className).toContain("rounded");
    expect(select.className).toContain("border-border-default");
    expect(select.className).toContain("bg-bg-elevated");
  });

  it("renders chevron icon", () => {
    render(SelectWrapper, { props: { "aria-label": "test" } });
    const select = screen.getByRole("combobox");
    const wrapper = select.parentElement;
    expect(wrapper).not.toBeNull();
    const chevron = wrapper?.querySelector("svg");
    expect(chevron).toBeInTheDocument();
  });

  it("sets id attribute", () => {
    render(SelectWrapper, {
      props: { id: "my-select", "aria-label": "test" },
    });
    expect(document.getElementById("my-select")).toBeInTheDocument();
  });

  it("is disabled when disabled prop is true", () => {
    render(SelectWrapper, {
      props: { disabled: true, "aria-label": "test" },
    });
    expect(screen.getByRole("combobox")).toBeDisabled();
  });

  it("calls onchange handler", async () => {
    const handler = vi.fn();
    const user = userEvent.setup();
    render(SelectWrapper, {
      props: { onchange: handler, "aria-label": "test" },
    });
    await user.selectOptions(screen.getByRole("combobox"), "b");
    expect(handler).toHaveBeenCalled();
  });

  it("appends extra class", () => {
    render(SelectWrapper, {
      props: { class: "!w-auto", "aria-label": "test" },
    });
    const select = screen.getByRole("combobox");
    expect(select.className).toContain("!w-auto");
  });

  it("sets aria-label", () => {
    render(SelectWrapper, { props: { "aria-label": "Risk profile" } });
    expect(screen.getByLabelText("Risk profile")).toBeInTheDocument();
  });
});
