<script lang="ts">
import type { HoldingDetailResponse } from "../../lib/types";
import DGRChart from "./DGRChart.svelte";
import DividendCalendarPanel from "./DividendCalendarPanel.svelte";
import DividendIncomeChart from "./DividendIncomeChart.svelte";
import YoCProgressionChart from "./YoCProgressionChart.svelte";

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
    <DGRChart {tickers} />
    <YoCProgressionChart {portfolioId} {tickers} />
  </div>

  <DividendCalendarPanel {portfolioId} />
</div>
