import { render, screen, within } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import BrokeragePage from "./BrokeragePage.svelte";

const mockListBrokerageAccounts = vi.fn();
const mockListBrokerConfigs = vi.fn();
const mockDeleteBrokerageAccount = vi.fn();
const mockCreateBrokerageAccount = vi.fn();
const mockUpdateBrokerageAccount = vi.fn();

vi.mock("../../wailsjs/go/backend/App", () => ({
  ListBrokerageAccounts: (...args: unknown[]) => mockListBrokerageAccounts(...args),
  ListBrokerConfigs: (...args: unknown[]) => mockListBrokerConfigs(...args),
  DeleteBrokerageAccount: (...args: unknown[]) => mockDeleteBrokerageAccount(...args),
  CreateBrokerageAccount: (...args: unknown[]) => mockCreateBrokerageAccount(...args),
  UpdateBrokerageAccount: (...args: unknown[]) => mockUpdateBrokerageAccount(...args),
}));

describe("BrokeragePage", () => {
  beforeEach(() => {
    mockListBrokerageAccounts.mockReset();
    mockListBrokerConfigs.mockReset();
    mockDeleteBrokerageAccount.mockReset();
    mockCreateBrokerageAccount.mockReset();
    mockUpdateBrokerageAccount.mockReset();

    mockListBrokerConfigs.mockResolvedValue([
      {
        code: "XC",
        name: "Ajaib Sekuritas",
        buyFeePct: 0.15,
        sellFeePct: 0.15,
        sellTaxPct: 0.1,
        notes: "",
      },
    ]);
  });

  it("shows empty state when no accounts", async () => {
    mockListBrokerageAccounts.mockResolvedValue([]);

    render(BrokeragePage);

    expect(await screen.findByText(/no brokerage accounts yet/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /add account/i })).toBeInTheDocument();
  });

  it("renders account list", async () => {
    mockListBrokerageAccounts.mockResolvedValue([
      {
        id: "b1",
        brokerName: "Ajaib Sekuritas",
        brokerCode: "XC",
        buyFeePct: 0.15,
        sellFeePct: 0.15,
        sellTaxPct: 0.1,
        isManualFee: false,
        createdAt: "2025-01-01T00:00:00Z",
        updatedAt: "2025-01-01T00:00:00Z",
      },
    ]);

    render(BrokeragePage);

    expect(await screen.findByText("Ajaib Sekuritas")).toBeInTheDocument();
    expect(screen.getByText("XC")).toBeInTheDocument();
  });

  it("shows create form when Add Account is clicked", async () => {
    mockListBrokerageAccounts.mockResolvedValue([]);
    render(BrokeragePage);

    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: /add account/i }));

    expect(screen.getByText(/new brokerage account/i)).toBeInTheDocument();
  });

  it("shows edit form when Edit is clicked", async () => {
    mockListBrokerageAccounts.mockResolvedValue([
      {
        id: "b1",
        brokerName: "Ajaib Sekuritas",
        brokerCode: "XC",
        buyFeePct: 0.15,
        sellFeePct: 0.15,
        sellTaxPct: 0.1,
        isManualFee: false,
        createdAt: "2025-01-01T00:00:00Z",
        updatedAt: "2025-01-01T00:00:00Z",
      },
    ]);

    render(BrokeragePage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: /edit/i }));

    expect(screen.getByText(/edit brokerage account/i)).toBeInTheDocument();
  });

  it("shows delete confirmation dialog", async () => {
    mockListBrokerageAccounts.mockResolvedValue([
      {
        id: "b1",
        brokerName: "Ajaib Sekuritas",
        brokerCode: "XC",
        buyFeePct: 0.15,
        sellFeePct: 0.15,
        sellTaxPct: 0.1,
        isManualFee: false,
        createdAt: "2025-01-01T00:00:00Z",
        updatedAt: "2025-01-01T00:00:00Z",
      },
    ]);

    render(BrokeragePage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: /delete/i }));

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
  });

  it("deletes account on confirm", async () => {
    mockListBrokerageAccounts
      .mockResolvedValueOnce([
        {
          id: "b1",
          brokerName: "Ajaib Sekuritas",
          brokerCode: "XC",
          buyFeePct: 0.15,
          sellFeePct: 0.15,
          sellTaxPct: 0.1,
          isManualFee: false,
          createdAt: "2025-01-01T00:00:00Z",
          updatedAt: "2025-01-01T00:00:00Z",
        },
      ])
      .mockResolvedValueOnce([]);
    mockDeleteBrokerageAccount.mockResolvedValue(undefined);

    render(BrokeragePage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: /delete/i }));
    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: /delete/i }));

    expect(mockDeleteBrokerageAccount).toHaveBeenCalledWith("b1");
  });

  it("shows error when delete fails with dependent portfolios", async () => {
    mockListBrokerageAccounts.mockResolvedValue([
      {
        id: "b1",
        brokerName: "Ajaib Sekuritas",
        brokerCode: "XC",
        buyFeePct: 0.15,
        sellFeePct: 0.15,
        sellTaxPct: 0.1,
        isManualFee: false,
        createdAt: "2025-01-01T00:00:00Z",
        updatedAt: "2025-01-01T00:00:00Z",
      },
    ]);
    mockDeleteBrokerageAccount.mockRejectedValue(
      new Error("has dependent portfolios: 1 portfolio(s) linked"),
    );

    render(BrokeragePage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: /delete/i }));
    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: /delete/i }));

    expect(await screen.findByText(/dependent portfolios/i)).toBeInTheDocument();
  });

  it("shows error state on load failure", async () => {
    mockListBrokerageAccounts.mockRejectedValue(new Error("connection failed"));

    render(BrokeragePage);

    expect(await screen.findByRole("alert")).toHaveTextContent(/connection failed/i);
  });
});
