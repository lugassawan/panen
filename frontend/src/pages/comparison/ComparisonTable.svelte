<script lang="ts">
import { t } from "../../i18n";
import Badge from "../../lib/components/Badge.svelte";
import DataTimestamp from "../../lib/components/DataTimestamp.svelte";
import { formatRupiah } from "../../lib/format";
import type { StockValuationResponse, Verdict } from "../../lib/types";
import { getVerdictDisplay } from "../../lib/verdict";
import ComparisonBandRow from "./ComparisonBandRow.svelte";
import {
  FUNDAMENTAL_METRICS,
  findBestIndex,
  type MetricConfig,
  VALUATION_METRICS,
} from "./comparison-metrics";

let { results }: { results: (StockValuationResponse | null)[] } = $props();

let successResults = $derived(results.filter((r): r is StockValuationResponse => r !== null));

function percentInRange(value: number, min: number, max: number): number {
  if (max === min) return 50;
  return Math.min(100, Math.max(0, ((value - min) / (max - min)) * 100));
}

function getMetricValues(metric: MetricConfig): (number | null)[] {
  return results.map((r) => (r ? (r[metric.key] as number) : null));
}

function verdictBadgeVariant(verdict: Verdict): "profit" | "loss" | "warning" {
  if (verdict === "UNDERVALUED") return "profit";
  if (verdict === "OVERVALUED") return "loss";
  return "warning";
}
</script>

{#if successResults.length > 0}
  <div class="overflow-x-auto" data-testid="comparison-table">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border-default">
          <th class="py-2 pr-4 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">
            {t("screener.ticker")}
          </th>
          {#each results as r}
            <th class="py-2 px-3 text-center text-base font-bold font-display text-text-primary">
              {r ? r.ticker : ""}
            </th>
          {/each}
        </tr>
      </thead>
      <tbody>
        <!-- Price -->
        <tr class="border-t border-border-default">
          <td class="py-2 pr-4 text-sm text-text-secondary whitespace-nowrap">{t("screener.price")}</td>
          {#each results as r}
            <td class="py-2 px-3 text-center font-mono font-semibold text-green-700">
              {r ? formatRupiah(r.price) : "--"}
            </td>
          {/each}
        </tr>

        <!-- 52W Range -->
        <tr class="border-t border-border-default">
          <td class="py-2 pr-4 text-sm text-text-secondary whitespace-nowrap">{t("comparison.week52Range")}</td>
          {#each results as r}
            <td class="py-2 px-3">
              {#if r}
                {@const pct = percentInRange(r.price, r.low52Week, r.high52Week)}
                <div class="flex flex-col items-center gap-1">
                  <div class="relative w-full h-2 rounded-full bg-bg-tertiary">
                    <div
                      class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-green-700"
                      style="left: {pct}%"
                    ></div>
                  </div>
                  <div class="flex justify-between w-full text-[10px] text-text-muted font-mono">
                    <span>{formatRupiah(r.low52Week)}</span>
                    <span>{formatRupiah(r.high52Week)}</span>
                  </div>
                </div>
              {:else}
                <span class="text-xs text-text-muted text-center block">--</span>
              {/if}
            </td>
          {/each}
        </tr>

        <!-- Valuation Section -->
        <tr>
          <td colspan={results.length + 1} class="pt-4 pb-1 text-xs font-semibold uppercase tracking-wider text-text-muted">
            {t("comparison.sectionValuation")}
          </td>
        </tr>
        {#each VALUATION_METRICS as metric (metric.key)}
          {@const values = getMetricValues(metric)}
          {@const bestIdx = findBestIndex(values, metric.direction)}
          <tr class="border-t border-border-default">
            <td class="py-2 pr-4 text-sm text-text-secondary whitespace-nowrap">{t(metric.labelKey)}</td>
            {#each results as r, i}
              <td class="py-2 px-3 text-center font-mono {i === bestIdx ? 'text-profit font-semibold' : 'text-text-primary'}">
                {r ? metric.format(r[metric.key] as number) : "--"}
              </td>
            {/each}
          </tr>
        {/each}

        <!-- Verdict -->
        <tr class="border-t border-border-default">
          <td class="py-2 pr-4 text-sm text-text-secondary whitespace-nowrap">{t("holding.verdict")}</td>
          {#each results as r}
            <td class="py-2 px-3 text-center">
              {#if r}
                {@const vd = getVerdictDisplay(r.verdict)}
                <Badge variant={verdictBadgeVariant(r.verdict as Verdict)}>
                  {vd.label}
                </Badge>
              {:else}
                <span class="text-xs text-text-muted">--</span>
              {/if}
            </td>
          {/each}
        </tr>

        <!-- Fundamentals Section -->
        <tr>
          <td colspan={results.length + 1} class="pt-4 pb-1 text-xs font-semibold uppercase tracking-wider text-text-muted">
            {t("comparison.sectionFundamentals")}
          </td>
        </tr>
        {#each FUNDAMENTAL_METRICS as metric (metric.key)}
          {@const values = getMetricValues(metric)}
          {@const bestIdx = findBestIndex(values, metric.direction)}
          <tr class="border-t border-border-default">
            <td class="py-2 pr-4 text-sm text-text-secondary whitespace-nowrap">{t(metric.labelKey)}</td>
            {#each results as r, i}
              <td class="py-2 px-3 text-center font-mono {i === bestIdx ? 'text-profit font-semibold' : 'text-text-primary'}">
                {r ? metric.format(r[metric.key] as number) : "--"}
              </td>
            {/each}
          </tr>
        {/each}

        <!-- Bands Section -->
        <tr>
          <td colspan={results.length + 1} class="pt-4 pb-1 text-xs font-semibold uppercase tracking-wider text-text-muted">
            {t("comparison.sectionBands")}
          </td>
        </tr>
        <ComparisonBandRow label={t("lookup.pbvBand")} {results} bandKey="pbvBand" valueKey="pbv" />
        <ComparisonBandRow label={t("lookup.perBand")} {results} bandKey="perBand" valueKey="per" />

        <!-- Timestamp Footer -->
        <tr class="border-t border-border-default">
          <td class="py-3 pr-4"></td>
          {#each results as r}
            <td class="py-3 px-3 text-center">
              {#if r}
                <DataTimestamp date={r.fetchedAt} label={t("lookup.fetched")} />
              {/if}
            </td>
          {/each}
        </tr>
      </tbody>
    </table>
  </div>
{/if}
