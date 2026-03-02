import { render, screen, within } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type {
  BrokerageAccountResponse,
  HoldingDetailResponse,
  PortfolioDetailResponse,
  PortfolioResponse,
} from "../lib/types";
import PortfolioPage from "./PortfolioPage.svelte";

const mockListBrokerageAccounts = vi.fn();
const mockListPortfolios = vi.fn();
const mockGetPortfolio = vi.fn();
const mockCreateBrokerageAccount = vi.fn();
const mockCreatePortfolio = vi.fn();
const mockAddHolding = vi.fn();

vi.mock("../../wailsjs/go/backend/App", () => ({
  ListBrokerageAccounts: (...args: unknown[]) => mockListBrokerageAccounts(...args),
  ListPortfolios: (...args: unknown[]) => mockListPortfolios(...args),
  GetPortfolio: (...args: unknown[]) => mockGetPortfolio(...args),
  CreateBrokerageAccount: (...args: unknown[]) => mockCreateBrokerageAccount(...args),
  CreatePortfolio: (...args: unknown[]) => mockCreatePortfolio(...args),
  AddHolding: (...args: unknown[]) => mockAddHolding(...args),
}));

function makeBrokerage(
  overrides: Partial<BrokerageAccountResponse> = {},
): BrokerageAccountResponse {
  return {
    id: "b1",
    brokerName: "Ajaib",
    buyFeePct: 0.15,
    sellFeePct: 0.25,
    isManualFee: false,
    createdAt: "2025-01-01T00:00:00Z",
    updatedAt: "2025-01-01T00:00:00Z",
    ...overrides,
  };
}

function makePortfolio(overrides: Partial<PortfolioResponse> = {}): PortfolioResponse {
  return {
    id: "p1",
    brokerageAcctId: "b1",
    name: "Value Portfolio",
    mode: "VALUE",
    riskProfile: "MODERATE",
    capital: 10000000,
    monthlyAddition: 1000000,
    maxStocks: 10,
    createdAt: "2025-01-01T00:00:00Z",
    updatedAt: "2025-01-01T00:00:00Z",
    ...overrides,
  };
}

function makeHolding(overrides: Partial<HoldingDetailResponse> = {}): HoldingDetailResponse {
  return {
    id: "h1",
    ticker: "BBCA",
    avgBuyPrice: 8500,
    lots: 10,
    ...overrides,
  };
}

function makePortfolioDetail(
  overrides: Partial<PortfolioDetailResponse> = {},
): PortfolioDetailResponse {
  return {
    portfolio: makePortfolio(),
    holdings: [makeHolding()],
    ...overrides,
  };
}

