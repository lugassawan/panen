<script lang="ts">
import { LoaderCircle } from "lucide-svelte";
import { LookupStock } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import AlertComponent from "../../lib/components/Alert.svelte";
import DataTimestamp from "../../lib/components/DataTimestamp.svelte";
import Input from "../../lib/components/Input.svelte";
import Select from "../../lib/components/Select.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import { formatDecimal, formatPercent, formatRupiah } from "../../lib/format";
import { alerts } from "../../lib/stores/alerts.svelte";
import type {
  FundamentalAlertResponse,
  RiskProfile,
  StockValuationResponse,
} from "../../lib/types";
import { getVerdictDisplay } from "../../lib/verdict";

let ticker = $state("");
let riskProfile = $state<RiskProfile>("MODERATE");
let result = $state<StockValuationResponse | null>(null);
let loading = $state(false);
let error = $state<string | null>(null);
let tickerAlerts = $state<FundamentalAlertResponse[]>([]);

async function lookup() {
  const code = ticker.trim().toUpperCase();
  if (!code) return;

  loading = true;
  error = null;
  result = null;

  try {
    result = await LookupStock(code, riskProfile);
    tickerAlerts = (await alerts.loadAlertsByTicker(code)).filter(
      (a: FundamentalAlertResponse) => a.status === "ACTIVE",
    );
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
  } finally {
    loading = false;
  }
}

function percentInRange(value: number, min: number, max: number): number {
  if (max === min) return 50;
  return Math.min(100, Math.max(0, ((value - min) / (max - min)) * 100));
}
</script>

