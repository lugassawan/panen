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

export interface FundamentalExitResponse {
  key: string;
  label: string;
  detail: string;
  triggered: boolean;
}

export interface TrailingStopResponse {
  peakPrice: number;
  stopPercentage: number;
  stopPrice: number;
  triggered: boolean;
  fundamentalExits: FundamentalExitResponse[];
}

export interface DividendMetricsResponse {
  indicator: string;
  annualDPS: number;
  yieldOnCost: number;
  projectedYoC: number;
  portfolioYield: number;
}

export interface DividendRankItemResponse {
  ticker: string;
  indicator: string;
  dividendYield: number;
  yieldOnCost: number;
  payoutRatio: number;
  positionPct: number;
  score: number;
  isHolding: boolean;
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
  trailingStop?: TrailingStopResponse;
  dividendMetrics?: DividendMetricsResponse;
}

export interface PortfolioDetailResponse {
  portfolio: PortfolioResponse;
  holdings: HoldingDetailResponse[];
}

export type RiskProfile = "CONSERVATIVE" | "MODERATE" | "AGGRESSIVE";

export type Verdict = "UNDERVALUED" | "FAIR" | "OVERVALUED";

export type Page =
  | "dashboard"
  | "lookup"
  | "watchlist"
  | "screener"
  | "comparison"
  | "portfolio"
  | "payday"
  | "crashplaybook"
  | "transactions"
  | "alerts"
  | "brokerage"
  | "settings";

export type PaydayStatus = "SCHEDULED" | "PENDING" | "CONFIRMED" | "DEFERRED" | "SKIPPED";

export interface MonthlyPaydayResponse {
  month: string;
  paydayDay: number;
  portfolios: PortfolioPaydayItemResponse[];
  totalExpected: number;
}

export interface PortfolioPaydayItemResponse {
  portfolioId: string;
  portfolioName: string;
  mode: Mode;
  expected: number;
  actual: number;
  status: PaydayStatus;
  deferUntil?: string;
}

export interface CashFlowSummaryResponse {
  items: CashFlowItemResponse[];
  totalInflow: number;
  totalDeployed: number;
  balance: number;
}

export interface CashFlowItemResponse {
  id: string;
  portfolioId: string;
  type: string;
  amount: number;
  date: string;
  note: string;
  createdAt: string;
}

export type MarketCondition = "NORMAL" | "ELEVATED" | "CORRECTION" | "CRASH" | "RECOVERY";

export type CrashLevel = "NORMAL_DIP" | "CRASH" | "EXTREME";

export type DiagnosticSignal = "OPPORTUNITY" | "FALLING_KNIFE" | "INCONCLUSIVE";

export interface MarketStatusResponse {
  condition: MarketCondition;
  ihsgPrice: number;
  ihsgPeak: number;
  drawdownPct: number;
  fetchedAt: string;
}

export interface ResponseLevelResponse {
  level: CrashLevel;
  triggerPrice: number;
  deployPct: number;
}

export interface StockPlaybookResponse {
  ticker: string;
  currentPrice: number;
  entryPrice: number;
  levels: ResponseLevelResponse[];
  activeLevel?: CrashLevel;
}

export interface PortfolioPlaybookResponse {
  market: MarketStatusResponse;
  stocks: StockPlaybookResponse[];
  refreshMin: number;
}

export interface DiagnosticResponse {
  marketCrashed: boolean;
  companyBadNews: boolean | null;
  fundamentalsOK: boolean | null;
  belowEntry: boolean;
  signal: DiagnosticSignal;
}

export interface CrashCapitalResponse {
  portfolioId: string;
  amount: number;
  deployed: number;
}

export interface DeploymentLevelPlanResponse {
  level: CrashLevel;
  pct: number;
  amount: number;
}

export interface DeploymentPlanResponse {
  total: number;
  deployed: number;
  remaining: number;
  levels: DeploymentLevelPlanResponse[];
}

export interface ValuationZone {
  grahamNumber: number;
  entryPrice: number;
  exitTarget: number;
}

