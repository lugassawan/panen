/**
 * Panen Locale Store — Svelte 5 (Runes)
 *
 * Manages EN/ID locale preference.
 * Persists to localStorage, detects system language.
 *
 * Usage:
 *   import { locale, t } from "../i18n";
 *
 *   <button onclick={() => locale.set('id')}>ID</button>
 *   <span>{t("nav.portfolio")}</span>
 */

import en from "./en.json";
import id from "./id.json";
import type { Locale, Translations } from "./types";

const browser = typeof window !== "undefined";
const STORAGE_KEY = "panen-locale";

const messages: Record<Locale, Translations> = { en, id };

export function detectLocale(): Locale {
  if (!browser) return "en";
  const lang = navigator.language;
  return lang.startsWith("id") ? "id" : "en";
}

function loadLocale(): Locale {
  if (!browser) return "en";
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "en" || stored === "id") return stored;
  } catch {
    // localStorage may be unavailable in test environments
  }
  return detectLocale();
}

function resolve(key: string, translations: Translations): string | undefined {
  const parts = key.split(".");
  let current: string | Translations = translations;
  for (const part of parts) {
    if (typeof current !== "object" || current === null) return undefined;
    current = (current as Translations)[part];
  }
  return typeof current === "string" ? current : undefined;
}

function applyLang(loc: Locale) {
  if (browser) {
    document.documentElement.lang = loc;
  }
}

function createLocaleStore() {
  const initial = loadLocale();
  let active = $state<Locale>(initial);
  applyLang(initial);

  return {
    get current(): Locale {
      return active;
    },

    set(loc: Locale) {
      active = loc;
      if (browser) {
        localStorage.setItem(STORAGE_KEY, loc);
      }
      applyLang(loc);
    },

    toggle() {
      this.set(active === "en" ? "id" : "en");
    },
  };
}

export const locale = createLocaleStore();

export function t(key: string, params?: Record<string, string | number>): string {
  const value = resolve(key, messages[locale.current]) ?? key;
  if (!params) return value;
  return value.replace(/\{(\w+)\}/g, (_, name) => String(params[name] ?? `{${name}}`));
}
