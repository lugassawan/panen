import { t } from "../i18n";

export type DividendIndicator = "BUY_ZONE" | "AVERAGE_UP" | "HOLD" | "OVERVALUED";

export interface DividendIndicatorDisplay {
  label: string;
  colorClass: string;
  bgClass: string;
  icon: string;
}

const indicatorStyles: Record<
  DividendIndicator,
  { colorClass: string; bgClass: string; icon: string }
> = {
  BUY_ZONE: {
    colorClass: "text-positive",
    bgClass: "bg-positive-bg border-positive/20",
    icon: "\u25B2",
  },
  AVERAGE_UP: {
    colorClass: "text-info",
    bgClass: "bg-info-bg border-info/20",
    icon: "\u25B2",
  },
  HOLD: {
    colorClass: "text-warning",
    bgClass: "bg-warning-bg border-warning/20",
    icon: "\u25C6",
  },
  OVERVALUED: {
    colorClass: "text-negative",
    bgClass: "bg-negative-bg border-negative/20",
    icon: "\u25BC",
  },
};

const fallbackStyle = {
  colorClass: "text-text-muted",
  bgClass: "bg-bg-tertiary border-border-default",
  icon: "?",
};

const indicatorKeys: Record<DividendIndicator, string> = {
  BUY_ZONE: "indicator.buyZone",
  AVERAGE_UP: "indicator.averageUp",
  HOLD: "indicator.hold",
  OVERVALUED: "indicator.overvalued",
};

export function getDividendIndicatorDisplay(indicator: string): DividendIndicatorDisplay {
  const style = indicatorStyles[indicator as DividendIndicator];
  const labelKey = indicatorKeys[indicator as DividendIndicator];
  if (style && labelKey) {
    return { ...style, label: t(labelKey) };
  }
  return { ...fallbackStyle, label: t("indicator.unknown") };
}
