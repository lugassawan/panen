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

export type ThemePreference = "light" | "dark" | "system";
export type ResolvedTheme = "light" | "dark";

const STORAGE_KEY = "panen-theme";

function loadPreference(): ThemePreference {
  if (!browser) return "system";
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === "light" || stored === "dark" || stored === "system") return stored;
  return "system";
}

function resolve(pref: ThemePreference): ResolvedTheme {
  if (pref === "system") {
    if (!browser) return "light";
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
  return pref;
}

function applyTheme(resolvedTheme: ResolvedTheme) {
  if (!browser) return;
  const root = document.documentElement;
  root.classList.toggle("dark", resolvedTheme === "dark");
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) {
    meta.setAttribute("content", resolvedTheme === "dark" ? "#111112" : "#FEFCF7");
  }
}

function createThemeStore() {
  const initialPref = loadPreference();
  const initialResolved = resolve(initialPref);
  let preference = $state<ThemePreference>(initialPref);
  let resolved = $state<ResolvedTheme>(initialResolved);

  if (browser) {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    mq.addEventListener("change", () => {
      if (preference === "system") {
        const newResolved = resolve("system");
        resolved = newResolved;
        applyTheme(newResolved);
      }
    });

    applyTheme(initialResolved);
  }

  return {
    get current(): ResolvedTheme {
      return resolved;
    },

    get preference(): ThemePreference {
      return preference;
    },

    get isDark(): boolean {
      return resolved === "dark";
    },

    set(pref: ThemePreference) {
      preference = pref;
      const newResolved = resolve(pref);
      resolved = newResolved;
      if (browser) {
        localStorage.setItem(STORAGE_KEY, pref);
        applyTheme(newResolved);
      }
    },

    toggle() {
      const cycle: ThemePreference[] = ["light", "dark", "system"];
      const next = cycle[(cycle.indexOf(preference) + 1) % cycle.length];
      this.set(next);
    },
  };
}

export const theme = createThemeStore();
