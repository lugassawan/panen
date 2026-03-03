import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { BrokerConfigResponse } from "../lib/types";
import BrokerageAccountForm from "./BrokerageAccountForm.svelte";

const mockCreateBrokerageAccount = vi.fn();
const mockUpdateBrokerageAccount = vi.fn();
vi.mock("../../wailsjs/go/backend/App", () => ({
  CreateBrokerageAccount: (...args: unknown[]) => mockCreateBrokerageAccount(...args),
  UpdateBrokerageAccount: (...args: unknown[]) => mockUpdateBrokerageAccount(...args),
}));

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

describe("BrokerageAccountForm", () => {
  beforeEach(() => {
    mockCreateBrokerageAccount.mockReset();
    mockUpdateBrokerageAccount.mockReset();
  });

  it("renders all form fields", () => {
    render(BrokerageAccountForm, { props: { onSaved: vi.fn() } });

    expect(screen.getByLabelText(/broker name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/buy fee/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/sell fee/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/sell tax/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create/i })).toBeInTheDocument();
  });

  it("renders broker picker when configs provided", () => {
    render(BrokerageAccountForm, { props: { onSaved: vi.fn(), brokerConfigs } });

    expect(screen.getByLabelText(/broker$/i)).toBeInTheDocument();
    expect(screen.getByText(/ajaib sekuritas/i)).toBeInTheDocument();
  });

  it("auto-fills fees when selecting a broker", async () => {
    render(BrokerageAccountForm, { props: { onSaved: vi.fn(), brokerConfigs } });
    const user = userEvent.setup();

    const select = screen.getByLabelText(/broker$/i);
    await user.selectOptions(select, "XC");

    expect(screen.getByLabelText(/broker name/i)).toHaveValue("Ajaib Sekuritas");
    expect(screen.getByLabelText(/buy fee/i)).toHaveValue(0.15);
    expect(screen.getByLabelText(/sell fee/i)).toHaveValue(0.15);
    expect(screen.getByLabelText(/sell tax/i)).toHaveValue(0.1);
  });

  it("submits valid create form data", async () => {
    const onSaved = vi.fn();
    mockCreateBrokerageAccount.mockResolvedValueOnce({
      id: "b1",
      brokerName: "Ajaib Sekuritas",
      brokerCode: "XC",
      buyFeePct: 0.15,
      sellFeePct: 0.15,
      sellTaxPct: 0.1,
    });

    render(BrokerageAccountForm, { props: { onSaved, brokerConfigs } });
    const user = userEvent.setup();

    await user.selectOptions(screen.getByLabelText(/broker$/i), "XC");
    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(mockCreateBrokerageAccount).toHaveBeenCalledWith(
      "Ajaib Sekuritas",
      "XC",
      0.15,
      0.15,
      0.1,
      false,
    );
    expect(onSaved).toHaveBeenCalled();
  });

  it("shows error when name is empty", async () => {
    render(BrokerageAccountForm, { props: { onSaved: vi.fn() } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/broker name is required/i);
    expect(mockCreateBrokerageAccount).not.toHaveBeenCalled();
  });

  it("shows error on API failure", async () => {
    mockCreateBrokerageAccount.mockRejectedValueOnce(new Error("network error"));

    render(BrokerageAccountForm, { props: { onSaved: vi.fn(), brokerConfigs } });
    const user = userEvent.setup();

    await user.selectOptions(screen.getByLabelText(/broker$/i), "XC");
    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/network error/i);
  });

  it("renders in edit mode with existing account", () => {
    const existingAccount = {
      id: "b1",
      brokerName: "Ajaib Sekuritas",
      brokerCode: "XC",
      buyFeePct: 0.15,
      sellFeePct: 0.15,
      sellTaxPct: 0.1,
      isManualFee: false,
      createdAt: "2025-01-01T00:00:00Z",
      updatedAt: "2025-01-01T00:00:00Z",
    };

    render(BrokerageAccountForm, { props: { onSaved: vi.fn(), brokerConfigs, existingAccount } });

    expect(screen.getByLabelText(/broker name/i)).toHaveValue("Ajaib Sekuritas");
    expect(screen.getByRole("button", { name: /save/i })).toBeInTheDocument();
  });

  it("calls cancel callback", async () => {
    const onCancel = vi.fn();
    render(BrokerageAccountForm, { props: { onSaved: vi.fn(), onCancel } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(onCancel).toHaveBeenCalled();
  });
});
