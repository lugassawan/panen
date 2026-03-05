import { render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockAddToWatchlist = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  AddToWatchlist: (...args: unknown[]) => mockAddToWatchlist(...args),
}));

import WatchlistAddTicker from "./WatchlistAddTicker.svelte";

describe("WatchlistAddTicker", () => {
  beforeEach(() => {
    mockAddToWatchlist.mockReset();
  });

  it("renders input and add button", () => {
    render(WatchlistAddTicker, {
      props: { watchlistId: "w1", onAdded: vi.fn() },
    });
    expect(screen.getByLabelText("Add ticker to watchlist")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Add/i })).toBeInTheDocument();
  });

  it("disables submit when ticker is empty", () => {
    render(WatchlistAddTicker, {
      props: { watchlistId: "w1", onAdded: vi.fn() },
    });
    expect(screen.getByRole("button", { name: /Add/i })).toBeDisabled();
  });

  it("calls AddToWatchlist with uppercase ticker", async () => {
    const user = userEvent.setup();
    const onAdded = vi.fn();
    mockAddToWatchlist.mockResolvedValueOnce(undefined);

    render(WatchlistAddTicker, {
      props: { watchlistId: "w1", onAdded },
    });

    await user.type(screen.getByLabelText("Add ticker to watchlist"), "bbca");
    await user.click(screen.getByRole("button", { name: /Add/i }));

    await waitFor(() => {
      expect(mockAddToWatchlist).toHaveBeenCalledWith("w1", "BBCA");
    });
    expect(onAdded).toHaveBeenCalledOnce();
  });

  it("shows error on failure", async () => {
    const user = userEvent.setup();
    mockAddToWatchlist.mockRejectedValueOnce(new Error("ticker not found"));

    render(WatchlistAddTicker, {
      props: { watchlistId: "w1", onAdded: vi.fn() },
    });

    await user.type(screen.getByLabelText("Add ticker to watchlist"), "XXXX");
    await user.click(screen.getByRole("button", { name: /Add/i }));

    await waitFor(() => {
      expect(screen.getByText("ticker not found")).toBeInTheDocument();
    });
  });
});
