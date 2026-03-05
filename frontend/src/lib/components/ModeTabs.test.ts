import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { mode } from "../stores/mode.svelte";
import ModeTabs from "./ModeTabs.svelte";

describe("ModeTabs", () => {
  beforeEach(() => {
    mode.set("value");
  });

  it("renders both tabs", () => {
    render(ModeTabs);
    expect(screen.getByRole("tab", { name: /Value/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /Dividend/i })).toBeInTheDocument();
  });

  it("marks value tab as selected by default", () => {
    render(ModeTabs);
    expect(screen.getByRole("tab", { name: /Value/i })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tab", { name: /Dividend/i })).toHaveAttribute(
      "aria-selected",
      "false",
    );
  });

  it("switches mode on tab click", async () => {
    const user = userEvent.setup();
    render(ModeTabs);

    await user.click(screen.getByRole("tab", { name: /Dividend/i }));
    expect(mode.current).toBe("dividend");

    await user.click(screen.getByRole("tab", { name: /Value/i }));
    expect(mode.current).toBe("value");
  });
});
