/**
 * Panen Theme Store — Svelte 5 (Runes)
 *
 * Manages light/dark/system theme preference.
 * Persists to localStorage, syncs with OS preference.
 *
 * Usage in components:
 *   import { theme } from "$lib/stores/theme.svelte";
 *
 *   <button onclick={() => theme.set('dark')}>Dark</button>
 *   <span>{theme.current}</span>    // 'light' | 'dark'
 *   <span>{theme.preference}</span> // 'light' | 'dark' | 'system'
 */

const browser = typeof window !== "undefined";

type ThemePreference = "light" | "dark" | "system";
type ResolvedTheme = "light" | "dark";

const STORAGE_KEY = "panen-theme";

function createThemeStore() {
  let preference = $state<ThemePreference>(loadPreference());
  let resolved = $state<ResolvedTheme>(resolve(preference));

  // Watch OS preference changes
  if (browser) {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    mq.addEventListener("change", () => {
      if (preference === "system") {
        resolved = resolve("system");
        applyTheme(resolved);
      }
    });

    // Apply on init
    applyTheme(resolved);
  }

  function loadPreference(): ThemePreference {
    if (!browser) return "system";
    return (localStorage.getItem(STORAGE_KEY) as ThemePreference) ?? "system";
  }

  function resolve(pref: ThemePreference): ResolvedTheme {
    if (pref === "system") {
      if (!browser) return "light";
      return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    }
    return pref;
  }

  function applyTheme(theme: ResolvedTheme) {
    if (!browser) return;
    const root = document.documentElement;
    root.classList.toggle("dark", theme === "dark");
    // Update meta theme-color for native title bar
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) {
      meta.setAttribute("content", theme === "dark" ? "#111112" : "#FEFCF7");
    }
  }

  return {
    /** The resolved theme: 'light' | 'dark' */
    get current(): ResolvedTheme {
      return resolved;
    },

    /** The user's preference: 'light' | 'dark' | 'system' */
    get preference(): ThemePreference {
      return preference;
    },

    /** Whether dark mode is active */
    get isDark(): boolean {
      return resolved === "dark";
    },

    /** Set theme preference */
    set(pref: ThemePreference) {
      preference = pref;
      resolved = resolve(pref);
      if (browser) {
        localStorage.setItem(STORAGE_KEY, pref);
        applyTheme(resolved);
      }
    },

    /** Cycle: light → dark → system → light */
    toggle() {
      const cycle: ThemePreference[] = ["light", "dark", "system"];
      const next = cycle[(cycle.indexOf(preference) + 1) % cycle.length];
      this.set(next);
    },
  };
}

export const theme = createThemeStore();
