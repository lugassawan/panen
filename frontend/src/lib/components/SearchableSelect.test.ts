import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import SearchableSelectHarness from "./SearchableSelectHarness.test.svelte";

const items = [
  { id: "a", label: "Apple" },
  { id: "b", label: "Banana" },
  { id: "c", label: "Cherry" },
];

describe("SearchableSelect", () => {
  it("renders with placeholder", () => {
    render(SearchableSelectHarness, { props: { items, placeholder: "Pick fruit..." } });
    expect(screen.getByRole("combobox")).toHaveAttribute("placeholder", "Pick fruit...");
  });

  it("opens on click and shows all items", async () => {
    render(SearchableSelectHarness, { props: { items } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));

    expect(screen.getByRole("listbox")).toBeInTheDocument();
    expect(screen.getAllByRole("option")).toHaveLength(3);
  });

  it("filters items on typing", async () => {
    render(SearchableSelectHarness, { props: { items } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByRole("combobox"), "ban");

    expect(screen.getAllByRole("option")).toHaveLength(1);
    expect(screen.getByText("Banana")).toBeInTheDocument();
  });

  it("selects item on click", async () => {
    const onselect = vi.fn();
    render(SearchableSelectHarness, { props: { items, onselect } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByText("Banana"));

    expect(onselect).toHaveBeenCalledWith("b");
  });

  it("navigates with keyboard and selects with Enter", async () => {
    const onselect = vi.fn();
    render(SearchableSelectHarness, { props: { items, onselect } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    await user.keyboard("{ArrowDown}{Enter}");

    expect(onselect).toHaveBeenCalledWith("b");
  });

  it("closes on Escape", async () => {
    render(SearchableSelectHarness, { props: { items } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    expect(screen.getByRole("listbox")).toBeInTheDocument();

    await user.keyboard("{Escape}");
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
  });

  it("shows empty state when no matches", async () => {
    render(SearchableSelectHarness, { props: { items } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByRole("combobox"), "xyz");

    expect(screen.queryAllByRole("option")).toHaveLength(0);
  });

  it("selects footer with keyboard Enter", async () => {
    const onfooterselect = vi.fn();
    render(SearchableSelectHarness, {
      props: { items, showFooter: true, onfooterselect },
    });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    // Navigate past all 3 items to reach footer
    await user.keyboard("{ArrowDown}{ArrowDown}{ArrowDown}{Enter}");

    expect(onfooterselect).toHaveBeenCalledOnce();
    // Dropdown should be closed
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
  });

  it("footer is always visible", async () => {
    render(SearchableSelectHarness, { props: { items, showFooter: true } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));

    expect(screen.getByText("Footer action")).toBeInTheDocument();
  });
});
