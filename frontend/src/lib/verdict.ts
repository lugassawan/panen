import { t } from "../i18n";
import type { Verdict } from "./types";

export interface VerdictDisplay {
  label: string;
  colorClass: string;
  bgClass: string;
  icon: string;
  description: string;
}

const verdictStyles: Record<Verdict, { colorClass: string; bgClass: string; icon: string }> = {
  UNDERVALUED: {
    colorClass: "text-positive",
    bgClass: "bg-positive-bg border-positive/20",
    icon: "\u25B2",
  },
  FAIR: {
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

const verdictI18n: Record<Verdict, { labelKey: string; descKey: string }> = {
  UNDERVALUED: { labelKey: "verdict.undervalued", descKey: "verdict.undervaluedDesc" },
  FAIR: { labelKey: "verdict.fairValue", descKey: "verdict.fairValueDesc" },
  OVERVALUED: { labelKey: "verdict.overvalued", descKey: "verdict.overvaluedDesc" },
};

export function getVerdictDisplay(verdict: string): VerdictDisplay {
  const style = verdictStyles[verdict as Verdict];
  const i18n = verdictI18n[verdict as Verdict];
  if (style && i18n) {
    return { ...style, label: t(i18n.labelKey), description: t(i18n.descKey) };
  }
  return { ...fallbackStyle, label: t("verdict.unknown"), description: t("verdict.unknownDesc") };
}
