import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import BadgeWrapper from "./__tests__/BadgeWrapper.svelte";

describe("Badge", () => {
  it("renders badge text", () => {
    render(BadgeWrapper, { props: { text: "Active" } });
    expect(screen.getByText("Active")).toBeInTheDocument();
  });

  it("defaults to value variant", () => {
    render(BadgeWrapper, { props: { text: "Default" } });
    const badge = screen.getByText("Default");
    expect(badge.className).toContain("bg-green-100");
  });

  it.each([
    ["value", "bg-green-100"],
    ["dividend", "bg-gold-100"],
    ["profit", "bg-positive-bg"],
    ["loss", "bg-negative-bg"],
    ["warning", "bg-warning-bg"],
  ] as const)("renders %s variant with class %s", (variant, expectedClass) => {
    render(BadgeWrapper, { props: { text: "Test", variant } });
    const badge = screen.getByText("Test");
    expect(badge.className).toContain(expectedClass);
  });
});
