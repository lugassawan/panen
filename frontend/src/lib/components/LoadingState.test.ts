import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import LoadingStateWrapper from "./__tests__/LoadingStateWrapper.svelte";

describe("LoadingState", () => {
  it("renders spinner with role status", () => {
    render(LoadingStateWrapper);
    expect(screen.getByRole("status")).toBeInTheDocument();
  });

  it("shows message text when provided", () => {
    render(LoadingStateWrapper, { props: { message: "Loading data..." } });
    expect(screen.getByText("Loading data...")).toBeInTheDocument();
  });

  it("does not render message text when omitted", () => {
    const { container } = render(LoadingStateWrapper);
    expect(container.querySelector("span")).toBeNull();
  });

  it("applies size sm with smaller icon and text-sm class", () => {
    const { container } = render(LoadingStateWrapper, {
      props: { size: "sm", message: "Loading..." },
    });
    const svg = container.querySelector("svg");
    expect(svg?.getAttribute("width")).toBe("16");
    expect(svg?.getAttribute("height")).toBe("16");
    const span = container.querySelector("span");
    expect(span?.classList.contains("text-sm")).toBe(true);
  });

  it("applies custom class prop", () => {
    render(LoadingStateWrapper, {
      props: { message: "Loading...", className: "flex-1 py-16" },
    });
    const status = screen.getByRole("status");
    expect(status.classList.contains("flex-1")).toBe(true);
    expect(status.classList.contains("py-16")).toBe(true);
  });
});
