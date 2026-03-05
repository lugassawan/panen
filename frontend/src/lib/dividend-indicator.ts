export type DividendIndicator = "BUY_ZONE" | "AVERAGE_UP" | "HOLD" | "OVERVALUED";

export interface DividendIndicatorDisplay {
  label: string;
  colorClass: string;
  bgClass: string;
  icon: string;
}

const indicatorMap: Record<DividendIndicator, DividendIndicatorDisplay> = {
  BUY_ZONE: {
    label: "Buy Zone",
    colorClass: "text-positive",
    bgClass: "bg-positive-bg border-positive/20",
    icon: "\u25B2",
  },
  AVERAGE_UP: {
    label: "Average Up",
    colorClass: "text-accent-blue",
    bgClass: "bg-accent-blue/10 border-accent-blue/20",
    icon: "\u25B2",
  },
  HOLD: {
    label: "Hold",
    colorClass: "text-warning",
    bgClass: "bg-warning-bg border-warning/20",
    icon: "\u25C6",
  },
  OVERVALUED: {
    label: "Overvalued",
    colorClass: "text-negative",
    bgClass: "bg-negative-bg border-negative/20",
    icon: "\u25BC",
  },
};

const fallback: DividendIndicatorDisplay = {
  label: "Unknown",
  colorClass: "text-text-muted",
  bgClass: "bg-bg-tertiary border-border-default",
  icon: "?",
};

export function getDividendIndicatorDisplay(indicator: string): DividendIndicatorDisplay {
  return indicatorMap[indicator as DividendIndicator] ?? fallback;
}
