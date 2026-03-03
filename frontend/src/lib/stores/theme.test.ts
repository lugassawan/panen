import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// Mock matchMedia before importing theme store
const mockMatchMedia = vi.fn().mockReturnValue({
  matches: false,
  addEventListener: vi.fn(),
});
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: mockMatchMedia,
});

// Create a proper localStorage mock
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

describe("theme store", () => {
  beforeEach(() => {
    localStorageMock.clear();
    document.documentElement.classList.remove("dark");
    mockMatchMedia.mockReturnValue({
      matches: false,
      addEventListener: vi.fn(),
    });
  });

  afterEach(() => {
    vi.resetModules();
  });

  async function loadTheme() {
    const mod = await import("./theme.svelte");
    return mod.theme;
  }

  it("defaults to system preference when no localStorage", async () => {
    const theme = await loadTheme();
    expect(theme.preference).toBe("system");
  });

  it("resolves system preference to light when OS is light", async () => {
    const theme = await loadTheme();
    expect(theme.current).toBe("light");
    expect(theme.isDark).toBe(false);
  });

  it("resolves system preference to dark when OS is dark", async () => {
    mockMatchMedia.mockReturnValue({
      matches: true,
      addEventListener: vi.fn(),
    });
    const theme = await loadTheme();
    expect(theme.current).toBe("dark");
    expect(theme.isDark).toBe(true);
  });

  it("persists preference to localStorage on set", async () => {
    const theme = await loadTheme();
    theme.set("dark");
    expect(localStorageMock.getItem("panen-theme")).toBe("dark");
  });

  it("loads preference from localStorage", async () => {
    localStorageMock.setItem("panen-theme", "dark");
    const theme = await loadTheme();
    expect(theme.preference).toBe("dark");
    expect(theme.current).toBe("dark");
  });

  it("set changes current and preference", async () => {
    const theme = await loadTheme();
    theme.set("dark");
    expect(theme.preference).toBe("dark");
    expect(theme.current).toBe("dark");
    expect(theme.isDark).toBe(true);
  });

  it("toggle cycles light → dark → system", async () => {
    const theme = await loadTheme();
    theme.set("light");
    expect(theme.preference).toBe("light");

    theme.toggle();
    expect(theme.preference).toBe("dark");

    theme.toggle();
    expect(theme.preference).toBe("system");

    theme.toggle();
    expect(theme.preference).toBe("light");
  });

  it("toggles dark class on documentElement", async () => {
    const theme = await loadTheme();
    theme.set("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);

    theme.set("light");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
  });
});
