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

export interface BrokerageAccountResponse {
  id: string;
  brokerName: string;
  buyFeePct: number;
  sellFeePct: number;
  isManualFee: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface PortfolioResponse {
  id: string;
  brokerageAcctId: string;
  name: string;
  mode: string;
  riskProfile: string;
  capital: number;
  monthlyAddition: number;
  maxStocks: number;
  createdAt: string;
  updatedAt: string;
}

export interface HoldingDetailResponse {
  id: string;
  ticker: string;
  avgBuyPrice: number;
  lots: number;
  currentPrice?: number;
  grahamNumber?: number;
  entryPrice?: number;
  exitTarget?: number;
  verdict?: string;
  marginOfSafety?: number;
}

export interface PortfolioDetailResponse {
  portfolio: PortfolioResponse;
  holdings: HoldingDetailResponse[];
}

export type RiskProfile = "CONSERVATIVE" | "MODERATE" | "AGGRESSIVE";

export type Verdict = "UNDERVALUED" | "FAIR" | "OVERVALUED";

export type Page = "lookup" | "portfolio" | "settings";