<div class="mx-auto max-w-2xl px-4 py-8">
  <h1 class="mb-6 text-2xl font-bold text-text-primary font-display">{t("lookup.title")}</h1>
  <!-- Search Form -->
  <form
    onsubmit={(e) => { e.preventDefault(); lookup(); }}
    class="mb-8 flex gap-2"
  >
    <Input
      bind:value={ticker}
      placeholder={t("lookup.tickerPlaceholder")}
      aria-label="Stock ticker"
      class="flex-1 uppercase placeholder:normal-case placeholder:text-text-muted"
    />
    <Select
      bind:value={riskProfile}
      aria-label="Risk profile"
      class="!w-auto"
    >
      <option value="CONSERVATIVE">{t("screener.conservative")}</option>
      <option value="MODERATE">{t("screener.moderate")}</option>
      <option value="AGGRESSIVE">{t("screener.aggressive")}</option>
    </Select>
    <button
      type="submit"
      disabled={loading}
      class="rounded bg-green-700 px-5 py-2 text-sm font-medium text-text-inverse hover:bg-green-800 disabled:opacity-50 focus-ring transition-fast"
    >
      {loading ? t("lookup.lookingUp") : t("lookup.lookup")}
    </button>
  </form>

  <!-- Loading -->
  {#if loading}
    <div class="flex items-center justify-center gap-2 py-12 text-text-secondary" role="status">
      <LoaderCircle size={20} strokeWidth={2} class="animate-spin" />
      <span>{t("lookup.fetchingValuation")}</span>
    </div>
  {/if}

  <!-- Error -->
  {#if error}
    <div class="mb-6 rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
      {error}
    </div>
  {/if}

  <!-- Results -->
  {#if result}
    {@const verdict = getVerdictDisplay(result.verdict)}
    {@const pct52 = percentInRange(result.price, result.low52Week, result.high52Week)}

    <!-- Verdict Banner -->
    <div class="mb-6 rounded border p-4 text-center {verdict.bgClass}">
      <div class="text-2xl font-bold {verdict.colorClass}">
        <span aria-hidden="true">{verdict.icon}</span>
        {verdict.label}
      </div>
      <p class="mt-1 text-sm text-text-secondary">{verdict.description}</p>
      <p class="mt-1 text-xs text-text-muted">
        {result.ticker} &middot; {result.riskProfile} risk profile
      </p>
    </div>

    <!-- Inline Alerts -->
    {#each tickerAlerts as a (a.id)}
      <div class="mb-4">
        <AlertComponent variant={a.severity === "CRITICAL" ? "negative" : a.severity === "WARNING" ? "warning" : "info"}>
          <span class="font-semibold">{a.severity}:</span>
          {a.metric.toUpperCase().replace("_", " ")} {t("alerts.changed")} {formatDecimal(a.oldValue)} → {formatDecimal(a.newValue)}
          ({a.changePct > 0 ? "+" : ""}{formatDecimal(a.changePct)}%)
        </AlertComponent>
      </div>
    {/each}

    <!-- Price Card -->
    <div class="mb-4 rounded border border-border-default bg-bg-elevated p-4">
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">{t("lookup.currentPrice")}</h2>
      <div class="text-2xl font-bold font-mono text-green-700">{formatRupiah(result.price)}</div>
      <div class="mt-3">
        <div class="flex justify-between text-xs text-text-muted">
          <span>{t("lookup.week52Low")} {formatRupiah(result.low52Week)}</span>
          <span>{t("lookup.week52High")} {formatRupiah(result.high52Week)}</span>
        </div>
        <div class="relative mt-1 h-2 rounded-full bg-bg-tertiary">
          <div
            class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-green-700"
            style="left: {pct52}%"
            title="Current price position"
          ></div>
        </div>
      </div>
    </div>

    <!-- Graham Valuation Card -->
    <div class="mb-4 rounded border border-border-default bg-bg-elevated p-4">
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">{t("lookup.grahamValuation")}</h2>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <Tooltip text="Intrinsic value estimate using Benjamin Graham's formula: sqrt(22.5 * EPS * BVPS)">
            <div class="text-xs text-text-muted underline decoration-dotted cursor-help">{t("lookup.grahamNumber")}</div>
          </Tooltip>
          <div class="text-lg font-semibold font-mono" data-testid="graham-number">{formatRupiah(result.grahamNumber)}</div>
        </div>
        <div>
          <Tooltip text="How much the current price is below the Graham Number. Higher = more undervalued.">
            <div class="text-xs text-text-muted underline decoration-dotted cursor-help">{t("lookup.marginOfSafety")}</div>
          </Tooltip>
          <div class="text-lg font-semibold font-mono">{formatPercent(result.marginOfSafety)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">{t("lookup.entryPrice")}</div>
          <div class="text-lg font-semibold font-mono text-profit" data-testid="entry-price">{formatRupiah(result.entryPrice)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">{t("lookup.exitTarget")}</div>
          <div class="text-lg font-semibold font-mono text-loss">{formatRupiah(result.exitTarget)}</div>
        </div>
      </div>
    </div>

    <!-- PBV Band Card (conditional) -->
    {#if result.pbvBand}
      {@const pbvPct = percentInRange(result.pbv, result.pbvBand.min, result.pbvBand.max)}
      <div class="mb-4 rounded border border-border-default bg-bg-elevated p-4" data-testid="pbv-band">
        <Tooltip text="Price-to-Book Value historical range. Position shows where current PBV sits relative to its 5-year band." position="right">
          <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted underline decoration-dotted cursor-help">{t("lookup.pbvBand")}</h2>
        </Tooltip>
        <div class="mb-2 text-lg font-semibold font-mono">{t("lookup.currentPbv")} {formatDecimal(result.pbv)}</div>
        <div class="grid grid-cols-4 gap-2 text-center text-xs">
          <div>
            <div class="text-text-muted">{t("lookup.min")}</div>
            <div class="font-mono">{formatDecimal(result.pbvBand.min)}</div>
          </div>
          <div>
            <div class="text-text-muted">{t("lookup.avg")}</div>
            <div class="font-mono">{formatDecimal(result.pbvBand.avg)}</div>
          </div>
          <div>
            <div class="text-text-muted">{t("lookup.median")}</div>
            <div class="font-mono">{formatDecimal(result.pbvBand.median)}</div>
          </div>
          <div>
            <div class="text-text-muted">{t("lookup.max")}</div>
            <div class="font-mono">{formatDecimal(result.pbvBand.max)}</div>
          </div>
        </div>
        <div class="relative mt-2 h-2 rounded-full bg-bg-tertiary">
          <div
            class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-green-700"
            style="left: {pbvPct}%"
            title="Current PBV position"
          ></div>
        </div>
      </div>
    {/if}

    <!-- PER Band Card (conditional) -->
    {#if result.perBand}
      {@const perPct = percentInRange(result.per, result.perBand.min, result.perBand.max)}
      <div class="mb-4 rounded border border-border-default bg-bg-elevated p-4" data-testid="per-band">
        <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">{t("lookup.perBand")}</h2>
        <div class="mb-2 text-lg font-semibold font-mono">{t("lookup.currentPer")} {formatDecimal(result.per)}</div>
        <div class="grid grid-cols-4 gap-2 text-center text-xs">
          <div>
            <div class="text-text-muted">{t("lookup.min")}</div>
            <div class="font-mono">{formatDecimal(result.perBand.min)}</div>
          </div>
          <div>
            <div class="text-text-muted">{t("lookup.avg")}</div>
            <div class="font-mono">{formatDecimal(result.perBand.avg)}</div>
          </div>
          <div>
            <div class="text-text-muted">{t("lookup.median")}</div>
            <div class="font-mono">{formatDecimal(result.perBand.median)}</div>
          </div>
          <div>
            <div class="text-text-muted">{t("lookup.max")}</div>
            <div class="font-mono">{formatDecimal(result.perBand.max)}</div>
          </div>
        </div>
        <div class="relative mt-2 h-2 rounded-full bg-bg-tertiary">
          <div
            class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-green-700"
            style="left: {perPct}%"
            title="Current PER position"
          ></div>
        </div>
      </div>
    {/if}

    <!-- Key Metrics Grid -->
    <div class="mb-4 rounded border border-border-default bg-bg-elevated p-4">
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">{t("lookup.keyMetrics")}</h2>
      <div class="grid grid-cols-3 gap-4 text-sm">
        <div>
          <div class="text-xs text-text-muted">EPS</div>
          <div class="font-semibold font-mono">{formatRupiah(result.eps)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">BVPS</div>
          <div class="font-semibold font-mono">{formatRupiah(result.bvps)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">ROE</div>
          <div class="font-semibold font-mono">{formatPercent(result.roe)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">DER</div>
          <div class="font-semibold font-mono">{formatDecimal(result.der)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">{t("lookup.dividendYield")}</div>
          <div class="font-semibold font-mono">{formatPercent(result.dividendYield)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">{t("lookup.payoutRatio")}</div>
          <div class="font-semibold font-mono">{formatPercent(result.payoutRatio)}</div>
        </div>
      </div>
    </div>

    <!-- Metadata Footer -->
    <div class="flex items-center justify-center gap-3 text-xs text-text-muted">
      <span>{t("lookup.source")} {result.source}</span>
      <DataTimestamp date={result.fetchedAt} label={t("lookup.fetched")} />
    </div>
  {/if}
</div>
