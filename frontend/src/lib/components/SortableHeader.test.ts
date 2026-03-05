import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import SortableHeaderWrapper from "./__tests__/SortableHeaderWrapper.svelte";

describe("SortableHeader", () => {
  it("renders label text", () => {
    render(SortableHeaderWrapper, { props: { label: "Price" } });
    expect(screen.getByRole("button", { name: /Price/ })).toBeInTheDocument();
  });

  it("renders as button inside th", () => {
    render(SortableHeaderWrapper, { props: { label: "Price" } });
    const button = screen.getByRole("button", { name: /Price/ });
    expect(button.closest("th")).toBeInTheDocument();
  });

  it("calls onclick with field on click", async () => {
    const user = userEvent.setup();
    const onclick = vi.fn();
    render(SortableHeaderWrapper, { props: { label: "Price", field: "price", onclick } });

    await user.click(screen.getByRole("button", { name: /Price/ }));
    expect(onclick).toHaveBeenCalledWith("price");
  });

  it("shows aria-sort ascending when active and ascending", () => {
    render(SortableHeaderWrapper, {
      props: { label: "Price", field: "price", currentSort: "price", ascending: true },
    });
    const th = screen.getByRole("columnheader");
    expect(th.getAttribute("aria-sort")).toBe("ascending");
  });

  it("shows aria-sort descending when active and descending", () => {
    render(SortableHeaderWrapper, {
      props: { label: "Price", field: "price", currentSort: "price", ascending: false },
    });
    const th = screen.getByRole("columnheader");
    expect(th.getAttribute("aria-sort")).toBe("descending");
  });

  it("shows aria-sort none when not active", () => {
    render(SortableHeaderWrapper, {
      props: { label: "Price", field: "price", currentSort: "name", ascending: true },
    });
    const th = screen.getByRole("columnheader");
    expect(th.getAttribute("aria-sort")).toBe("none");
  });
});
