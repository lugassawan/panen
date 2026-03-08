import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import EmptyStateWrapper from "./__tests__/EmptyStateWrapper.svelte";

describe("EmptyState", () => {
  it("renders title", () => {
    render(EmptyStateWrapper, { props: { title: "No holdings" } });
    expect(screen.getByText("No holdings")).toBeInTheDocument();
  });

  it("renders description when provided", () => {
    render(EmptyStateWrapper, {
      props: { title: "Empty", description: "Add some items to get started." },
    });
    expect(screen.getByText("Add some items to get started.")).toBeInTheDocument();
  });

  it("does not render description when not provided", () => {
    const { container } = render(EmptyStateWrapper, { props: { title: "Empty" } });
    const paragraphs = container.querySelectorAll("p");
    expect(paragraphs).toHaveLength(0);
  });

  it("renders action slot when provided", () => {
    render(EmptyStateWrapper, {
      props: { title: "Empty", showAction: true },
    });
    expect(screen.getByRole("button", { name: "Add item" })).toBeInTheDocument();
  });

  it("applies compact styles when compact is true", () => {
    const { container } = render(EmptyStateWrapper, {
      props: { title: "No records", compact: true },
    });
    const wrapper = container.querySelector("div > div") as HTMLElement;
    expect(wrapper.className).toContain("py-4");
    expect(wrapper.className).not.toContain("py-12");
    const heading = screen.getByText("No records");
    expect(heading.className).toContain("text-sm");
    expect(heading.className).not.toContain("text-lg");
  });

  it("applies default styles when compact is false", () => {
    const { container } = render(EmptyStateWrapper, {
      props: { title: "No items" },
    });
    const wrapper = container.querySelector("div > div") as HTMLElement;
    expect(wrapper.className).toContain("py-12");
    const heading = screen.getByText("No items");
    expect(heading.className).toContain("text-lg");
  });
});
