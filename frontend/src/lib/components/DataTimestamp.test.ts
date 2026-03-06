import { render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("../../i18n", () => ({
  locale: { current: "en" },
  t: (key: string, params?: Record<string, string | number>) => {
    const keys: Record<string, string> = {
      "format.justNow": "just now",
      "format.minutesAgo": "{count}m ago",
      "format.lastUpdated": "Last updated",
    };
    let value = keys[key] ?? key;
    if (params) {
      value = value.replace(/\{(\w+)\}/g, (_, name) => String(params[name] ?? `{${name}}`));
    }
    return value;
  },
}));

import DataTimestampWrapper from "./__tests__/DataTimestampWrapper.svelte";

describe("DataTimestamp", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders with default label and relative time", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-01-01T12:05:00Z"));
    render(DataTimestampWrapper, { props: { date: "2025-01-01T12:00:00Z" } });
    expect(screen.getByText(/Last updated/)).toBeInTheDocument();
    expect(screen.getByText("5m ago")).toBeInTheDocument();
  });

  it("renders with custom label", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-01-01T12:05:00Z"));
    render(DataTimestampWrapper, {
      props: { date: "2025-01-01T12:00:00Z", label: "Fetched" },
    });
    expect(screen.getByText(/Fetched/)).toBeInTheDocument();
  });

  it("shows just now for recent timestamps", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-01-01T12:00:30Z"));
    render(DataTimestampWrapper, { props: { date: "2025-01-01T12:00:00Z" } });
    expect(screen.getByText("just now")).toBeInTheDocument();
  });

  it("renders a time element with datetime attribute", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-01-01T12:05:00Z"));
    render(DataTimestampWrapper, { props: { date: "2025-01-01T12:00:00Z" } });
    const timeEl = document.querySelector("time");
    expect(timeEl).toBeInTheDocument();
    expect(timeEl?.getAttribute("datetime")).toBe("2025-01-01T12:00:00Z");
  });
});
