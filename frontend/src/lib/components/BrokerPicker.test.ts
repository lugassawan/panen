import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { BrokerConfigResponse } from "../types";
import BrokerPicker from "./BrokerPicker.svelte";

const brokerConfigs: BrokerConfigResponse[] = [
  {
    code: "XC",
    name: "Ajaib Sekuritas",
    buyFeePct: 0.15,
    sellFeePct: 0.15,
    sellTaxPct: 0.1,
    notes: "",
  },
  {
    code: "XL",
    name: "Stockbit Sekuritas",
    buyFeePct: 0.15,
    sellFeePct: 0.15,
    sellTaxPct: 0.1,
    notes: "",
  },
];

describe("BrokerPicker", () => {
  it("renders broker items with fee details", async () => {
    render(BrokerPicker, { props: { brokerConfigs } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));

    expect(screen.getByText("Ajaib Sekuritas")).toBeInTheDocument();
    expect(screen.getByText("(XC)")).toBeInTheDocument();
    expect(screen.getByText("Stockbit Sekuritas")).toBeInTheDocument();
  });

  it("filters by name", async () => {
    render(BrokerPicker, { props: { brokerConfigs } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByRole("combobox"), "ajaib");

    expect(screen.getAllByRole("option")).toHaveLength(1);
    expect(screen.getByText("Ajaib Sekuritas")).toBeInTheDocument();
  });

  it("filters by code", async () => {
    render(BrokerPicker, { props: { brokerConfigs } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByRole("combobox"), "XL");

    expect(screen.getAllByRole("option")).toHaveLength(1);
    expect(screen.getByText("Stockbit Sekuritas")).toBeInTheDocument();
  });

  it("shows Other at bottom", async () => {
    render(BrokerPicker, { props: { brokerConfigs } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));

    const otherBtn = screen.getByRole("button", { name: /other.*manual/i });
    expect(otherBtn).toBeInTheDocument();
  });

  it("selects broker", async () => {
    const onselect = vi.fn();
    render(BrokerPicker, { props: { brokerConfigs, onselect } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByText("Ajaib Sekuritas"));

    expect(onselect).toHaveBeenCalledWith("XC");
  });

  it("selects Other by click", async () => {
    const onselect = vi.fn();
    render(BrokerPicker, { props: { brokerConfigs, onselect } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByRole("button", { name: /other.*manual/i }));

    expect(onselect).toHaveBeenCalledWith("OTHER");
  });

  it("selects Other by keyboard", async () => {
    const onselect = vi.fn();
    render(BrokerPicker, { props: { brokerConfigs, onselect } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("combobox"));
    // Navigate past 2 brokers to reach footer
    await user.keyboard("{ArrowDown}{ArrowDown}{Enter}");

    expect(onselect).toHaveBeenCalledWith("OTHER");
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
  });
});
