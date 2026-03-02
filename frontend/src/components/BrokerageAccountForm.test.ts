import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import BrokerageAccountForm from "./BrokerageAccountForm.svelte";

const mockCreateBrokerageAccount = vi.fn();
vi.mock("../../wailsjs/go/backend/App", () => ({
  CreateBrokerageAccount: (...args: unknown[]) => mockCreateBrokerageAccount(...args),
}));

describe("BrokerageAccountForm", () => {
  beforeEach(() => {
    mockCreateBrokerageAccount.mockReset();
  });

  it("renders all form fields", () => {
    render(BrokerageAccountForm, { props: { onCreated: vi.fn() } });

    expect(screen.getByLabelText(/broker name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/buy fee/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/sell fee/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create/i })).toBeInTheDocument();
  });

  it("submits valid form data", async () => {
    const onCreated = vi.fn();
    mockCreateBrokerageAccount.mockResolvedValueOnce({
      id: "b1",
      brokerName: "Ajaib",
      buyFeePct: 0.15,
      sellFeePct: 0.25,
    });

    render(BrokerageAccountForm, { props: { onCreated } });
    const user = userEvent.setup();

    await user.type(screen.getByLabelText(/broker name/i), "Ajaib");
    await user.clear(screen.getByLabelText(/buy fee/i));
    await user.type(screen.getByLabelText(/buy fee/i), "0.15");
    await user.clear(screen.getByLabelText(/sell fee/i));
    await user.type(screen.getByLabelText(/sell fee/i), "0.25");
    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(mockCreateBrokerageAccount).toHaveBeenCalledWith("Ajaib", 0.15, 0.25);
    expect(onCreated).toHaveBeenCalled();
  });

  it("shows error when name is empty", async () => {
    render(BrokerageAccountForm, { props: { onCreated: vi.fn() } });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/broker name is required/i);
    expect(mockCreateBrokerageAccount).not.toHaveBeenCalled();
  });

  it("shows error on API failure", async () => {
    mockCreateBrokerageAccount.mockRejectedValueOnce(new Error("network error"));

    render(BrokerageAccountForm, { props: { onCreated: vi.fn() } });
    const user = userEvent.setup();

    await user.type(screen.getByLabelText(/broker name/i), "Ajaib");
    await user.clear(screen.getByLabelText(/buy fee/i));
    await user.type(screen.getByLabelText(/buy fee/i), "0.15");
    await user.clear(screen.getByLabelText(/sell fee/i));
    await user.type(screen.getByLabelText(/sell fee/i), "0.25");
    await user.click(screen.getByRole("button", { name: /create/i }));

    expect(screen.getByRole("alert")).toHaveTextContent(/network error/i);
  });
});
