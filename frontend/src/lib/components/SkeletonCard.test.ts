import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import SkeletonCardWrapper from "./__tests__/SkeletonCardWrapper.svelte";

describe("SkeletonCard", () => {
  it("has role status with loading label", () => {
    render(SkeletonCardWrapper);
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("applies elevated card styling", () => {
    render(SkeletonCardWrapper);
    const el = screen.getByRole("status");
    expect(el.classList.contains("bg-bg-elevated")).toBe(true);
  });

  it("renders 1 title + 3 body skeleton lines by default", () => {
    const { container } = render(SkeletonCardWrapper);
    const skeletons = container.querySelectorAll(".skeleton");
    expect(skeletons).toHaveLength(4); // 1 title + 3 body
  });

  it("renders custom number of body lines", () => {
    const { container } = render(SkeletonCardWrapper, { props: { lines: 6 } });
    const skeletons = container.querySelectorAll(".skeleton");
    expect(skeletons).toHaveLength(7); // 1 title + 6 body
  });
});
