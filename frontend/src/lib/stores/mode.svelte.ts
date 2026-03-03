/**
 * Panen Mode Store — Svelte 5 (Runes)
 *
 * Manages the active investment mode (Value vs Dividend).
 * Mode determines accent color, visible metrics, and UI behavior.
 *
 * Usage:
 *   import { mode } from "$lib/stores/mode.svelte";
 *
 *   <div class={mode.containerClass}>...</div>
 *   <span style:color={mode.accentColor}>Active</span>
 */

type InvestmentMode = "value" | "dividend";

interface ModeConfig {
  label: string;
  emoji: string;
  accent: string;
  accentLight: string;
  badgeClass: string;
  containerClass: string;
}

const MODE_CONFIG: Record<InvestmentMode, ModeConfig> = {
  value: {
    label: "Value",
    emoji: "\u{1F4C8}",
    accent: "var(--color-green-700)",
    accentLight: "var(--color-green-100)",
    badgeClass: "bg-green-100 text-green-700",
    containerClass: "mode-value",
  },
  dividend: {
    label: "Dividend",
    emoji: "\u{1F4B0}",
    accent: "var(--color-gold-500)",
    accentLight: "var(--color-gold-100)",
    badgeClass: "bg-gold-100 text-gold-700",
    containerClass: "mode-dividend",
  },
};

function createModeStore() {
  let active = $state<InvestmentMode>("value");

  return {
    /** Current mode: 'value' | 'dividend' */
    get current(): InvestmentMode {
      return active;
    },

    /** Full config for the active mode */
    get config(): ModeConfig {
      return MODE_CONFIG[active];
    },

    /** Whether value mode is active */
    get isValue(): boolean {
      return active === "value";
    },

    /** Whether dividend mode is active */
    get isDividend(): boolean {
      return active === "dividend";
    },

    /** CSS accent color for current mode */
    get accentColor(): string {
      return MODE_CONFIG[active].accent;
    },

    /** Tailwind class for container (applies mode CSS vars) */
    get containerClass(): string {
      return MODE_CONFIG[active].containerClass;
    },

    /** Tailwind badge classes for current mode */
    get badgeClass(): string {
      return MODE_CONFIG[active].badgeClass;
    },

    /** Switch mode */
    set(m: InvestmentMode) {
      active = m;
    },

    /** Toggle between modes */
    toggle() {
      active = active === "value" ? "dividend" : "value";
    },
  };
}

export const mode = createModeStore();
