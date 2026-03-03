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
  brokerCode: string;
  buyFeePct: number;
  sellFeePct: number;
  sellTaxPct: number;
  isManualFee: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface BrokerConfigResponse {
  code: string;
  name: string;
  buyFeePct: number;
  sellFeePct: number;
  sellTaxPct: number;
  notes: string;
}

export type Mode = "VALUE" | "DIVIDEND";

export interface PortfolioResponse {
  id: string;
  brokerageAcctId: string;
  name: string;
  mode: Mode;
  riskProfile: RiskProfile;
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

export type Page = "lookup" | "watchlist" | "portfolio" | "brokerage" | "settings";

export interface WatchlistResponse {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface WatchlistItemResponse {
  ticker: string;
  sector: string;
  price?: number;
  roe?: number;
  der?: number;
  eps?: number;
  dividendYield?: number;
  payoutRatio?: number;
  grahamNumber?: number;
  entryPrice?: number;
  exitTarget?: number;
  verdict?: string;
  fetchedAt?: string;
}

export interface RefreshProgress {
  ticker: string;
  index: number;
  total: number;
  status: "success" | "skipped" | "error";
  error?: string;
}

export interface RefreshSummary {
  total: number;
  fetched: number;
  skipped: number;
  failed: number;
  duration: string;
}

export interface RefreshStatus {
  state: "idle" | "syncing" | "error";
  lastRefresh: string;
  error?: string;
}

export interface RefreshSettingsResponse {
  autoRefreshEnabled: boolean;
  intervalMinutes: number;
  lastRefreshedAt: string;
}
