import type { Page } from "./types";

const PAGE_ORDER: Page[] = [
  "lookup",
  "watchlist",
  "screener",
  "portfolio",
  "payday",
  "crashplaybook",
  "alerts",
  "brokerage",
  "settings",
];

export interface ShortcutHandlers {
  onNavigate: (page: Page) => void;
  onToggleCommandPalette: () => void;
}

export function handleGlobalShortcut(e: KeyboardEvent, handlers: ShortcutHandlers): void {
  const meta = e.metaKey || e.ctrlKey;
  if (!meta) return;

  // Cmd+1 through Cmd+8 — page navigation
  const num = Number.parseInt(e.key, 10);
  if (num >= 1 && num <= PAGE_ORDER.length) {
    e.preventDefault();
    handlers.onNavigate(PAGE_ORDER[num - 1]);
    return;
  }

  // Cmd+K — command palette
  if (e.key === "k") {
    e.preventDefault();
    handlers.onToggleCommandPalette();
  }
}
