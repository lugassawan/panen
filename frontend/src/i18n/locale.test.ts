import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
    get length() {
      return Object.keys(store).length;
    },
    key: (index: number) => Object.keys(store)[index] ?? null,
  };
})();
Object.defineProperty(window, "localStorage", {
  writable: true,
  value: localStorageMock,
});

describe("locale store", () => {
  beforeEach(() => {
    localStorageMock.clear();
    Object.defineProperty(navigator, "language", {
      writable: true,
      value: "en-US",
    });
  });

  afterEach(() => {
    vi.resetModules();
  });

  async function loadLocale() {
    const mod = await import("./locale.svelte");
    return { locale: mod.locale, t: mod.t, detectLocale: mod.detectLocale };
  }

  it("defaults to system-detected locale", async () => {
    const { locale } = await loadLocale();
    expect(locale.current).toBe("en");
  });

  it("detects 'id' from navigator.language = 'id-ID'", async () => {
    Object.defineProperty(navigator, "language", { writable: true, value: "id-ID" });
    const { detectLocale } = await loadLocale();
    expect(detectLocale()).toBe("id");
  });

  it("falls back to 'en' for unsupported languages", async () => {
    Object.defineProperty(navigator, "language", { writable: true, value: "ja-JP" });
    const { detectLocale } = await loadLocale();
    expect(detectLocale()).toBe("en");
  });

  it("loads locale from localStorage", async () => {
    localStorageMock.setItem("panen-locale", "id");
    const { locale } = await loadLocale();
    expect(locale.current).toBe("id");
  });

  it("persists locale to localStorage on set", async () => {
    const { locale } = await loadLocale();
    locale.set("id");
    expect(localStorageMock.getItem("panen-locale")).toBe("id");
  });

  it("toggle switches between en and id", async () => {
    const { locale } = await loadLocale();
    expect(locale.current).toBe("en");
    locale.toggle();
    expect(locale.current).toBe("id");
    locale.toggle();
    expect(locale.current).toBe("en");
  });

  it("t() resolves nested dot-notation keys", async () => {
    const { t } = await loadLocale();
    expect(t("common.save")).toBe("Save");
    expect(t("nav.portfolio")).toBe("Portfolio");
  });

  it("t() interpolates {placeholder} params", async () => {
    const { t } = await loadLocale();
    expect(t("format.minutesAgo", { count: 5 })).toBe("5m ago");
    expect(t("settings.updateAvailable", { version: "1.2.0" })).toBe("Panen 1.2.0 is available.");
  });

  it("t() returns key string for missing keys", async () => {
    const { t } = await loadLocale();
    expect(t("does.not.exist")).toBe("does.not.exist");
  });

  it("t() returns translated text for Indonesian locale", async () => {
    localStorageMock.setItem("panen-locale", "id");
    const { t } = await loadLocale();
    expect(t("common.save")).toBe("Simpan");
    expect(t("nav.portfolio")).toBe("Portofolio");
  });

  it("t() interpolates params in Indonesian locale", async () => {
    localStorageMock.setItem("panen-locale", "id");
    const { t } = await loadLocale();
    expect(t("format.minutesAgo", { count: 5 })).toBe("5 menit lalu");
  });
});
