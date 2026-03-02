<script lang="ts">
import { LookupStock } from "../../wailsjs/go/backend/App";
import { formatDecimal, formatPercent, formatRupiah } from "../lib/format";
import type { RiskProfile, StockValuationResponse } from "../lib/types";
import { getVerdictDisplay } from "../lib/verdict";

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
    <input
      bind:value={ticker}
      placeholder="Ticker (e.g. BBCA)"
      aria-label="Stock ticker"
      class="flex-1 rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm uppercase placeholder:normal-case placeholder:text-neutral-500 outline-none focus:border-amber-500"
    />
    <select
      bind:value={riskProfile}
      aria-label="Risk profile"
      class="rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
    >
      <option value="CONSERVATIVE">Conservative</option>
      <option value="MODERATE">Moderate</option>
      <option value="AGGRESSIVE">Aggressive</option>
    </select>
    <button
      type="submit"
      disabled={loading}
      class="rounded bg-amber-600 px-5 py-2 text-sm font-medium hover:bg-amber-500 disabled:opacity-50"
    >
      {loading ? "Looking up\u2026" : "Lookup"}
    </button>
  </form>

  <!-- Loading -->
  {#if loading}
    <div class="flex items-center justify-center gap-2 py-12 text-neutral-400" role="status">
      <svg class="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25"></circle>
        <path d="M4 12a8 8 0 018-8" stroke="currentColor" stroke-width="4" stroke-linecap="round" class="opacity-75"></path>
      </svg>
      <span>Fetching valuation data...</span>
    </div>
  {/if}

  <!-- Error -->
  {#if error}
    <div class="mb-6 rounded border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400" role="alert">
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
      <p class="mt-1 text-sm text-neutral-300">{verdict.description}</p>
      <p class="mt-1 text-xs text-neutral-500">
        {result.ticker} &middot; {result.riskProfile} risk profile
      </p>
    </div>

    <!-- Price Card -->
    <div class="mb-4 rounded border border-neutral-800 bg-neutral-900 p-4">
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-neutral-500">Current Price</h2>
      <div class="text-2xl font-bold text-amber-400">{formatRupiah(result.price)}</div>
      <div class="mt-3">
        <div class="flex justify-between text-xs text-neutral-500">
          <span>52W Low: {formatRupiah(result.low52Week)}</span>
          <span>52W High: {formatRupiah(result.high52Week)}</span>
        </div>
        <div class="relative mt-1 h-2 rounded-full bg-neutral-800">
          <div
            class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-amber-400"
            style="left: {pct52}%"
            title="Current price position"
          ></div>
        </div>
      </div>
    </div>

    <!-- Graham Valuation Card -->
    <div class="mb-4 rounded border border-neutral-800 bg-neutral-900 p-4">
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-neutral-500">Graham Valuation</h2>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <div class="text-xs text-neutral-500">Graham Number</div>
          <div class="text-lg font-semibold" data-testid="graham-number">{formatRupiah(result.grahamNumber)}</div>
        </div>
        <div>
          <div class="text-xs text-neutral-500">Margin of Safety</div>
          <div class="text-lg font-semibold">{formatPercent(result.marginOfSafety)}</div>
        </div>
        <div>
          <div class="text-xs text-neutral-500">Entry Price</div>
          <div class="text-lg font-semibold text-emerald-400" data-testid="entry-price">{formatRupiah(result.entryPrice)}</div>
        </div>
        <div>
          <div class="text-xs text-neutral-500">Exit Target</div>
          <div class="text-lg font-semibold text-red-400">{formatRupiah(result.exitTarget)}</div>
        </div>
      </div>
    </div>

    <!-- PBV Band Card (conditional) -->
    {#if result.pbvBand}
      {@const pbvPct = percentInRange(result.pbv, result.pbvBand.min, result.pbvBand.max)}
      <div class="mb-4 rounded border border-neutral-800 bg-neutral-900 p-4" data-testid="pbv-band">
        <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-neutral-500">PBV Band</h2>
        <div class="mb-2 text-lg font-semibold">Current PBV: {formatDecimal(result.pbv)}</div>
        <div class="grid grid-cols-4 gap-2 text-center text-xs">
          <div>
            <div class="text-neutral-500">Min</div>
            <div>{formatDecimal(result.pbvBand.min)}</div>
          </div>
          <div>
            <div class="text-neutral-500">Avg</div>
            <div>{formatDecimal(result.pbvBand.avg)}</div>
          </div>
          <div>
            <div class="text-neutral-500">Median</div>
            <div>{formatDecimal(result.pbvBand.median)}</div>
          </div>
          <div>
            <div class="text-neutral-500">Max</div>
            <div>{formatDecimal(result.pbvBand.max)}</div>
          </div>
        </div>
        <div class="relative mt-2 h-2 rounded-full bg-neutral-800">
          <div
            class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-amber-400"
            style="left: {pbvPct}%"
            title="Current PBV position"
          ></div>
        </div>
      </div>
    {/if}

    <!-- PER Band Card (conditional) -->
    {#if result.perBand}
      {@const perPct = percentInRange(result.per, result.perBand.min, result.perBand.max)}
      <div class="mb-4 rounded border border-neutral-800 bg-neutral-900 p-4" data-testid="per-band">
        <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-neutral-500">PER Band</h2>
        <div class="mb-2 text-lg font-semibold">Current PER: {formatDecimal(result.per)}</div>
        <div class="grid grid-cols-4 gap-2 text-center text-xs">
          <div>
            <div class="text-neutral-500">Min</div>
            <div>{formatDecimal(result.perBand.min)}</div>
          </div>
          <div>
            <div class="text-neutral-500">Avg</div>
            <div>{formatDecimal(result.perBand.avg)}</div>
          </div>
          <div>
            <div class="text-neutral-500">Median</div>
            <div>{formatDecimal(result.perBand.median)}</div>
          </div>
          <div>
            <div class="text-neutral-500">Max</div>
            <div>{formatDecimal(result.perBand.max)}</div>
          </div>
        </div>
        <div class="relative mt-2 h-2 rounded-full bg-neutral-800">
          <div
            class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-amber-400"
            style="left: {perPct}%"
            title="Current PER position"
          ></div>
        </div>
      </div>
    {/if}

    <!-- Key Metrics Grid -->
    <div class="mb-4 rounded border border-neutral-800 bg-neutral-900 p-4">
      <h2 class="mb-3 text-xs font-semibold uppercase tracking-wider text-neutral-500">Key Metrics</h2>
      <div class="grid grid-cols-3 gap-4 text-sm">
        <div>
          <div class="text-xs text-neutral-500">EPS</div>
          <div class="font-semibold">{formatRupiah(result.eps)}</div>
        </div>
        <div>
          <div class="text-xs text-neutral-500">BVPS</div>
          <div class="font-semibold">{formatRupiah(result.bvps)}</div>
        </div>
        <div>
          <div class="text-xs text-neutral-500">ROE</div>
          <div class="font-semibold">{formatPercent(result.roe)}</div>
        </div>
        <div>
          <div class="text-xs text-neutral-500">DER</div>
          <div class="font-semibold">{formatDecimal(result.der)}</div>
        </div>
        <div>
          <div class="text-xs text-neutral-500">Dividend Yield</div>
          <div class="font-semibold">{formatPercent(result.dividendYield)}</div>
        </div>
        <div>
          <div class="text-xs text-neutral-500">Payout Ratio</div>
          <div class="font-semibold">{formatPercent(result.payoutRatio)}</div>
        </div>
      </div>
    </div>

    <!-- Metadata Footer -->
    <div class="text-center text-xs text-neutral-600">
      Source: {result.source} &middot; Fetched: {result.fetchedAt}
    </div>
  {/if}
</div>
