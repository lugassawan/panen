<script lang="ts">
import { AlertTriangle, Check, ShieldAlert, TrendingDown } from "lucide-svelte";
import { t } from "../../i18n";
import { formatRupiah } from "../../lib/format";
import type { TrailingStopResponse } from "../../lib/types";

let {
  trailingStop,
}: {
  trailingStop: TrailingStopResponse;
} = $props();
</script>

<div
  class="rounded border {trailingStop.triggered ? 'border-negative/40 bg-negative-bg' : 'border-border-default bg-bg-elevated'} p-4"
  data-testid="trailing-stop-panel"
>
  <div class="mb-3 flex items-center gap-2">
    {#if trailingStop.triggered}
      <ShieldAlert size={16} strokeWidth={2} class="text-loss" />
      <h4 class="text-sm font-semibold text-loss">{t("trailingStop.triggered")}</h4>
    {:else}
      <TrendingDown size={16} strokeWidth={2} class="text-text-muted" />
      <h4 class="text-sm font-semibold text-text-primary">{t("trailingStop.title")}</h4>
    {/if}
  </div>

  <div class="grid grid-cols-3 gap-4">
    <div>
      <p class="text-xs text-text-muted">{t("trailingStop.peakPrice")}</p>
      <p class="font-mono text-sm text-text-primary">{formatRupiah(trailingStop.peakPrice)}</p>
    </div>
    <div>
      <p class="text-xs text-text-muted">{t("trailingStop.stopLevel")}</p>
      <p class="font-mono text-sm {trailingStop.triggered ? 'text-loss' : 'text-text-primary'}">
        {formatRupiah(trailingStop.stopPrice)}
      </p>
    </div>
    <div>
      <p class="text-xs text-text-muted">{t("trailingStop.stopPercent")}</p>
      <p class="font-mono text-sm text-text-secondary">-{trailingStop.stopPercentage}%</p>
    </div>
  </div>

  {#if trailingStop.fundamentalExits.some((f) => f.triggered)}
    <div class="mt-4 border-t border-border-default pt-3">
      <div class="mb-2 flex items-center gap-1.5">
        <AlertTriangle size={14} strokeWidth={2} class="text-loss" />
        <p class="text-xs font-semibold text-text-muted">{t("trailingStop.fundamentalWarnings")}</p>
      </div>
      <ul class="space-y-1">
        {#each trailingStop.fundamentalExits as exit}
          <li class="flex items-center gap-2 text-sm" data-testid="fundamental-exit-{exit.key}">
            {#if exit.triggered}
              <AlertTriangle size={12} strokeWidth={2} class="shrink-0 text-loss" />
              <span class="text-loss">{exit.label}</span>
              <span class="font-mono text-xs text-text-muted">({exit.detail})</span>
            {:else}
              <Check size={12} strokeWidth={2} class="shrink-0 text-profit" />
              <span class="text-text-secondary">{exit.label}</span>
              <span class="font-mono text-xs text-text-muted">({exit.detail})</span>
            {/if}
          </li>
        {/each}
      </ul>
    </div>
  {/if}

  <p class="mt-3 text-xs text-text-muted">
    {t("trailingStop.disclaimer")}
  </p>
</div>
