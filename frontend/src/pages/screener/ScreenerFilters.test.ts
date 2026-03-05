import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import ScreenerFilters from "./ScreenerFilters.svelte";

describe("ScreenerFilters", () => {
  const defaultProps = {
    universeType: "INDEX",
    universeName: "",
    riskProfile: "MODERATE" as const,
    sectorFilter: "",
    customTickers: "",
    indices: ["IDX30", "LQ45"],
    sectors: ["Financials", "Industrials"],
    onrun: vi.fn(),
  };

  it("renders universe type selector", () => {
    render(ScreenerFilters, { props: defaultProps });
    const select = screen.getByLabelText("Universe");
    expect(select).toBeInTheDocument();
  });

  it("shows index dropdown when INDEX selected", () => {
    render(ScreenerFilters, {
      props: { ...defaultProps, universeType: "INDEX" },
    });
    expect(screen.getByLabelText("Index")).toBeInTheDocument();
    expect(screen.getByText("IDX30")).toBeInTheDocument();
    expect(screen.getByText("LQ45")).toBeInTheDocument();
  });

  it("renders risk profile radio buttons", () => {
    render(ScreenerFilters, { props: defaultProps });
    expect(screen.getByRole("radio", { name: "Conservative" })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "Moderate" })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "Aggressive" })).toBeInTheDocument();
  });

  it("marks current risk profile as checked", () => {
    render(ScreenerFilters, { props: defaultProps });
    expect(screen.getByRole("radio", { name: "Moderate" })).toHaveAttribute("aria-checked", "true");
  });

  it("renders run screen button", () => {
    render(ScreenerFilters, { props: defaultProps });
    expect(screen.getByRole("button", { name: /Run Screen/i })).toBeInTheDocument();
  });

  it("disables run button when no universe name selected", () => {
    render(ScreenerFilters, {
      props: { ...defaultProps, universeName: "" },
    });
    expect(screen.getByRole("button", { name: /Run Screen/i })).toBeDisabled();
  });

  it("enables run button when universe name is set", () => {
    render(ScreenerFilters, {
      props: { ...defaultProps, universeName: "IDX30" },
    });
    expect(screen.getByRole("button", { name: /Run Screen/i })).not.toBeDisabled();
  });

  it("calls onrun when Run Screen clicked", async () => {
    const user = userEvent.setup();
    const onrun = vi.fn();
    render(ScreenerFilters, {
      props: { ...defaultProps, universeName: "IDX30", onrun },
    });

    await user.click(screen.getByRole("button", { name: /Run Screen/i }));
    expect(onrun).toHaveBeenCalledOnce();
  });
});
