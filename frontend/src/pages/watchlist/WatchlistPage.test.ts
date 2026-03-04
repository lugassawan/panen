import { render, screen, within } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import WatchlistPage from "./WatchlistPage.svelte";

const mockListWatchlists = vi.fn();
const mockCreateWatchlist = vi.fn();
const mockDeleteWatchlist = vi.fn();
const mockAddToWatchlist = vi.fn();
const mockRemoveFromWatchlist = vi.fn();
const mockGetWatchlistItems = vi.fn();
const mockGetPresetItems = vi.fn();
const mockListIndexNames = vi.fn();
const mockListWatchlistSectors = vi.fn();
const mockRenameWatchlist = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  ListWatchlists: (...args: unknown[]) => mockListWatchlists(...args),
  CreateWatchlist: (...args: unknown[]) => mockCreateWatchlist(...args),
  DeleteWatchlist: (...args: unknown[]) => mockDeleteWatchlist(...args),
  AddToWatchlist: (...args: unknown[]) => mockAddToWatchlist(...args),
  RemoveFromWatchlist: (...args: unknown[]) => mockRemoveFromWatchlist(...args),
  GetWatchlistItems: (...args: unknown[]) => mockGetWatchlistItems(...args),
  GetPresetItems: (...args: unknown[]) => mockGetPresetItems(...args),
  ListIndexNames: (...args: unknown[]) => mockListIndexNames(...args),
  ListWatchlistSectors: (...args: unknown[]) => mockListWatchlistSectors(...args),
  RenameWatchlist: (...args: unknown[]) => mockRenameWatchlist(...args),
}));

const mockWatchlists = [
  {
    id: "wl1",
    name: "My Banking",
    createdAt: "2025-01-01T00:00:00Z",
    updatedAt: "2025-01-01T00:00:00Z",
  },
];

const mockIndexNames = ["IDX30", "LQ45"];

const mockPresetItems = [
  {
    ticker: "BBCA",
    sector: "Banking",
    price: 8500,
    roe: 18.5,
    der: 5.2,
    eps: 500,
    dividendYield: 2.1,
    verdict: "UNDERVALUED",
  },
  {
    ticker: "TLKM",
    sector: "Telco",
    price: 4200,
    roe: 12.0,
    der: 1.1,
    eps: 300,
    dividendYield: 4.5,
    verdict: "FAIR",
  },
];

const mockWatchlistItems = [
  {
    ticker: "BBCA",
    sector: "Banking",
    price: 8500,
    roe: 18.5,
    der: 5.2,
    eps: 500,
    dividendYield: 2.1,
    verdict: "UNDERVALUED",
  },
  {
    ticker: "BBRI",
    sector: "Banking",
  },
];

const mockSectors = ["Banking", "Telco"];

