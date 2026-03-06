/**
 * Shared style maps for backend Mode enum strings ("VALUE", "DIVIDEND").
 *
 * These map uppercase backend enum values to Tailwind classes.
 * Distinct from the mode store (stores/mode.svelte.ts) which uses lowercase
 * frontend IDs and contains similar values in MODE_CONFIG.badgeClass.
 */

type BackendMode = "VALUE" | "DIVIDEND";

export const MODE_BADGE: Record<BackendMode, string> = {
  VALUE: "bg-green-100 text-green-700",
  DIVIDEND: "bg-gold-100 text-gold-700",
};

export const TAB_ACCENT: Record<BackendMode, string> = {
  VALUE: "border-green-700 text-green-700",
  DIVIDEND: "border-gold-500 text-gold-500",
};
