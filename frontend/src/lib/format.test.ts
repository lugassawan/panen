import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

let mockLocale = "en";

vi.mock("../i18n", () => ({
  locale: {
    get current() {
      return mockLocale;
    },
  },
  t: (key: string, params?: Record<string, string | number>) => {
    const en: Record<string, string> = {
      "format.justNow": "just now",
      "format.minutesAgo": "{count}m ago",
      "format.hoursAgo": "{count}h ago",
      "format.daysAgo": "{count}d ago",
      "format.notSynced": "Not synced yet",
    };
    const id: Record<string, string> = {
      "format.justNow": "baru saja",
      "format.minutesAgo": "{count} menit lalu",
      "format.hoursAgo": "{count} jam lalu",
      "format.daysAgo": "{count} hari lalu",
      "format.notSynced": "Belum disinkronkan",
    };
    const translations = mockLocale === "id" ? id : en;
    let value = translations[key] ?? key;
    if (params) {
      value = value.replace(/\{(\w+)\}/g, (_, name) => String(params[name] ?? `{${name}}`));
    }
    return value;
  },
}));

import {
  formatDate,
  formatDecimal,
  formatFileSize,
  formatPercent,
  formatRelativeTime,
  formatRupiah,
} from "./format";

describe("formatRupiah", () => {
  beforeEach(() => {
    mockLocale = "en";
  });

  it("formats with EN locale", () => {
    mockLocale = "en";
    expect(formatRupiah(9250)).toBe("IDR\u00A09,250");
  });

  it("formats with ID locale", () => {
    mockLocale = "id";
    expect(formatRupiah(9250)).toBe("Rp\u00A09.250");
  });

  it("formats zero", () => {
    mockLocale = "id";
    expect(formatRupiah(0)).toBe("Rp\u00A00");
  });

  it("formats large values", () => {
    mockLocale = "id";
    expect(formatRupiah(1500000)).toBe("Rp\u00A01.500.000");
  });
});

describe("formatDecimal", () => {
  it("formats with EN locale", () => {
    mockLocale = "en";
    expect(formatDecimal(1.5678)).toBe("1.57");
  });

  it("formats with ID locale", () => {
    mockLocale = "id";
    expect(formatDecimal(1.5678)).toBe("1,57");
  });

  it("formats with custom digits", () => {
    mockLocale = "en";
    expect(formatDecimal(12.3, 1)).toBe("12.3");
  });

  it("formats zero", () => {
    mockLocale = "en";
    expect(formatDecimal(0)).toBe("0.00");
  });

  it("formats negative values", () => {
    mockLocale = "id";
    expect(formatDecimal(-7.89, 2)).toBe("-7,89");
  });
});

describe("formatPercent", () => {
  it("formats with EN locale", () => {
    mockLocale = "en";
    expect(formatPercent(25.5)).toBe("25.50%");
  });

  it("formats with ID locale", () => {
    mockLocale = "id";
    expect(formatPercent(25.5)).toBe("25,50%");
  });

  it("formats zero", () => {
    mockLocale = "en";
    expect(formatPercent(0)).toBe("0.00%");
  });

  it("formats negative percentages", () => {
    mockLocale = "id";
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

  it("returns translated 'Not synced yet' for empty string", () => {
    mockLocale = "en";
    expect(formatRelativeTime("")).toBe("Not synced yet");
  });

  it("returns translated 'just now' for less than 1 minute ago", () => {
    mockLocale = "en";
    expect(formatRelativeTime("2025-06-15T11:59:30Z")).toBe("just now");
  });

  it("returns minutes ago for less than 1 hour", () => {
    mockLocale = "en";
    expect(formatRelativeTime("2025-06-15T11:45:00Z")).toBe("15m ago");
  });

  it("returns hours ago for less than 24 hours", () => {
    mockLocale = "en";
    expect(formatRelativeTime("2025-06-15T06:00:00Z")).toBe("6h ago");
  });

  it("returns days ago for 24+ hours", () => {
    mockLocale = "en";
    expect(formatRelativeTime("2025-06-13T12:00:00Z")).toBe("2d ago");
  });

  it("returns Indonesian translations when locale is id", () => {
    mockLocale = "id";
    expect(formatRelativeTime("")).toBe("Belum disinkronkan");
    expect(formatRelativeTime("2025-06-15T11:59:30Z")).toBe("baru saja");
    expect(formatRelativeTime("2025-06-15T11:45:00Z")).toBe("15 menit lalu");
    expect(formatRelativeTime("2025-06-15T06:00:00Z")).toBe("6 jam lalu");
    expect(formatRelativeTime("2025-06-13T12:00:00Z")).toBe("2 hari lalu");
  });
});

describe("formatFileSize", () => {
  it("formats bytes", () => {
    expect(formatFileSize(500)).toBe("500 B");
  });

  it("formats kilobytes", () => {
    expect(formatFileSize(1536)).toBe("1.5 KB");
  });

  it("formats megabytes", () => {
    expect(formatFileSize(2621440)).toBe("2.5 MB");
  });

  it("formats zero", () => {
    expect(formatFileSize(0)).toBe("0 B");
  });
});

describe("formatDate", () => {
  it("formats date in EN locale", () => {
    mockLocale = "en";
    const result = formatDate("2026-03-01T00:00:00Z");
    expect(result).toContain("March");
    expect(result).toContain("2026");
  });

  it("formats date in ID locale", () => {
    mockLocale = "id";
    const result = formatDate("2026-03-01T00:00:00Z");
    expect(result).toContain("Maret");
    expect(result).toContain("2026");
  });

  it("returns empty string for empty input", () => {
    expect(formatDate("")).toBe("");
  });
});
