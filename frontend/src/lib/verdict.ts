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
    colorClass: "text-emerald-400",
    bgClass: "bg-emerald-500/10 border-emerald-500/30",
    icon: "\u25B2",
    description: "Trading below estimated intrinsic value",
  },
  FAIR: {
    label: "Fair Value",
    colorClass: "text-amber-400",
    bgClass: "bg-amber-500/10 border-amber-500/30",
    icon: "\u25C6",
    description: "Trading near estimated intrinsic value",
  },
  OVERVALUED: {
    label: "Overvalued",
    colorClass: "text-red-400",
    bgClass: "bg-red-500/10 border-red-500/30",
    icon: "\u25BC",
    description: "Trading above estimated intrinsic value",
  },
};

const fallback: VerdictDisplay = {
  label: "Unknown",
  colorClass: "text-neutral-400",
  bgClass: "bg-neutral-500/10 border-neutral-500/30",
  icon: "?",
  description: "Verdict not recognized",
};

export function getVerdictDisplay(verdict: string): VerdictDisplay {
  return verdictMap[verdict as Verdict] ?? fallback;
}
