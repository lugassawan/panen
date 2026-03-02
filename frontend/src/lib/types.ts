export interface BandStats {
  min: number;
  max: number;
  avg: number;
  median: number;
}

export interface StockValuationResponse {
  ticker: string;
  price: number;
  high52Week: number;
  low52Week: number;
  eps: number;
  bvps: number;
  roe: number;
  der: number;
  pbv: number;
  per: number;
  dividendYield: number;
  payoutRatio: number;
  grahamNumber: number;
  marginOfSafety: number;
  entryPrice: number;
  exitTarget: number;
  pbvBand?: BandStats;
  perBand?: BandStats;
  verdict: Verdict;
  riskProfile: RiskProfile;
  fetchedAt: string;
  source: string;
}

export type RiskProfile = "CONSERVATIVE" | "MODERATE" | "AGGRESSIVE";

export type Verdict = "UNDERVALUED" | "FAIR" | "OVERVALUED";

export type Page = "lookup" | "portfolio" | "settings";
