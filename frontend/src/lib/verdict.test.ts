import { describe, expect, it } from "vitest";
import type { Verdict } from "./types";
import { getVerdictDisplay } from "./verdict";

describe("getVerdictDisplay", () => {
  it("maps UNDERVALUED to emerald with up arrow", () => {
    const display = getVerdictDisplay("UNDERVALUED");
    expect(display.label).toBe("Undervalued");
    expect(display.icon).toBe("\u25B2");
    expect(display.colorClass).toContain("emerald");
    expect(display.bgClass).toContain("emerald");
    expect(display.description).toBeTruthy();
  });

  it("maps FAIR to amber with diamond", () => {
    const display = getVerdictDisplay("FAIR");
    expect(display.label).toBe("Fair Value");
    expect(display.icon).toBe("\u25C6");
    expect(display.colorClass).toContain("amber");
    expect(display.bgClass).toContain("amber");
    expect(display.description).toBeTruthy();
  });

  it("maps OVERVALUED to red with down arrow", () => {
    const display = getVerdictDisplay("OVERVALUED");
    expect(display.label).toBe("Overvalued");
    expect(display.icon).toBe("\u25BC");
    expect(display.colorClass).toContain("red");
    expect(display.bgClass).toContain("red");
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
