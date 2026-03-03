import type { Verdict } from "./types";

export interface VerdictDisplay {
  label: string;
  colorClass: string;
  bgClass: string;
  icon: string;
  description: string;
}

const verdictMap: Record<Verdict, VerdictDisplay> = {
  UNDERVALUED: {
    label: "Undervalued",
    colorClass: "text-positive",
    bgClass: "bg-positive-bg border-positive/20",
    icon: "\u25B2",
    description: "Trading below estimated intrinsic value",
  },
  FAIR: {
    label: "Fair Value",
    colorClass: "text-warning",
    bgClass: "bg-warning-bg border-warning/20",
    icon: "\u25C6",
    description: "Trading near estimated intrinsic value",
  },
  OVERVALUED: {
    label: "Overvalued",
    colorClass: "text-negative",
    bgClass: "bg-negative-bg border-negative/20",
    icon: "\u25BC",
    description: "Trading above estimated intrinsic value",
  },
};

const fallback: VerdictDisplay = {
  label: "Unknown",
  colorClass: "text-text-muted",
  bgClass: "bg-bg-tertiary border-border-default",
  icon: "?",
  description: "Verdict not recognized",
};

export function getVerdictDisplay(verdict: string): VerdictDisplay {
  return verdictMap[verdict as Verdict] ?? fallback;
}
