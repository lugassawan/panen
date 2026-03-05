import type { ChartOptions } from "chart.js";
import { mode } from "./stores/mode.svelte";
import { theme } from "./stores/theme.svelte";

const browser = typeof window !== "undefined";

function getCSSVar(name: string): string {
  if (!browser) return "";
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

function resolveColors() {
  // Access reactive state to trigger re-resolution
  void theme.current;
  void mode.current;

  return {
    profit: getCSSVar("--fin-profit") || "#1b7d4e",
    loss: getCSSVar("--fin-loss") || "#c4342d",
    textPrimary: getCSSVar("--text-primary") || "#1a1a1a",
    textSecondary: getCSSVar("--text-secondary") || "#4b5060",
    textMuted: getCSSVar("--text-muted") || "#9ca3af",
    borderDefault: getCSSVar("--border-default") || "#e0dbd2",
    bgElevated: getCSSVar("--bg-elevated") || "#ffffff",
  };
}

export function chartColors() {
  return resolveColors();
}

const GREEN_HUES = [
  "#1b6b4a",
  "#2fa06b",
  "#5db88c",
  "#8fd4b2",
  "#c2e8d4",
  "#e6f5ec",
  "#228b5b",
  "#0f4a32",
  "#0a3524",
  "#f2faf5",
];

const GOLD_HUES = [
  "#d4a12a",
  "#e8c456",
  "#f0d87e",
  "#f7ecba",
  "#fbf4dc",
  "#b8891a",
  "#9e7614",
  "#fdf9ee",
  "#c99220",
  "#dbb445",
];

export function accentPalette(n: number): string[] {
  const hues = mode.current === "dividend" ? GOLD_HUES : GREEN_HUES;
  const result: string[] = [];
  for (let i = 0; i < n; i++) {
    result.push(hues[i % hues.length]);
  }
  return result;
}

export function defaultChartOptions(): ChartOptions {
  const colors = resolveColors();
  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: { duration: 200 },
    plugins: {
      legend: {
        labels: {
          color: colors.textSecondary,
          font: { family: "DM Sans, sans-serif", size: 12 },
        },
      },
      tooltip: {
        titleFont: { family: "DM Sans, sans-serif" },
        bodyFont: { family: "DM Mono, monospace" },
      },
    },
    scales: {
      x: {
        ticks: {
          color: colors.textMuted,
          font: { family: "DM Mono, monospace", size: 11 },
        },
        grid: { color: `${colors.borderDefault}40` },
      },
      y: {
        ticks: {
          color: colors.textMuted,
          font: { family: "DM Mono, monospace", size: 11 },
        },
        grid: { color: `${colors.borderDefault}40` },
      },
    },
  };
}
