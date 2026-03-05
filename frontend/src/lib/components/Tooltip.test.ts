import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import TooltipWrapper from "./__tests__/TooltipWrapper.svelte";

describe("Tooltip", () => {
  it("renders trigger content", () => {
    render(TooltipWrapper);
    expect(screen.getByRole("button", { name: "Hover me" })).toBeInTheDocument();
  });

  it("hides tooltip by default", () => {
    render(TooltipWrapper, { props: { text: "Help text" } });
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();
  });

  it("shows tooltip on hover", async () => {
    const user = userEvent.setup();
    render(TooltipWrapper, { props: { text: "Help text" } });

    await user.hover(screen.getByRole("button"));
    expect(screen.getByRole("tooltip")).toHaveTextContent("Help text");
  });

  it("hides tooltip on unhover", async () => {
    const user = userEvent.setup();
    render(TooltipWrapper, { props: { text: "Help text" } });

    await user.hover(screen.getByRole("button"));
    expect(screen.getByRole("tooltip")).toBeInTheDocument();

    await user.unhover(screen.getByRole("button"));
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();
  });

  it("shows tooltip on focus", async () => {
    const user = userEvent.setup();
    render(TooltipWrapper, { props: { text: "Help text" } });

    await user.tab();
    expect(screen.getByRole("tooltip")).toHaveTextContent("Help text");
  });

  it("hides tooltip on blur", async () => {
    const user = userEvent.setup();
    render(TooltipWrapper, { props: { text: "Help text" } });

    await user.tab();
    expect(screen.getByRole("tooltip")).toBeInTheDocument();

    await user.tab();
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();
  });

  it("links tooltip to trigger via aria-describedby", async () => {
    const user = userEvent.setup();
    render(TooltipWrapper, { props: { text: "Help text" } });

    await user.hover(screen.getByRole("button"));
    const tooltip = screen.getByRole("tooltip");
    const trigger = screen.getByRole("button").parentElement;
    expect(trigger?.getAttribute("aria-describedby")).toBe(tooltip.id);
  });
});
