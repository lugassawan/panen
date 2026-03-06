<script lang="ts">
import { ArrowLeft } from "lucide-svelte";
import Button from "../../lib/components/Button.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import { formatPercent, formatRupiah } from "../../lib/format";
import {
  currentValue as calcCurrentValue,
  overallPL as calcOverallPL,
  totalInvested as calcTotalInvested,
} from "../../lib/portfolio";
import type { PortfolioDetailResponse } from "../../lib/types";
import AddHoldingForm from "./AddHoldingForm.svelte";
import ChartsTab from "./ChartsTab.svelte";
import DividendChartsTab from "./DividendChartsTab.svelte";
import DividendMetricsPanel from "./DividendMetricsPanel.svelte";
import DividendRankingPanel from "./DividendRankingPanel.svelte";
import HoldingsTable from "./HoldingsTable.svelte";
import TrailingStopPanel from "./TrailingStopPanel.svelte";

interface Props {
  detail: PortfolioDetailResponse;
  onBack: () => void;
  onChecklist: (ticker: string) => void;
  onHoldingAdded: () => void;
}

let { detail, onBack, onChecklist, onHoldingAdded }: Props = $props();

type TabId = "holdings" | "charts" | "dividends";
let activeTab = $state<TabId>("holdings");

const TAB_ACCENT: Record<string, string> = {
  VALUE: "border-green-700 text-green-700",
  DIVIDEND: "border-gold-500 text-gold-500",
};

let TABS: TabId[] = $derived(
  detail.portfolio.mode === "DIVIDEND"
    ? ["holdings", "charts", "dividends"]
    : ["holdings", "charts"],
);

function handleTabKeydown(e: KeyboardEvent) {
  const idx = TABS.indexOf(activeTab);
  let next = idx;
  if (e.key === "ArrowRight") next = (idx + 1) % TABS.length;
  else if (e.key === "ArrowLeft") next = (idx - 1 + TABS.length) % TABS.length;
  else if (e.key === "Home") next = 0;
  else if (e.key === "End") next = TABS.length - 1;
  else return;
  e.preventDefault();
  activeTab = TABS[next];
  const tablist = (e.currentTarget as HTMLElement).parentElement;
  const btn = tablist?.querySelectorAll<HTMLButtonElement>('[role="tab"]')[next];
  btn?.focus();
}

const MODE_BADGE: Record<string, string> = {
  VALUE: "bg-green-100 text-green-700",
  DIVIDEND: "bg-gold-100 text-gold-700",
};

let totalInvested = $derived(calcTotalInvested(detail.holdings));
let currentValue = $derived(calcCurrentValue(detail.holdings));
let overallPL = $derived(calcOverallPL(detail.holdings));
</script>

<div class="mb-6 flex items-center gap-3">
  <button
    type="button"
    class="rounded p-1 text-text-secondary hover:bg-bg-tertiary hover:text-text-primary focus-ring transition-fast"
    onclick={onBack}
    aria-label="Back to list"
  >
    <ArrowLeft size={20} strokeWidth={2} aria-hidden="true" />
  </button>
  <h2 class="text-xl font-semibold text-text-primary">{detail.portfolio.name}</h2>
  <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {MODE_BADGE[detail.portfolio.mode]}">
    {detail.portfolio.mode === "VALUE" ? "Value" : "Dividend"}
  </span>
</div>

