import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { formatDecimal, formatPercent, formatRelativeTime, formatRupiah } from "./format";

describe("formatRupiah", () => {
  it("formats a typical stock price", () => {
    expect(formatRupiah(9250)).toBe("Rp\u00A09.250");
  });

  it("formats zero", () => {
    expect(formatRupiah(0)).toBe("Rp\u00A00");
  });

  it("formats large values with thousand separators", () => {
    expect(formatRupiah(1500000)).toBe("Rp\u00A01.500.000");
  });

  it("formats decimal values by rounding", () => {
    expect(formatRupiah(9250.75)).toBe("Rp\u00A09.251");
  });
});

describe("formatDecimal", () => {
  it("formats with default 2 digits", () => {
    expect(formatDecimal(1.5678)).toBe("1,57");
  });

  it("formats with custom digits", () => {
    expect(formatDecimal(12.3, 1)).toBe("12,3");
  });

  it("formats zero", () => {
    expect(formatDecimal(0)).toBe("0,00");
  });

  it("formats negative values", () => {
    expect(formatDecimal(-7.89, 2)).toBe("-7,89");
  });
});

describe("formatPercent", () => {
  it("formats a percentage value", () => {
    expect(formatPercent(25.5)).toBe("25,50%");
  });

  it("formats with custom digits", () => {
    expect(formatPercent(33.333, 1)).toBe("33,3%");
  });

  it("formats zero", () => {
    expect(formatPercent(0)).toBe("0,00%");
  });

  it("formats negative percentages", () => {
    expect(formatPercent(-12.5)).toBe("-12,50%");
  });
});

describe("formatRelativeTime", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-06-15T12:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns 'Not synced yet' for empty string", () => {
    expect(formatRelativeTime("")).toBe("Not synced yet");
  });

  it("returns 'just now' for less than 1 minute ago", () => {
    expect(formatRelativeTime("2025-06-15T11:59:30Z")).toBe("just now");
  });

  it("returns minutes ago for less than 1 hour", () => {
    expect(formatRelativeTime("2025-06-15T11:45:00Z")).toBe("15m ago");
  });

  it("returns hours ago for less than 24 hours", () => {
    expect(formatRelativeTime("2025-06-15T06:00:00Z")).toBe("6h ago");
  });

  it("returns days ago for 24+ hours", () => {
    expect(formatRelativeTime("2025-06-13T12:00:00Z")).toBe("2d ago");
  });
});
