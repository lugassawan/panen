import { render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockCreateWatchlist = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  CreateWatchlist: (...args: unknown[]) => mockCreateWatchlist(...args),
}));

import WatchlistCreateForm from "./WatchlistCreateForm.svelte";

describe("WatchlistCreateForm", () => {
  beforeEach(() => {
    mockCreateWatchlist.mockReset();
  });

  it("renders input and buttons", () => {
    render(WatchlistCreateForm, {
      props: { onCreated: vi.fn(), onCancel: vi.fn() },
    });
    expect(screen.getByLabelText("Watchlist name")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Add/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Cancel/i })).toBeInTheDocument();
  });

  it("disables submit when name is empty", () => {
    render(WatchlistCreateForm, {
      props: { onCreated: vi.fn(), onCancel: vi.fn() },
    });
    expect(screen.getByRole("button", { name: /Add/i })).toBeDisabled();
  });

  it("calls CreateWatchlist and onCreated on submit", async () => {
    const user = userEvent.setup();
    const onCreated = vi.fn();
    mockCreateWatchlist.mockResolvedValueOnce({
      id: "w1",
      name: "My List",
    });

    render(WatchlistCreateForm, {
      props: { onCreated, onCancel: vi.fn() },
    });

    await user.type(screen.getByLabelText("Watchlist name"), "My List");
    await user.click(screen.getByRole("button", { name: /Add/i }));

    await waitFor(() => {
      expect(mockCreateWatchlist).toHaveBeenCalledWith("My List");
    });
    expect(onCreated).toHaveBeenCalledOnce();
  });

  it("shows error on failure", async () => {
    const user = userEvent.setup();
    mockCreateWatchlist.mockRejectedValueOnce(new Error("name already taken"));

    render(WatchlistCreateForm, {
      props: { onCreated: vi.fn(), onCancel: vi.fn() },
    });

    await user.type(screen.getByLabelText("Watchlist name"), "Duplicate");
    await user.click(screen.getByRole("button", { name: /Add/i }));

    await waitFor(() => {
      expect(screen.getByText("name already taken")).toBeInTheDocument();
    });
  });

  it("calls onCancel when cancel clicked", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(WatchlistCreateForm, {
      props: { onCreated: vi.fn(), onCancel },
    });

    await user.click(screen.getByRole("button", { name: /Cancel/i }));
    expect(onCancel).toHaveBeenCalledOnce();
  });
});
