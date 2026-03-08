import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import SkeletonTableWrapper from "./__tests__/SkeletonTableWrapper.svelte";

describe("SkeletonTable", () => {
  it("has role status with loading label", () => {
    render(SkeletonTableWrapper);
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("renders default 5 rows and 4 columns", () => {
    const { container } = render(SkeletonTableWrapper);
    const bodyRows = container.querySelectorAll("tbody tr");
    expect(bodyRows).toHaveLength(5);
    const firstRowCells = bodyRows[0].querySelectorAll("td");
    expect(firstRowCells).toHaveLength(4);
  });

  it("renders custom rows and columns", () => {
    const { container } = render(SkeletonTableWrapper, {
      props: { rows: 3, columns: 6 },
    });
    const bodyRows = container.querySelectorAll("tbody tr");
    expect(bodyRows).toHaveLength(3);
    const firstRowCells = bodyRows[0].querySelectorAll("td");
    expect(firstRowCells).toHaveLength(6);
  });

  it("applies skeleton class to all shimmer elements", () => {
    const { container } = render(SkeletonTableWrapper, {
      props: { rows: 2, columns: 3 },
    });
    const skeletons = container.querySelectorAll(".skeleton");
    // 3 header + (2 rows × 3 columns) = 9
    expect(skeletons).toHaveLength(9);
  });
});
