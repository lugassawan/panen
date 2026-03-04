<script lang="ts">
import { LoaderCircle } from "lucide-svelte";
import { LookupStock } from "../../../wailsjs/go/backend/App";
import Input from "../../lib/components/Input.svelte";
import Select from "../../lib/components/Select.svelte";
import { formatDecimal, formatPercent, formatRupiah } from "../../lib/format";
import type { RiskProfile, StockValuationResponse } from "../../lib/types";
import { getVerdictDisplay } from "../../lib/verdict";

let ticker = $state("");
let riskProfile = $state<RiskProfile>("MODERATE");
let result = $state<StockValuationResponse | null>(null);
let loading = $state(false);
let error = $state<string | null>(null);

async function lookup() {
  const t = ticker.trim().toUpperCase();
  if (!t) return;

  loading = true;
  error = null;
  result = null;

  try {
    result = await LookupStock(t, riskProfile);
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
  <!-- Search Form -->
  <form
    onsubmit={(e) => { e.preventDefault(); lookup(); }}
    class="mb-8 flex gap-2"
  >
    <Input
      bind:value={ticker}
      placeholder="Ticker (e.g. BBCA)"
      aria-label="Stock ticker"
      class="flex-1 uppercase placeholder:normal-case placeholder:text-text-muted"
    />
    <Select
      bind:value={riskProfile}
      aria-label="Risk profile"
      class="!w-auto"
    >
      <option value="CONSERVATIVE">Conservative</option>
      <option value="MODERATE">Moderate</option>
      <option value="AGGRESSIVE">Aggressive</option>
    </Select>
    <button
      type="submit"
      disabled={loading}
      class="rounded bg-green-700 px-5 py-2 text-sm font-medium text-text-inverse hover:bg-green-800 disabled:opacity-50 focus-ring transition-fast"
    >
      {loading ? "Looking up\u2026" : "Lookup"}
    </button>
  </form>

  <!-- Loading -->
  {#if loading}
    <div class="flex items-center justify-center gap-2 py-12 text-text-secondary" role="status">
      <LoaderCircle size={20} strokeWidth={2} class="animate-spin" />
      <span>Fetching valuation data...</span>
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

    <!-- Price Card -->
    <div class="mb-4 rounded border border-border-default bg-bg-elevated p-4">
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Current Price</h2>
      <div class="text-2xl font-bold font-mono text-green-700">{formatRupiah(result.price)}</div>
      <div class="mt-3">
        <div class="flex justify-between text-xs text-text-muted">
          <span>52W Low: {formatRupiah(result.low52Week)}</span>
          <span>52W High: {formatRupiah(result.high52Week)}</span>
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
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Graham Valuation</h2>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <div class="text-xs text-text-muted">Graham Number</div>
          <div class="text-lg font-semibold font-mono" data-testid="graham-number">{formatRupiah(result.grahamNumber)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">Margin of Safety</div>
          <div class="text-lg font-semibold font-mono">{formatPercent(result.marginOfSafety)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">Entry Price</div>
          <div class="text-lg font-semibold font-mono text-profit" data-testid="entry-price">{formatRupiah(result.entryPrice)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">Exit Target</div>
          <div class="text-lg font-semibold font-mono text-loss">{formatRupiah(result.exitTarget)}</div>
        </div>
      </div>
    </div>

    <!-- PBV Band Card (conditional) -->
    {#if result.pbvBand}
      {@const pbvPct = percentInRange(result.pbv, result.pbvBand.min, result.pbvBand.max)}
      <div class="mb-4 rounded border border-border-default bg-bg-elevated p-4" data-testid="pbv-band">
        <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">PBV Band</h2>
        <div class="mb-2 text-lg font-semibold font-mono">Current PBV: {formatDecimal(result.pbv)}</div>
        <div class="grid grid-cols-4 gap-2 text-center text-xs">
          <div>
            <div class="text-text-muted">Min</div>
            <div class="font-mono">{formatDecimal(result.pbvBand.min)}</div>
          </div>
          <div>
            <div class="text-text-muted">Avg</div>
            <div class="font-mono">{formatDecimal(result.pbvBand.avg)}</div>
          </div>
          <div>
            <div class="text-text-muted">Median</div>
            <div class="font-mono">{formatDecimal(result.pbvBand.median)}</div>
          </div>
          <div>
            <div class="text-text-muted">Max</div>
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
        <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">PER Band</h2>
        <div class="mb-2 text-lg font-semibold font-mono">Current PER: {formatDecimal(result.per)}</div>
        <div class="grid grid-cols-4 gap-2 text-center text-xs">
          <div>
            <div class="text-text-muted">Min</div>
            <div class="font-mono">{formatDecimal(result.perBand.min)}</div>
          </div>
          <div>
            <div class="text-text-muted">Avg</div>
            <div class="font-mono">{formatDecimal(result.perBand.avg)}</div>
          </div>
          <div>
            <div class="text-text-muted">Median</div>
            <div class="font-mono">{formatDecimal(result.perBand.median)}</div>
          </div>
          <div>
            <div class="text-text-muted">Max</div>
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
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Key Metrics</h2>
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
          <div class="text-xs text-text-muted">Dividend Yield</div>
          <div class="font-semibold font-mono">{formatPercent(result.dividendYield)}</div>
        </div>
        <div>
          <div class="text-xs text-text-muted">Payout Ratio</div>
          <div class="font-semibold font-mono">{formatPercent(result.payoutRatio)}</div>
        </div>
      </div>
    </div>

    <!-- Metadata Footer -->
    <div class="text-center text-xs text-text-muted">
      Source: {result.source} &middot; Fetched: {result.fetchedAt}
    </div>
  {/if}
</div>