<!-- Summary Bar -->
<div class="mb-6 grid {detail.portfolio.mode === 'DIVIDEND' ? 'grid-cols-4' : 'grid-cols-3'} gap-4">
  <div class="rounded border border-border-default bg-bg-elevated p-4" data-testid="total-invested">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Total Invested</p>
    <p class="mt-1 text-lg font-medium">{formatRupiah(totalInvested)}</p>
  </div>
  <div class="rounded border border-border-default bg-bg-elevated p-4" data-testid="current-value">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Current Value</p>
    <p class="mt-1 text-lg font-medium">{formatRupiah(currentValue)}</p>
  </div>
  <div class="rounded border border-border-default bg-bg-elevated p-4" data-testid="overall-pl">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">
      <Tooltip text="Unrealized profit/loss across all holdings based on current market prices">
        <span class="underline decoration-dotted cursor-help">Overall P/L</span>
      </Tooltip>
    </p>
    <p class="mt-1 text-lg font-medium font-mono {overallPL >= 0 ? 'text-profit' : 'text-loss'}">
      {overallPL >= 0 ? "+" : ""}{formatPercent(overallPL)}
    </p>
  </div>
  {#if detail.portfolio.mode === "DIVIDEND"}
    {@const portfolioYield = detail.holdings.find((h) => h.dividendMetrics)?.dividendMetrics?.portfolioYield ?? 0}
    <div class="rounded border border-border-default bg-bg-elevated p-4" data-testid="portfolio-yield">
      <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">
        <Tooltip text="Weighted average dividend yield across all holdings in this portfolio">
          <span class="underline decoration-dotted cursor-help">Portfolio Yield</span>
        </Tooltip>
      </p>
      <p class="mt-1 text-lg font-medium font-mono text-text-primary">
        {formatPercent(portfolioYield)}
      </p>
    </div>
  {/if}
</div>

<!-- Tab Bar -->
<div class="mb-6 flex gap-0 border-b border-border-default" role="tablist">
  {#each TABS as tab}
    <button
      type="button"
      role="tab"
      aria-selected={activeTab === tab}
      aria-controls="panel-{tab}"
      tabindex={activeTab === tab ? 0 : -1}
      class="px-4 py-2 text-sm font-medium transition-fast focus-ring -mb-px capitalize {activeTab === tab ? `border-b-2 ${TAB_ACCENT[detail.portfolio.mode]}` : 'text-text-secondary hover:text-text-primary'}"
      onclick={() => (activeTab = tab)}
      onkeydown={handleTabKeydown}
    >
      {tab}
    </button>
  {/each}
</div>

{#if activeTab === "holdings"}
  <div id="panel-holdings" role="tabpanel">
    <!-- Add Holding -->
    <div class="mb-6 rounded border border-border-default bg-bg-elevated p-4">
      <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Add Holding</h3>
      <AddHoldingForm portfolioId={detail.portfolio.id} onAdded={onHoldingAdded} />
    </div>

    <!-- Holdings Table -->
    <div class="mb-6">
      <HoldingsTable holdings={detail.holdings} {onChecklist} />
    </div>

    <!-- Trailing Stops (VALUE mode only) -->
    {#if detail.portfolio.mode === "VALUE" && detail.holdings.some((h) => h.trailingStop)}
      <div class="mb-6">
        <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Trailing Stops</h3>
        <div class="space-y-3">
          {#each detail.holdings as holding}
            {#if holding.trailingStop}
              <div>
                <p class="mb-1 text-sm font-medium text-text-primary">{holding.ticker}</p>
                <TrailingStopPanel trailingStop={holding.trailingStop} />
              </div>
            {/if}
          {/each}
        </div>
      </div>
    {/if}

    <!-- Dividend Metrics (DIVIDEND mode only) -->
    {#if detail.portfolio.mode === "DIVIDEND" && detail.holdings.some((h) => h.dividendMetrics)}
      <div class="mb-6">
        <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Dividend Metrics</h3>
        <div class="space-y-3">
          {#each detail.holdings as holding}
            {#if holding.dividendMetrics}
              <DividendMetricsPanel ticker={holding.ticker} dividendMetrics={holding.dividendMetrics} />
            {/if}
          {/each}
        </div>
      </div>

      <div class="mb-6">
        <DividendRankingPanel portfolioId={detail.portfolio.id} />
      </div>
    {/if}
  </div>
{/if}

{#if activeTab === "charts"}
  <div id="panel-charts" role="tabpanel">
    <ChartsTab holdings={detail.holdings} portfolioMode={detail.portfolio.mode} />
  </div>
{/if}

{#if activeTab === "dividends"}
  <div id="panel-dividends" role="tabpanel">
    <DividendChartsTab portfolioId={detail.portfolio.id} holdings={detail.holdings} />
  </div>
{/if}
