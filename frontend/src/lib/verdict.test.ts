import { describe, expect, it } from "vitest";
import type { Verdict } from "./types";
import { getVerdictDisplay } from "./verdict";

describe("getVerdictDisplay", () => {
  it("maps UNDERVALUED to positive with up arrow", () => {
    const display = getVerdictDisplay("UNDERVALUED");
    expect(display.label).toBe("Undervalued");
    expect(display.icon).toBe("\u25B2");
    expect(display.colorClass).toContain("positive");
    expect(display.bgClass).toContain("positive");
    expect(display.description).toBeTruthy();
  });

  it("maps FAIR to warning with diamond", () => {
    const display = getVerdictDisplay("FAIR");
    expect(display.label).toBe("Fair Value");
    expect(display.icon).toBe("\u25C6");
    expect(display.colorClass).toContain("warning");
    expect(display.bgClass).toContain("warning");
    expect(display.description).toBeTruthy();
  });

  it("maps OVERVALUED to negative with down arrow", () => {
    const display = getVerdictDisplay("OVERVALUED");
    expect(display.label).toBe("Overvalued");
    expect(display.icon).toBe("\u25BC");
    expect(display.colorClass).toContain("negative");
    expect(display.bgClass).toContain("negative");
    expect(display.description).toBeTruthy();
  });

  it("returns distinct labels for each verdict", () => {
    const verdicts: Verdict[] = ["UNDERVALUED", "FAIR", "OVERVALUED"];
    const labels = verdicts.map((v) => getVerdictDisplay(v).label);
    expect(new Set(labels).size).toBe(3);
  });

  it("returns distinct icons for each verdict (colorblind accessible)", () => {
    const verdicts: Verdict[] = ["UNDERVALUED", "FAIR", "OVERVALUED"];
    const icons = verdicts.map((v) => getVerdictDisplay(v).icon);
    expect(new Set(icons).size).toBe(3);
  });
});
