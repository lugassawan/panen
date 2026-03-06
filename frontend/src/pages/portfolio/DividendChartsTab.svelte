<script lang="ts">
import type { HoldingDetailResponse } from "../../lib/types";
import DgrChart from "./DgrChart.svelte";
import DividendCalendarPanel from "./DividendCalendarPanel.svelte";
import DividendIncomeChart from "./DividendIncomeChart.svelte";
import YocProgressionChart from "./YocProgressionChart.svelte";

interface Props {
  portfolioId: string;
  holdings: HoldingDetailResponse[];
}

let { portfolioId, holdings }: Props = $props();

let tickers = $derived(holdings.map((h) => h.ticker));
</script>

<div class="space-y-6" data-testid="dividend-charts-tab">
  <DividendIncomeChart {portfolioId} />

  <div class="grid gap-6 lg:grid-cols-2">
    <DgrChart {tickers} />
    <YocProgressionChart {portfolioId} {tickers} />
  </div>

  <DividendCalendarPanel {portfolioId} />
</div>
