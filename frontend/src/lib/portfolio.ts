import type { HoldingDetailResponse, HoldingWeight, SectorWeight } from "./types";

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

export function calcPLAbsolute(holding: HoldingDetailResponse): number | null {
  if (holding.currentPrice == null) return null;
  return (holding.currentPrice - holding.avgBuyPrice) * holding.lots * 100;
}

export function holdingWeights(holdings: HoldingDetailResponse[]): HoldingWeight[] {
  const total = currentValue(holdings);
  if (total <= 0) return [];
  return holdings.map((h) => {
    const value = (h.currentPrice ?? h.avgBuyPrice) * h.lots * 100;
    return { ticker: h.ticker, value, pct: (value / total) * 100 };
  });
}

export function sectorWeights(
  holdings: HoldingDetailResponse[],
  sectorMap: Record<string, string>,
): SectorWeight[] {
  const total = currentValue(holdings);
  if (total <= 0) return [];
  const groups: Record<string, number> = {};
  for (const h of holdings) {
    const sector = sectorMap[h.ticker] || "Unknown";
    const value = (h.currentPrice ?? h.avgBuyPrice) * h.lots * 100;
    groups[sector] = (groups[sector] ?? 0) + value;
  }
  return Object.entries(groups).map(([sector, value]) => ({
    sector,
    value,
    pct: (value / total) * 100,
  }));
}
