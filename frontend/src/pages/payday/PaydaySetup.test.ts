import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import PaydaySetup from "./PaydaySetup.svelte";

describe("PaydaySetup", () => {
  it("renders heading and save button", () => {
    render(PaydaySetup, { props: { onSave: vi.fn() } });
    expect(screen.getByText("Set Your Payday")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Save Payday/i })).toBeInTheDocument();
  });

  it("renders day select with 31 options", () => {
    render(PaydaySetup, { props: { onSave: vi.fn() } });
    const select = screen.getByLabelText("Select payday day");
    const options = select.querySelectorAll("option");
    expect(options.length).toBe(31);
  });

  it("calls onSave with default day 25", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(PaydaySetup, { props: { onSave } });

    await user.click(screen.getByRole("button", { name: /Save Payday/i }));
    expect(onSave).toHaveBeenCalledWith(25);
  });
});
