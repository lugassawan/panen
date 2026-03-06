import { describe, expect, it, vi } from "vitest";

vi.mock("../i18n", () => ({
  t: (key: string) => {
    const translations: Record<string, string> = {
      "indicator.buyZone": "Buy Zone",
      "indicator.averageUp": "Average Up",
      "indicator.hold": "Hold",
      "indicator.overvalued": "Overvalued",
      "indicator.unknown": "Unknown",
    };
    return translations[key] ?? key;
  },
}));

import { getDividendIndicatorDisplay } from "./dividend-indicator";

describe("getDividendIndicatorDisplay", () => {
  it("returns Buy Zone for BUY_ZONE", () => {
    const display = getDividendIndicatorDisplay("BUY_ZONE");
    expect(display.label).toBe("Buy Zone");
    expect(display.colorClass).toContain("positive");
  });

  it("returns Average Up for AVERAGE_UP", () => {
    const display = getDividendIndicatorDisplay("AVERAGE_UP");
    expect(display.label).toBe("Average Up");
    expect(display.colorClass).toContain("info");
  });

  it("returns Hold for HOLD", () => {
    const display = getDividendIndicatorDisplay("HOLD");
    expect(display.label).toBe("Hold");
    expect(display.colorClass).toContain("warning");
  });

  it("returns Overvalued for OVERVALUED", () => {
    const display = getDividendIndicatorDisplay("OVERVALUED");
    expect(display.label).toBe("Overvalued");
    expect(display.colorClass).toContain("negative");
  });

  it("returns fallback for unknown indicator", () => {
    const display = getDividendIndicatorDisplay("UNKNOWN");
    expect(display.label).toBe("Unknown");
    expect(display.colorClass).toContain("muted");
  });
});
