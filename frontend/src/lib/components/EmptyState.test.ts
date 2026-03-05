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
});
