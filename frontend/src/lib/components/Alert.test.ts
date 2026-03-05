import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import AlertWrapper from "./__tests__/AlertWrapper.svelte";

describe("Alert", () => {
  it("renders alert text", () => {
    render(AlertWrapper, { props: { text: "Something happened" } });
    expect(screen.getByRole("alert")).toHaveTextContent("Something happened");
  });

  it("defaults to info variant", () => {
    render(AlertWrapper);
    const alert = screen.getByRole("alert");
    expect(alert.className).toContain("text-info");
  });

  it.each(["positive", "warning", "negative", "info"] as const)("renders %s variant", (variant) => {
    render(AlertWrapper, { props: { variant } });
    const alert = screen.getByRole("alert");
    expect(alert.className).toContain(`text-${variant}`);
  });

  it("hides dismiss button by default", () => {
    render(AlertWrapper);
    expect(screen.queryByLabelText("Dismiss")).not.toBeInTheDocument();
  });

  it("shows dismiss button when dismissible", () => {
    render(AlertWrapper, { props: { dismissible: true } });
    expect(screen.getByLabelText("Dismiss")).toBeInTheDocument();
  });

  it("dismisses alert on button click", async () => {
    const user = userEvent.setup();
    render(AlertWrapper, { props: { dismissible: true, text: "Gone soon" } });

    expect(screen.getByRole("alert")).toBeInTheDocument();
    await user.click(screen.getByLabelText("Dismiss"));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
