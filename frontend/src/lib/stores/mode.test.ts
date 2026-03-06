import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("../../i18n", () => ({
  t: (key: string) => {
    const translations: Record<string, string> = {
      "mode.value": "Value",
      "mode.dividend": "Dividend",
    };
    return translations[key] ?? key;
  },
}));

describe("mode store", () => {
  afterEach(() => {
    vi.resetModules();
  });

  async function loadMode() {
    const mod = await import("./mode.svelte");
    return mod.mode;
  }

  it("defaults to value mode", async () => {
    const mode = await loadMode();
    expect(mode.current).toBe("value");
    expect(mode.isValue).toBe(true);
    expect(mode.isDividend).toBe(false);
  });

  it("set changes mode", async () => {
    const mode = await loadMode();
    mode.set("dividend");
    expect(mode.current).toBe("dividend");
    expect(mode.isDividend).toBe(true);
    expect(mode.isValue).toBe(false);
  });

  it("toggle switches between modes", async () => {
    const mode = await loadMode();
    expect(mode.current).toBe("value");

    mode.toggle();
    expect(mode.current).toBe("dividend");

    mode.toggle();
    expect(mode.current).toBe("value");
  });

  it("returns correct containerClass for each mode", async () => {
    const mode = await loadMode();
    expect(mode.containerClass).toBe("mode-value");

    mode.set("dividend");
    expect(mode.containerClass).toBe("mode-dividend");
  });

  it("returns correct badgeClass for each mode", async () => {
    const mode = await loadMode();
    expect(mode.badgeClass).toBe("bg-green-100 text-green-700");

    mode.set("dividend");
    expect(mode.badgeClass).toBe("bg-gold-100 text-gold-700");
  });

  it("returns correct accentColor for each mode", async () => {
    const mode = await loadMode();
    expect(mode.accentColor).toBe("var(--color-green-700)");

    mode.set("dividend");
    expect(mode.accentColor).toBe("var(--color-gold-500)");
  });

  it("provides full config object", async () => {
    const mode = await loadMode();
    const config = mode.config;
    expect(config.label).toBe("Value");
    expect(config.accent).toBe("var(--color-green-700)");

    mode.set("dividend");
    const dConfig = mode.config;
    expect(dConfig.label).toBe("Dividend");
    expect(dConfig.accent).toBe("var(--color-gold-500)");
  });
});
