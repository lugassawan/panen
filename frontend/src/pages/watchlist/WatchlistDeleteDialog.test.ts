import { render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockDeleteWatchlist = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  DeleteWatchlist: (...args: unknown[]) => mockDeleteWatchlist(...args),
}));

import WatchlistDeleteDialog from "./WatchlistDeleteDialog.svelte";

describe("WatchlistDeleteDialog", () => {
  const watchlist = { id: "w1", name: "My Watchlist", createdAt: "", updatedAt: "" };

  beforeEach(() => {
    mockDeleteWatchlist.mockReset();
  });

  it("renders watchlist name in confirmation message", () => {
    render(WatchlistDeleteDialog, {
      props: { watchlist, onDeleted: vi.fn(), onCancel: vi.fn() },
    });
    expect(screen.getByText("Are you sure you want to delete My Watchlist?")).toBeInTheDocument();
    expect(screen.getByText("This action cannot be undone.")).toBeInTheDocument();
  });

  it("calls DeleteWatchlist and onDeleted on confirm", async () => {
    const user = userEvent.setup();
    const onDeleted = vi.fn();
    mockDeleteWatchlist.mockResolvedValueOnce(undefined);

    render(WatchlistDeleteDialog, {
      props: { watchlist, onDeleted, onCancel: vi.fn() },
    });

    await user.click(screen.getByRole("button", { name: /Delete/i }));

    await waitFor(() => {
      expect(mockDeleteWatchlist).toHaveBeenCalledWith("w1");
    });
    expect(onDeleted).toHaveBeenCalledOnce();
  });

  it("shows error on failure", async () => {
    const user = userEvent.setup();
    mockDeleteWatchlist.mockRejectedValueOnce(new Error("delete failed"));

    render(WatchlistDeleteDialog, {
      props: { watchlist, onDeleted: vi.fn(), onCancel: vi.fn() },
    });

    await user.click(screen.getByRole("button", { name: /Delete/i }));

    await waitFor(() => {
      expect(screen.getByText("delete failed")).toBeInTheDocument();
    });
  });
});
