<script lang="ts">
import { t } from "../../i18n";
import Badge from "../../lib/components/Badge.svelte";
import { formatDecimal } from "../../lib/format";
import type { MarketStatusResponse } from "../../lib/types";

let { market }: { market: MarketStatusResponse } = $props();

const conditionConfig: Record<
  string,
  { variant: "value" | "dividend" | "profit" | "loss" | "warning"; label: string }
> = {
  NORMAL: { variant: "value", label: t("marketStatus.normal") },
  ELEVATED: { variant: "warning", label: t("marketStatus.elevated") },
  CORRECTION: { variant: "loss", label: t("marketStatus.correction") },
  CRASH: { variant: "loss", label: t("marketStatus.crash") },
  RECOVERY: { variant: "profit", label: t("marketStatus.recovery") },
};

const config = $derived(conditionConfig[market.condition] ?? conditionConfig.NORMAL);
</script>

<section class="rounded-lg border border-border-default bg-bg-elevated p-4" aria-label="Market status">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <h3 class="text-sm font-medium text-text-secondary">{t("marketStatus.title")}</h3>
      <Badge variant={config.variant}>{config.label}</Badge>
    </div>
    <div class="flex items-center gap-4 text-sm">
      <span class="font-mono font-medium text-text-primary">{formatDecimal(market.ihsgPrice)}</span>
      <span class="font-mono {market.drawdownPct < 0 ? 'text-loss' : 'text-profit'}">
        {formatDecimal(market.drawdownPct)}%
      </span>
    </div>
  </div>
  <div class="mt-2 flex items-center justify-between text-xs text-text-secondary">
    <span>{t("marketStatus.peak")} <span class="font-mono">{formatDecimal(market.ihsgPeak)}</span></span>
    <span>{t("marketStatus.lastFetched")} {new Date(market.fetchedAt).toLocaleString("id-ID")}</span>
  </div>
</section>
