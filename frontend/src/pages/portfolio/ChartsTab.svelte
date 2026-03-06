<script lang="ts">
import { GetHoldingSectors } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import { holdingWeights, sectorWeights } from "../../lib/portfolio";
import type {
  HoldingDetailResponse,
  HoldingWeight,
  Mode,
  SectorWeight,
  ValuationZone,
} from "../../lib/types";
import CompositionChart from "./CompositionChart.svelte";
import PlBarChart from "./PlBarChart.svelte";
import PriceHistoryChart from "./PriceHistoryChart.svelte";
import SectorWarnings from "./SectorWarnings.svelte";

interface Props {
  holdings: HoldingDetailResponse[];
  portfolioMode: Mode;
}

let { holdings, portfolioMode }: Props = $props();

let sectorMap = $state<Record<string, string>>({});
let loading = $state(true);
let error = $state<string | null>(null);

let weights: HoldingWeight[] = $derived(holdingWeights(holdings));
let sectors: SectorWeight[] = $derived(sectorWeights(holdings, sectorMap));

let valuationMap: Record<string, ValuationZone> = $derived.by(() => {
  const map: Record<string, ValuationZone> = {};
  for (const h of holdings) {
    map[h.ticker] = {
      grahamNumber: h.grahamNumber ?? 0,
      entryPrice: h.entryPrice ?? 0,
      exitTarget: h.exitTarget ?? 0,
    };
  }
  return map;
});

$effect(() => {
  const tickers = holdings.map((h) => h.ticker);
  if (tickers.length === 0) {
    loading = false;
    return;
  }
  loading = true;
  error = null;
  GetHoldingSectors(tickers)
    .then((result) => {
      sectorMap = result;
    })
    .catch((e) => {
      error = e instanceof Error ? e.message : String(e);
    })
    .finally(() => {
      loading = false;
    });
});
</script>

{#if loading}
  <div class="flex items-center justify-center py-12">
    <p class="text-sm text-text-muted">{t("chart.loadingChart")}</p>
  </div>
{:else if error}
  <div class="rounded border border-border-default bg-bg-elevated p-6 text-center">
    <p class="text-sm text-loss">{error}</p>
  </div>
{:else}
  <div class="space-y-6">
    <PriceHistoryChart tickers={holdings.map(h => h.ticker)} valuations={valuationMap} />

    <SectorWarnings sectorWeights={sectors} />

    <div class="grid gap-6 lg:grid-cols-2">
      <PlBarChart {holdings} />
      <CompositionChart holdingWeights={weights} sectorWeights={sectors} {portfolioMode} />
    </div>
  </div>
{/if}
