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
const mockListBrokerConfigs = vi.fn();
const mockListPortfolios = vi.fn();
const mockGetPortfolio = vi.fn();
const mockCreateBrokerageAccount = vi.fn();
const mockCreatePortfolio = vi.fn();
const mockAddHolding = vi.fn();
const mockDeletePortfolio = vi.fn();
const mockUpdatePortfolio = vi.fn();
const mockAvailableActions = vi.fn();
const mockEvaluateChecklist = vi.fn();
const mockToggleManualCheck = vi.fn();
const mockResetChecklist = vi.fn();

vi.mock("../../wailsjs/go/backend/App", () => ({
  ListBrokerageAccounts: (...args: unknown[]) => mockListBrokerageAccounts(...args),
  ListBrokerConfigs: (...args: unknown[]) => mockListBrokerConfigs(...args),
  ListPortfolios: (...args: unknown[]) => mockListPortfolios(...args),
  GetPortfolio: (...args: unknown[]) => mockGetPortfolio(...args),
  CreateBrokerageAccount: (...args: unknown[]) => mockCreateBrokerageAccount(...args),
  CreatePortfolio: (...args: unknown[]) => mockCreatePortfolio(...args),
  AddHolding: (...args: unknown[]) => mockAddHolding(...args),
  DeletePortfolio: (...args: unknown[]) => mockDeletePortfolio(...args),
  UpdatePortfolio: (...args: unknown[]) => mockUpdatePortfolio(...args),
  AvailableActions: (...args: unknown[]) => mockAvailableActions(...args),
  EvaluateChecklist: (...args: unknown[]) => mockEvaluateChecklist(...args),
  ToggleManualCheck: (...args: unknown[]) => mockToggleManualCheck(...args),
  ResetChecklist: (...args: unknown[]) => mockResetChecklist(...args),
}));

