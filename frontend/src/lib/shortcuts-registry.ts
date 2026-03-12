import type { Page } from "./types";

export type ShortcutCategory = "global" | "navigation" | "action";

export type ShortcutDef = {
  keys: string;
  label: string;
  category: ShortcutCategory;
  page?: Page;
};

export const SHORTCUT_REGISTRY: ShortcutDef[] = [
  // Global
  { keys: "⌘K", label: "shortcuts.commandPalette", category: "global" },
  { keys: "⇧?", label: "shortcuts.showHelp", category: "global" },
  { keys: "Esc", label: "shortcuts.closeOverlay", category: "global" },

  // Navigation (Cmd+1 through Cmd+9)
  { keys: "⌘1", label: "nav.lookup", category: "navigation" },
  { keys: "⌘2", label: "nav.watchlist", category: "navigation" },
  { keys: "⌘3", label: "nav.screener", category: "navigation" },
  { keys: "⌘4", label: "nav.portfolio", category: "navigation" },
  { keys: "⌘5", label: "nav.payday", category: "navigation" },
  { keys: "⌘6", label: "nav.crashPlaybook", category: "navigation" },
  { keys: "⌘7", label: "nav.transactions", category: "navigation" },
  { keys: "⌘8", label: "nav.alerts", category: "navigation" },
  { keys: "⌘9", label: "nav.brokerage", category: "navigation" },

  // Actions
  { keys: "/", label: "shortcuts.search", category: "action" },
];