describe("PortfolioPage", () => {
  beforeEach(() => {
    mockListBrokerageAccounts.mockReset();
    mockListPortfolios.mockReset();
    mockGetPortfolio.mockReset();
    mockCreateBrokerageAccount.mockReset();
    mockCreatePortfolio.mockReset();
    mockAddHolding.mockReset();
  });

  it("shows loading state on mount", () => {
    mockListBrokerageAccounts.mockReturnValue(new Promise(() => {}));
    render(PortfolioPage);

    expect(screen.getByRole("status")).toBeInTheDocument();
  });

  describe("State A: Onboarding (no brokerage accounts)", () => {
    beforeEach(() => {
      mockListBrokerageAccounts.mockResolvedValue([]);
    });

    it("shows onboarding wizard with brokerage form", async () => {
      render(PortfolioPage);

      expect(await screen.findByText(/set up your brokerage/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/broker name/i)).toBeInTheDocument();
    });

    it("advances to portfolio form after creating brokerage", async () => {
      mockCreateBrokerageAccount.mockResolvedValue(makeBrokerage());
      render(PortfolioPage);
      const user = userEvent.setup();

      await screen.findByText(/set up your brokerage/i);
      await user.type(screen.getByLabelText(/broker name/i), "Ajaib");
      await user.click(screen.getByRole("button", { name: /create account/i }));

      expect(await screen.findByText(/create your portfolio/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument();
    });
  });

  describe("State B: Create portfolio (brokerage exists, no portfolios)", () => {
    beforeEach(() => {
      mockListBrokerageAccounts.mockResolvedValue([makeBrokerage()]);
      mockListPortfolios.mockResolvedValue([]);
    });

    it("shows portfolio form", async () => {
      render(PortfolioPage);

      expect(await screen.findByText(/create your portfolio/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument();
    });
  });

  describe("State C: Portfolio view (portfolio exists)", () => {
    beforeEach(() => {
      mockListBrokerageAccounts.mockResolvedValue([makeBrokerage()]);
      mockListPortfolios.mockResolvedValue([makePortfolio()]);
    });

    it("shows holdings table with data", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding({ currentPrice: 9500, verdict: "UNDERVALUED" })],
        }),
      );
      render(PortfolioPage);

      expect(await screen.findByText("BBCA")).toBeInTheDocument();
      const table = screen.getByRole("table");
      expect(within(table).getByText("BBCA")).toBeInTheDocument();
    });

    it("shows P/L % with color coding for gain", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [
            makeHolding({
              avgBuyPrice: 8500,
              lots: 10,
              currentPrice: 9500,
            }),
          ],
        }),
      );
      render(PortfolioPage);

      await screen.findByText("BBCA");
      // P/L = (9500 - 8500) / 8500 * 100 = 11.76%
      const plCell = screen.getByTestId("pl-BBCA");
      expect(plCell.textContent).toMatch(/11[.,]76/);
      expect(plCell.className).toMatch(/text-emerald/);
    });

    it("shows P/L % with color coding for loss", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [
            makeHolding({
              avgBuyPrice: 9500,
              lots: 10,
              currentPrice: 8500,
            }),
          ],
        }),
      );
      render(PortfolioPage);

      await screen.findByText("BBCA");
      const plCell = screen.getByTestId("pl-BBCA");
      expect(plCell.textContent).toMatch(/10[.,]53/);
      expect(plCell.className).toMatch(/text-red/);
    });

    it("shows verdict badge", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding({ verdict: "UNDERVALUED", currentPrice: 9500 })],
        }),
      );
      render(PortfolioPage);

      expect(await screen.findByText("Undervalued")).toBeInTheDocument();
    });

    it("shows 'Consider Selling' signal for overvalued above exit target", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [
            makeHolding({
              verdict: "OVERVALUED",
              currentPrice: 12000,
              exitTarget: 10000,
            }),
          ],
        }),
      );
      render(PortfolioPage);

      expect(await screen.findByText("Consider Selling")).toBeInTheDocument();
    });

    it("shows 'Hold / Add' signal for undervalued", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding({ verdict: "UNDERVALUED", currentPrice: 8000 })],
        }),
      );
      render(PortfolioPage);

      expect(await screen.findByText("Hold / Add")).toBeInTheDocument();
    });

    it("shows 'Hold' signal for fair value", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding({ verdict: "FAIR", currentPrice: 9000 })],
        }),
      );
      render(PortfolioPage);

      expect(await screen.findByText("Hold")).toBeInTheDocument();
    });

    it("shows dash when no current price", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding()],
        }),
      );
      render(PortfolioPage);

      await screen.findByText("BBCA");
      const table = screen.getByRole("table");
      const dashes = within(table).getAllByText("\u2014");
      expect(dashes.length).toBeGreaterThanOrEqual(1);
    });

    it("shows portfolio summary with totals", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [
            makeHolding({
              avgBuyPrice: 8500,
              lots: 10,
              currentPrice: 9500,
            }),
          ],
        }),
      );
      render(PortfolioPage);

      // Total Invested = 8500 * 10 * 100 = 8,500,000
      await screen.findByText("BBCA");
      expect(screen.getByTestId("total-invested")).toBeInTheDocument();
      expect(screen.getByTestId("current-value")).toBeInTheDocument();
      expect(screen.getByTestId("overall-pl")).toBeInTheDocument();
    });

    it("shows add holding form", async () => {
      mockGetPortfolio.mockResolvedValue(makePortfolioDetail());
      render(PortfolioPage);

      await screen.findByText("BBCA");
      expect(screen.getByLabelText(/ticker/i)).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /add holding/i })).toBeInTheDocument();
    });
  });

  it("shows error state on API failure", async () => {
    mockListBrokerageAccounts.mockRejectedValue(new Error("connection failed"));
    render(PortfolioPage);

    expect(await screen.findByRole("alert")).toHaveTextContent(/connection failed/i);
  });
});