function makeBrokerage(
  overrides: Partial<BrokerageAccountResponse> = {},
): BrokerageAccountResponse {
  return {
    id: "b1",
    brokerName: "Ajaib",
    brokerCode: "XC",
    buyFeePct: 0.15,
    sellFeePct: 0.15,
    sellTaxPct: 0.1,
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
    mockListBrokerConfigs.mockReset();
    mockListPortfolios.mockReset();
    mockGetPortfolio.mockReset();
    mockCreateBrokerageAccount.mockReset();
    mockCreatePortfolio.mockReset();
    mockAddHolding.mockReset();
    mockDeletePortfolio.mockReset();
    mockUpdatePortfolio.mockReset();

    mockListBrokerConfigs.mockResolvedValue([]);
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

  describe("State C: Portfolio list (portfolios exist)", () => {
    beforeEach(() => {
      mockListBrokerageAccounts.mockResolvedValue([makeBrokerage()]);
    });

    it("shows portfolio list when portfolios exist", async () => {
      mockListPortfolios.mockResolvedValue([makePortfolio()]);
      render(PortfolioPage);

      expect(await screen.findByText("Value Portfolio")).toBeInTheDocument();
      expect(screen.getByTestId("portfolio-card")).toBeInTheDocument();
    });

    it("shows mode badge on portfolio cards", async () => {
      mockListPortfolios.mockResolvedValue([
        makePortfolio({ mode: "VALUE" }),
        makePortfolio({ id: "p2", name: "Div Portfolio", mode: "DIVIDEND" }),
      ]);
      render(PortfolioPage);

      await screen.findByText("Value Portfolio");
      const badges = screen.getAllByTestId("mode-badge");
      expect(badges[0]).toHaveTextContent("Value");
      expect(badges[1]).toHaveTextContent("Dividend");
    });

    it("shows New Portfolio button when < 2 portfolios", async () => {
      mockListPortfolios.mockResolvedValue([makePortfolio()]);
      render(PortfolioPage);

      await screen.findByText("Value Portfolio");
      expect(screen.getByRole("button", { name: /new portfolio/i })).toBeInTheDocument();
    });

    it("hides New Portfolio button when 2 portfolios exist", async () => {
      mockListPortfolios.mockResolvedValue([
        makePortfolio(),
        makePortfolio({ id: "p2", name: "Div Portfolio", mode: "DIVIDEND" }),
      ]);
      render(PortfolioPage);

      await screen.findByText("Value Portfolio");
      expect(screen.queryByRole("button", { name: /new portfolio/i })).not.toBeInTheDocument();
    });

    it("navigates to edit form on Edit click", async () => {
      mockListPortfolios.mockResolvedValue([makePortfolio()]);
      render(PortfolioPage);
      const user = userEvent.setup();

      await screen.findByText("Value Portfolio");
      await user.click(screen.getByRole("button", { name: /edit/i }));

      expect(screen.getByText(/edit portfolio/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/portfolio name/i)).toHaveValue("Value Portfolio");
    });

    it("shows confirm dialog on Delete click", async () => {
      mockListPortfolios.mockResolvedValue([makePortfolio()]);
      render(PortfolioPage);
      const user = userEvent.setup();

      await screen.findByText("Value Portfolio");
      await user.click(screen.getByRole("button", { name: /delete/i }));

      expect(screen.getByRole("dialog")).toBeInTheDocument();
      expect(screen.getByText(/are you sure you want to delete/i)).toBeInTheDocument();
    });

    it("calls DeletePortfolio on confirm", async () => {
      mockDeletePortfolio.mockResolvedValueOnce(undefined);
      mockListPortfolios.mockResolvedValueOnce([makePortfolio()]).mockResolvedValueOnce([]);
      render(PortfolioPage);
      const user = userEvent.setup();

      await screen.findByText("Value Portfolio");
      await user.click(screen.getByRole("button", { name: /delete/i }));

      const dialog = screen.getByRole("dialog");
      await user.click(within(dialog).getByRole("button", { name: /delete/i }));

      expect(mockDeletePortfolio).toHaveBeenCalledWith("p1");
    });

    it("shows delete error in dialog", async () => {
      mockDeletePortfolio.mockRejectedValueOnce(
        new Error("portfolio has holdings: 1 holding(s) linked"),
      );
      mockListPortfolios.mockResolvedValue([makePortfolio()]);
      render(PortfolioPage);
      const user = userEvent.setup();

      await screen.findByText("Value Portfolio");
      await user.click(screen.getByRole("button", { name: /delete/i }));

      const dialog = screen.getByRole("dialog");
      await user.click(within(dialog).getByRole("button", { name: /delete/i }));

      expect(await within(dialog).findByRole("alert")).toHaveTextContent(/portfolio has holdings/i);
    });

    it("navigates to detail view on portfolio card click", async () => {
      mockListPortfolios.mockResolvedValue([makePortfolio()]);
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding({ currentPrice: 9500, verdict: "UNDERVALUED" })],
        }),
      );
      render(PortfolioPage);
      const user = userEvent.setup();

      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

      expect(await screen.findByText("BBCA")).toBeInTheDocument();
      expect(screen.getByRole("table")).toBeInTheDocument();
    });

    it("back button returns to list from detail view", async () => {
      mockListPortfolios.mockResolvedValue([makePortfolio()]);
      mockGetPortfolio.mockResolvedValue(makePortfolioDetail());
      render(PortfolioPage);
      const user = userEvent.setup();

      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

      await screen.findByText("BBCA");
      await user.click(screen.getByLabelText(/back to list/i));

      expect(await screen.findByTestId("portfolio-card")).toBeInTheDocument();
    });
  });

  describe("State D: Portfolio view (detail)", () => {
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
      // Navigate to detail
      render(PortfolioPage);
      const user = userEvent.setup();
      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

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
      const user = userEvent.setup();
      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

      await screen.findByText("BBCA");
      const plCell = screen.getByTestId("pl-BBCA");
      expect(plCell.textContent).toMatch(/11[.,]76/);
      expect(plCell.className).toMatch(/text-profit/);
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
      const user = userEvent.setup();
      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

      await screen.findByText("BBCA");
      const plCell = screen.getByTestId("pl-BBCA");
      expect(plCell.textContent).toMatch(/10[.,]53/);
      expect(plCell.className).toMatch(/text-loss/);
    });

    it("shows verdict badge", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding({ verdict: "UNDERVALUED", currentPrice: 9500 })],
        }),
      );
      render(PortfolioPage);
      const user = userEvent.setup();
      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

      expect(await screen.findByText("Undervalued")).toBeInTheDocument();
    });

    it("shows Checklist button for each holding", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding({ verdict: "UNDERVALUED", currentPrice: 8000 })],
        }),
      );
      render(PortfolioPage);
      const user = userEvent.setup();
      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

      expect(await screen.findByRole("button", { name: /checklist/i })).toBeInTheDocument();
    });

    it("shows dash when no current price", async () => {
      mockGetPortfolio.mockResolvedValue(
        makePortfolioDetail({
          holdings: [makeHolding()],
        }),
      );
      render(PortfolioPage);
      const user = userEvent.setup();
      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

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
      const user = userEvent.setup();
      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

      await screen.findByText("BBCA");
      expect(screen.getByTestId("total-invested")).toBeInTheDocument();
      expect(screen.getByTestId("current-value")).toBeInTheDocument();
      expect(screen.getByTestId("overall-pl")).toBeInTheDocument();
    });

    it("shows add holding form", async () => {
      mockGetPortfolio.mockResolvedValue(makePortfolioDetail());
      render(PortfolioPage);
      const user = userEvent.setup();
      await screen.findByText("Value Portfolio");
      const card = screen.getByTestId("portfolio-card");
      await user.click(within(card).getByRole("button", { name: /value portfolio/i }));

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
