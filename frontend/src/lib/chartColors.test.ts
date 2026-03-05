import { describe, expect, it, vi } from "vitest";

vi.mock("./stores/theme.svelte", () => ({
  theme: { current: "light" },
}));

vi.mock("./stores/mode.svelte", () => ({
  mode: { current: "value" },
}));

import { accentPalette, chartColors, defaultChartOptions } from "./chartColors.svelte";

describe("accentPalette", () => {
  it("returns n colors", () => {
    const palette = accentPalette(5);
    expect(palette).toHaveLength(5);
    for (const c of palette) {
      expect(c).toMatch(/^#[0-9a-f]{6}$/i);
    }
  });

  it("wraps around when n exceeds available hues", () => {
    const palette = accentPalette(12);
    expect(palette).toHaveLength(12);
    expect(palette[10]).toBe(palette[0]);
  });

  it("returns empty array for n=0", () => {
    expect(accentPalette(0)).toEqual([]);
  });
});

describe("chartColors", () => {
  it("returns fallback colors in jsdom", () => {
    const colors = chartColors();
    expect(colors.profit).toBeTruthy();
    expect(colors.loss).toBeTruthy();
    expect(colors.textPrimary).toBeTruthy();
    expect(colors.textSecondary).toBeTruthy();
    expect(colors.textMuted).toBeTruthy();
    expect(colors.borderDefault).toBeTruthy();
    expect(colors.bgElevated).toBeTruthy();
  });
});

describe("defaultChartOptions", () => {
  it("returns responsive options with animation", () => {
    const opts = defaultChartOptions();
    expect(opts.responsive).toBe(true);
    expect(opts.maintainAspectRatio).toBe(false);
    expect(opts.animation).toEqual({ duration: 200 });
  });

  it("configures tooltip with DM Mono font", () => {
    const opts = defaultChartOptions();
    const tooltip = opts.plugins?.tooltip;
    expect(tooltip).toBeDefined();
    expect((tooltip as Record<string, unknown>).bodyFont).toEqual({
      family: "DM Mono, monospace",
    });
  });
});
