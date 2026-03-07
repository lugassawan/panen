import { formatDecimal, formatPercent, formatRupiah } from "../../lib/format";
import type { StockValuationResponse } from "../../lib/types";

export interface MetricConfig {
  key: keyof StockValuationResponse;
  labelKey: string;
  format: (v: number) => string;
  direction: "higher" | "lower";
  section: "valuation" | "fundamental";
}

export const VALUATION_METRICS: MetricConfig[] = [
  {
    key: "grahamNumber",
    labelKey: "lookup.grahamNumber",
    format: formatRupiah,
    direction: "higher",
    section: "valuation",
  },
  {
    key: "marginOfSafety",
    labelKey: "lookup.marginOfSafety",
    format: formatPercent,
    direction: "higher",
    section: "valuation",
  },
  {
    key: "entryPrice",
    labelKey: "lookup.entryPrice",
    format: formatRupiah,
    direction: "higher",
    section: "valuation",
  },
  {
    key: "exitTarget",
    labelKey: "lookup.exitTarget",
    format: formatRupiah,
    direction: "higher",
    section: "valuation",
  },
];

export const FUNDAMENTAL_METRICS: MetricConfig[] = [
  {
    key: "eps",
    labelKey: "EPS",
    format: formatRupiah,
    direction: "higher",
    section: "fundamental",
  },
  {
    key: "bvps",
    labelKey: "BVPS",
    format: formatRupiah,
    direction: "higher",
    section: "fundamental",
  },
  {
    key: "roe",
    labelKey: "ROE",
    format: formatPercent,
    direction: "higher",
    section: "fundamental",
  },
  {
    key: "der",
    labelKey: "DER",
    format: formatDecimal,
    direction: "lower",
    section: "fundamental",
  },
  {
    key: "pbv",
    labelKey: "PBV",
    format: formatDecimal,
    direction: "lower",
    section: "fundamental",
  },
  {
    key: "per",
    labelKey: "PER",
    format: formatDecimal,
    direction: "lower",
    section: "fundamental",
  },
  {
    key: "dividendYield",
    labelKey: "lookup.dividendYield",
    format: formatPercent,
    direction: "higher",
    section: "fundamental",
  },
  {
    key: "payoutRatio",
    labelKey: "lookup.payoutRatio",
    format: formatPercent,
    direction: "higher",
    section: "fundamental",
  },
];

export function findBestIndex(
  values: (number | null)[],
  direction: "higher" | "lower",
): number | null {
  let bestIdx: number | null = null;
  let bestVal: number | null = null;
  for (let i = 0; i < values.length; i++) {
    const v = values[i];
    if (v === null || v === undefined) continue;
    if (bestVal === null) {
      bestIdx = i;
      bestVal = v;
    } else if (direction === "higher" ? v > bestVal : v < bestVal) {
      bestIdx = i;
      bestVal = v;
    }
  }
  return bestIdx;
}