describe("WatchlistPage", () => {
  beforeEach(() => {
    mockListWatchlists.mockReset();
    mockCreateWatchlist.mockReset();
    mockDeleteWatchlist.mockReset();
    mockAddToWatchlist.mockReset();
    mockRemoveFromWatchlist.mockReset();
    mockGetWatchlistItems.mockReset();
    mockGetPresetItems.mockReset();
    mockListIndexNames.mockReset();
    mockListWatchlistSectors.mockReset();
    mockRenameWatchlist.mockReset();

    mockListWatchlistSectors.mockResolvedValue(mockSectors);
  });

  it("shows loading state on mount", () => {
    mockListWatchlists.mockReturnValue(new Promise(() => {}));
    mockListIndexNames.mockReturnValue(new Promise(() => {}));

    render(WatchlistPage);

    const spinners = screen.getAllByRole("status");
    expect(spinners.length).toBeGreaterThanOrEqual(1);
  });

  it("shows preset index names in left panel after load", async () => {
    mockListWatchlists.mockResolvedValue([]);
    mockListIndexNames.mockResolvedValue(mockIndexNames);

    render(WatchlistPage);

    expect(await screen.findByRole("button", { name: "IDX30" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "LQ45" })).toBeInTheDocument();
  });

  it("shows custom watchlists in left panel after load", async () => {
    mockListWatchlists.mockResolvedValue(mockWatchlists);
    mockListIndexNames.mockResolvedValue([]);

    render(WatchlistPage);

    expect(await screen.findByRole("button", { name: "My Banking" })).toBeInTheDocument();
  });

  it("shows empty state when nothing is selected", async () => {
    mockListWatchlists.mockResolvedValue([]);
    mockListIndexNames.mockResolvedValue([]);

    render(WatchlistPage);

    expect(await screen.findByText(/select a watchlist or index/i)).toBeInTheDocument();
  });

  it("shows preset items table when a preset index is clicked", async () => {
    mockListWatchlists.mockResolvedValue([]);
    mockListIndexNames.mockResolvedValue(mockIndexNames);
    mockGetPresetItems.mockResolvedValue(mockPresetItems);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: "IDX30" }));

    expect(mockGetPresetItems).toHaveBeenCalledWith("IDX30", "");
    expect(await screen.findByRole("table", { name: /watchlist items/i })).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("TLKM")).toBeInTheDocument();
  });

  it("shows custom watchlist items when a watchlist is clicked", async () => {
    mockListWatchlists.mockResolvedValue(mockWatchlists);
    mockListIndexNames.mockResolvedValue([]);
    mockGetWatchlistItems.mockResolvedValue(mockWatchlistItems);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: "My Banking" }));

    expect(mockGetWatchlistItems).toHaveBeenCalledWith("wl1", "");
    expect(await screen.findByRole("table", { name: /watchlist items/i })).toBeInTheDocument();
    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.getByText("BBRI")).toBeInTheDocument();
  });

  it("shows sector filter chips and filters items by sector", async () => {
    mockListWatchlists.mockResolvedValue([]);
    mockListIndexNames.mockResolvedValue(mockIndexNames);
    mockGetPresetItems.mockResolvedValue(mockPresetItems);
    mockListWatchlistSectors.mockResolvedValue(mockSectors);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: "IDX30" }));

    await screen.findByRole("table", { name: /watchlist items/i });

    expect(screen.getByRole("button", { name: "Banking" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Telco" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Banking" }));

    expect(screen.getByText("BBCA")).toBeInTheDocument();
    expect(screen.queryByText("TLKM")).not.toBeInTheDocument();
  });

  it("opens create form when + button is clicked and submits", async () => {
    mockListWatchlists
      .mockResolvedValue([])
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([
        {
          id: "wl2",
          name: "My Telco",
          createdAt: "2025-01-01T00:00:00Z",
          updatedAt: "2025-01-01T00:00:00Z",
        },
      ]);
    mockListIndexNames.mockResolvedValue([]);
    mockCreateWatchlist.mockResolvedValue(undefined);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: /new watchlist/i }));

    expect(screen.getByLabelText(/new watchlist name/i)).toBeInTheDocument();

    await user.type(screen.getByLabelText(/new watchlist name/i), "My Telco");
    await user.click(screen.getByRole("button", { name: /^add$/i }));

    expect(mockCreateWatchlist).toHaveBeenCalledWith("My Telco");
  });

  it("shows confirm dialog and calls DeleteWatchlist on confirm", async () => {
    mockListWatchlists.mockResolvedValueOnce(mockWatchlists).mockResolvedValueOnce([]);
    mockListIndexNames.mockResolvedValue([]);
    mockDeleteWatchlist.mockResolvedValue(undefined);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: /delete my banking/i }));

    const dialog = screen.getByRole("dialog");
    expect(dialog).toBeInTheDocument();
    expect(within(dialog).getByText(/are you sure/i)).toBeInTheDocument();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    expect(mockDeleteWatchlist).toHaveBeenCalledWith("wl1");
  });

  it("adds a ticker when the add ticker form is submitted", async () => {
    mockListWatchlists.mockResolvedValue(mockWatchlists);
    mockListIndexNames.mockResolvedValue([]);
    mockGetWatchlistItems
      .mockResolvedValueOnce(mockWatchlistItems)
      .mockResolvedValueOnce([
        ...mockWatchlistItems,
        { ticker: "BMRI", sector: "Banking", price: 6000 },
      ]);
    mockAddToWatchlist.mockResolvedValue(undefined);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: "My Banking" }));

    await screen.findByRole("table", { name: /watchlist items/i });

    await user.type(screen.getByLabelText(/add ticker to watchlist/i), "bmri");
    await user.click(screen.getByRole("button", { name: /^add$/i }));

    expect(mockAddToWatchlist).toHaveBeenCalledWith("wl1", "BMRI");
  });

  it("removes a ticker when the remove button is clicked", async () => {
    mockListWatchlists.mockResolvedValue(mockWatchlists);
    mockListIndexNames.mockResolvedValue([]);
    mockGetWatchlistItems
      .mockResolvedValueOnce(mockWatchlistItems)
      .mockResolvedValueOnce([mockWatchlistItems[1]]);
    mockRemoveFromWatchlist.mockResolvedValue(undefined);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: "My Banking" }));

    await screen.findByRole("table", { name: /watchlist items/i });

    await user.click(screen.getByRole("button", { name: /remove bbca/i }));

    expect(mockRemoveFromWatchlist).toHaveBeenCalledWith("wl1", "BBCA");
  });

  it("shows empty panel when load fails", async () => {
    mockListWatchlists.mockRejectedValue(new Error("connection failed"));
    mockListIndexNames.mockResolvedValue([]);

    render(WatchlistPage);

    // When load errors, index names are empty and no watchlists are shown
    expect(await screen.findByText(/no indices available/i)).toBeInTheDocument();
    expect(screen.getByText(/select a watchlist or index/i)).toBeInTheDocument();
  });

  it("shows empty items state when items list is empty", async () => {
    mockListWatchlists.mockResolvedValue(mockWatchlists);
    mockListIndexNames.mockResolvedValue([]);
    mockGetWatchlistItems.mockResolvedValue([]);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: "My Banking" }));

    expect(await screen.findByText(/no items found/i)).toBeInTheDocument();
  });

  it("shows em dash for missing numeric data", async () => {
    mockListWatchlists.mockResolvedValue(mockWatchlists);
    mockListIndexNames.mockResolvedValue([]);
    mockGetWatchlistItems.mockResolvedValue([{ ticker: "BBRI", sector: "Banking" }]);

    render(WatchlistPage);
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: "My Banking" }));

    await screen.findByText("BBRI");
    const table = screen.getByRole("table", { name: /watchlist items/i });
    const dashes = within(table).getAllByText("\u2014");
    expect(dashes.length).toBeGreaterThanOrEqual(1);
  });
});
