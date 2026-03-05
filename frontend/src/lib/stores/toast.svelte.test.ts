import { beforeEach, describe, expect, it, vi } from "vitest";
import { toastStore } from "./toast.svelte";

describe("toastStore", () => {
  beforeEach(() => {
    toastStore.clear();
  });

  it("starts with no toasts", () => {
    expect(toastStore.toasts).toEqual([]);
  });

  it("adds a toast", () => {
    toastStore.add("Success!", "success");
    expect(toastStore.toasts).toHaveLength(1);
    expect(toastStore.toasts[0].message).toBe("Success!");
    expect(toastStore.toasts[0].variant).toBe("success");
  });

  it("adds multiple toasts", () => {
    toastStore.add("First", "info");
    toastStore.add("Second", "error");
    expect(toastStore.toasts).toHaveLength(2);
  });

  it("limits to max 3 visible toasts", () => {
    toastStore.add("One", "info");
    toastStore.add("Two", "info");
    toastStore.add("Three", "info");
    toastStore.add("Four", "info");
    expect(toastStore.toasts).toHaveLength(3);
    expect(toastStore.toasts[0].message).toBe("Two");
  });

  it("dismisses a toast by id", () => {
    toastStore.add("Remove me", "warning");
    const id = toastStore.toasts[0].id;
    toastStore.dismiss(id);
    expect(toastStore.toasts).toHaveLength(0);
  });

  it("auto-dismisses after duration", () => {
    vi.useFakeTimers();
    toastStore.add("Auto", "success", 2000);
    expect(toastStore.toasts).toHaveLength(1);

    vi.advanceTimersByTime(2000);
    expect(toastStore.toasts).toHaveLength(0);
    vi.useRealTimers();
  });

  it("uses default 4s duration", () => {
    vi.useFakeTimers();
    toastStore.add("Default", "info");
    expect(toastStore.toasts).toHaveLength(1);

    vi.advanceTimersByTime(3999);
    expect(toastStore.toasts).toHaveLength(1);

    vi.advanceTimersByTime(1);
    expect(toastStore.toasts).toHaveLength(0);
    vi.useRealTimers();
  });

  it("clears all toasts", () => {
    toastStore.add("One", "info");
    toastStore.add("Two", "error");
    toastStore.clear();
    expect(toastStore.toasts).toEqual([]);
  });
});
