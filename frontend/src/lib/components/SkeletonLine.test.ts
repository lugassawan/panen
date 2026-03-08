import { render } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import SkeletonLineWrapper from "./__tests__/SkeletonLineWrapper.svelte";

describe("SkeletonLine", () => {
  it("applies skeleton class", () => {
    const { container } = render(SkeletonLineWrapper);
    const el = container.querySelector(".skeleton");
    expect(el).toBeInTheDocument();
  });

  it("uses default dimensions", () => {
    const { container } = render(SkeletonLineWrapper);
    const el = container.querySelector(".skeleton") as HTMLElement;
    expect(el.style.width).toBe("100%");
    expect(el.style.height).toBe("1rem");
  });

  it("applies custom width and height", () => {
    const { container } = render(SkeletonLineWrapper, {
      props: { width: "40%", height: "1.25rem" },
    });
    const el = container.querySelector(".skeleton") as HTMLElement;
    expect(el.style.width).toBe("40%");
    expect(el.style.height).toBe("1.25rem");
  });
});
