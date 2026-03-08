<script lang="ts">
import { t } from "../../i18n";
import Badge from "../../lib/components/Badge.svelte";
import Button from "../../lib/components/Button.svelte";
import { formatRupiah } from "../../lib/format";
import type { CrashLevel, StockPlaybookResponse } from "../../lib/types";

let {
  stock,
  onDiagnostic,
}: {
  stock: StockPlaybookResponse;
  onDiagnostic: (ticker: string) => void;
} = $props();

const levelLabels: Record<CrashLevel, string> = {
  NORMAL_DIP: t("crashPlaybook.normalDip"),
  CRASH: t("crashPlaybook.crash"),
  EXTREME: t("crashPlaybook.extreme"),
};

const hasActiveLevel = $derived(stock.activeLevel != null);
</script>

<div class="rounded-lg border border-border-default bg-bg-elevated p-4">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <h3 class="font-display text-base font-semibold text-text-primary">{stock.ticker}</h3>
      {#if hasActiveLevel}
        <Badge variant="loss">{t("crashPlaybook.levelHit")}</Badge>
      {/if}
    </div>
    <span class="font-mono text-sm font-medium text-text-primary">{formatRupiah(stock.currentPrice)}</span>
  </div>

  <div class="mt-1 text-xs text-text-secondary">
    {t("crashPlaybook.entryTarget")} <span class="font-mono">{formatRupiah(stock.entryPrice)}</span>
  </div>

  <div class="mt-3 space-y-2">
    {#each stock.levels as level}
      {@const isActive = stock.activeLevel === level.level}
      <div class="flex items-center justify-between rounded-md px-3 py-2 text-sm {isActive ? 'bg-negative-bg border border-negative/20' : 'bg-bg-primary'}">
        <div class="flex items-center gap-2">
          <span class="font-medium {isActive ? 'text-loss' : 'text-text-secondary'}">{levelLabels[level.level]}</span>
          <span class="text-xs text-text-secondary">({level.deployPct}%)</span>
        </div>
        <span class="font-mono {isActive ? 'text-loss font-semibold' : 'text-text-primary'}">{formatRupiah(level.triggerPrice)}</span>
      </div>
    {/each}
  </div>

  {#if hasActiveLevel}
    <div class="mt-3">
      <Button variant="secondary" size="sm" onclick={() => onDiagnostic(stock.ticker)}>
        {t("common.runDiagnostic")}
      </Button>
    </div>
  {/if}
</div>
