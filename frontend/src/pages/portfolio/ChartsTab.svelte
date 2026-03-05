<script lang="ts">
import { GetHoldingSectors } from "../../../wailsjs/go/backend/App";
import { holdingWeights, sectorWeights } from "../../lib/portfolio";
import type { HoldingDetailResponse, HoldingWeight, Mode, SectorWeight } from "../../lib/types";
import CompositionChart from "./CompositionChart.svelte";
import PlBarChart from "./PlBarChart.svelte";
import SectorWarnings from "./SectorWarnings.svelte";

interface Props {
  holdings: HoldingDetailResponse[];
  portfolioId: string;
  portfolioMode: Mode;
}

let { holdings, portfolioId, portfolioMode }: Props = $props();

let sectorMap = $state<Record<string, string>>({});
let loading = $state(true);

let weights: HoldingWeight[] = $derived(holdingWeights(holdings));
let sectors: SectorWeight[] = $derived(sectorWeights(holdings, sectorMap));

$effect(() => {
  const tickers = holdings.map((h) => h.ticker);
  if (tickers.length === 0) {
    loading = false;
    return;
  }
  loading = true;
  GetHoldingSectors(tickers)
    .then((result) => {
      sectorMap = result;
    })
    .finally(() => {
      loading = false;
    });
});
</script>

{#if loading}
  <div class="flex items-center justify-center py-12">
    <p class="text-sm text-text-muted">Loading chart data…</p>
  </div>
{:else}
  <SectorWarnings sectorWeights={sectors} />

  <div class="grid gap-6 lg:grid-cols-2">
    <PlBarChart {holdings} />
    <CompositionChart holdingWeights={weights} sectorWeights={sectors} {portfolioMode} />
  </div>
{/if}