export interface PricePointResponse {
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export type PriceRange = "1M" | "3M" | "6M" | "1Y" | "ALL";

export interface HoldingWeight {
  ticker: string;
  value: number;
  pct: number;
}

export interface SectorWeight {
  sector: string;
  value: number;
  pct: number;
}

export interface DeploymentSettingsResponse {
  normal: number;
  crash: number;
  extreme: number;
}

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

export interface ScreenerCheckResponse {
  key: string;
  label: string;
  status: "PASS" | "FAIL";
  value: number;
  limit: number;
}

export interface ScreenerItemResponse {
  ticker: string;
  sector: string;
  price?: number;
  roe?: number;
  der?: number;
  eps?: number;
  pbv?: number;
  per?: number;
  dividendYield?: number;
  grahamNumber?: number;
  entryPrice?: number;
  exitTarget?: number;
  verdict?: string;
  checks: ScreenerCheckResponse[];
  passed: boolean;
  score: number;
  fetchedAt?: string;
}

export interface DividendHistoryItemResponse {
  exDate: string;
  amount: number;
}

export interface DGRItemResponse {
  year: number;
  dps: number;
  growthPct: number;
}

export interface YoCPointResponse {
  date: string;
  yoc: number;
}

export interface DividendCalendarEntryResponse {
  ticker: string;
  exDate: string;
  amount: number;
  isProjection: boolean;
  totalIncome: number;
}

export interface MonthlyIncomeItemResponse {
  month: number;
  amount: number;
}

export interface StockIncomeItemResponse {
  ticker: string;
  annualIncome: number;
  dividendYield: number;
  lots: number;
}

export interface DividendIncomeSummaryResponse {
  totalAnnualIncome: number;
  perStock: StockIncomeItemResponse[];
  monthlyBreakdown: MonthlyIncomeItemResponse[];
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

export type ActionType = "BUY" | "AVERAGE_DOWN" | "AVERAGE_UP" | "SELL_EXIT" | "SELL_STOP" | "HOLD";

export interface CheckResultResponse {
  key: string;
  label: string;
  type: "AUTO" | "MANUAL";
  status: "PASS" | "FAIL" | "PENDING";
  detail: string;
}

export interface SuggestionResponse {
  action: string;
  ticker: string;
  lots: number;
  pricePerShare: number;
  grossCost: number;
  fee: number;
  tax: number;
  netCost: number;
  newAvgBuyPrice: number;
  newPositionLots: number;
  newPositionPct: number;
  capitalGainPct: number;
}

export interface ChecklistEvaluationResponse {
  action: string;
  ticker: string;
  checks: CheckResultResponse[];
  allPassed: boolean;
  suggestion?: SuggestionResponse;
}

export type AlertSeverity = "MINOR" | "WARNING" | "CRITICAL";

export type AlertStatus = "ACTIVE" | "ACKNOWLEDGED" | "RESOLVED";

export interface FundamentalAlertResponse {
  id: string;
  ticker: string;
  metric: string;
  severity: AlertSeverity;
  oldValue: number;
  newValue: number;
  changePct: number;
  status: AlertStatus;
  detectedAt: string;
  resolvedAt?: string;
}

export type TransactionType = "BUY" | "SELL" | "DIVIDEND";

export interface TransactionRecordResponse {
  id: string;
  type: TransactionType;
  date: string;
  ticker: string;
  portfolioId: string;
  portfolioName: string;
  lots: number;
  price: number;
  fee: number;
  tax: number;
  total: number;
  createdAt: string;
}

export interface TransactionSummaryResponse {
  totalBuyAmount: number;
  totalSellAmount: number;
  totalDividendAmount: number;
  totalFees: number;
  transactionCount: number;
}

export interface TransactionListResponse {
  items: TransactionRecordResponse[];
  summary: TransactionSummaryResponse;
}

export interface DashboardOverviewResponse {
  totalMarketValue: number;
  totalCostBasis: number;
  totalPlAmount: number;
  totalPlPercent: number;
  totalDividendIncome: number;
  portfolios: DashboardPortfolioSummary[];
  topGainers: HoldingPLResponse[];
  topLosers: HoldingPLResponse[];
  portfolioAllocation: AllocationItemResponse[];
  sectorAllocation: AllocationItemResponse[];
  recentTransactions: TransactionRecordResponse[];
  winRate: number;
  holdingCount: number;
  winningCount: number;
}

export interface DashboardPortfolioSummary {
  id: string;
  name: string;
  mode: string;
  marketValue: number;
  costBasis: number;
  plAmount: number;
  plPercent: number;
  weight: number;
}

export interface HoldingPLResponse {
  ticker: string;
  portfolioId: string;
  portfolioName: string;
  marketValue: number;
  costBasis: number;
  plAmount: number;
  plPercent: number;
}

export interface AllocationItemResponse {
  label: string;
  value: number;
  pct: number;
}
