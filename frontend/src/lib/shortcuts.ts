import type { Page } from "./types";

const PAGE_ORDER: Page[] = [
  "lookup",
  "watchlist",
  "screener",
  "portfolio",
  "payday",
  "crashplaybook",
  "transactions",
  "alerts",
  "brokerage",
  "settings",
  "comparison",
];

export interface ShortcutHandlers {
  onNavigate: (page: Page) => void;
  onToggleCommandPalette: () => void;
  onToggleHelp: () => void;
  onAction?: (action: string) => void;
  currentPage?: Page;
}

export function handleGlobalShortcut(e: KeyboardEvent, handlers: ShortcutHandlers): void {
  const meta = e.metaKey || e.ctrlKey;

  // Input-focus guard: allow Cmd+K and Escape in inputs, block everything else
  const target = e.target as HTMLElement | null;
  const isInput =
    target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    target instanceof HTMLSelectElement ||
    target?.isContentEditable;
  if (isInput && !(meta && e.key === "k") && e.key !== "Escape") return;

  // Cmd+1 through Cmd+9 — page navigation
  if (meta) {
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
      return;
    }
  }

  // Shift+? — toggle help overlay
  if (e.shiftKey && e.key === "?") {
    e.preventDefault();
    handlers.onToggleHelp();
    return;
  }

  // "/" — open command palette (acts as search)
  if (e.key === "/") {
    e.preventDefault();
    handlers.onToggleCommandPalette();
    return;
  }

  // "n" — new holding (portfolio page only)
  if (e.key === "n" && handlers.currentPage === "portfolio") {
    e.preventDefault();
    handlers.onAction?.("newHolding");
  }
}
