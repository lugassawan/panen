import type { HoldingDetailResponse } from "./types";

export function calcPL(currentPrice: number | undefined, avgBuyPrice: number): number | null {
  if (currentPrice == null) return null;
  return ((currentPrice - avgBuyPrice) / avgBuyPrice) * 100;
}

export function totalInvested(holdings: HoldingDetailResponse[]): number {
  return holdings.reduce((sum, h) => sum + h.avgBuyPrice * h.lots * 100, 0);
}

export function currentValue(holdings: HoldingDetailResponse[]): number {
  return holdings.reduce((sum, h) => sum + (h.currentPrice ?? h.avgBuyPrice) * h.lots * 100, 0);
}

export function overallPL(holdings: HoldingDetailResponse[]): number {
  const invested = totalInvested(holdings);
  if (invested <= 0) return 0;
  const current = currentValue(holdings);
  return ((current - invested) / invested) * 100;
}
