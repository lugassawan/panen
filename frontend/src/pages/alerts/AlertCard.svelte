<script lang="ts">
import { t } from "../../i18n";
import Badge from "../../lib/components/Badge.svelte";
import Button from "../../lib/components/Button.svelte";
import { formatDecimal, formatRelativeTime } from "../../lib/format";
import type { AlertSeverity, FundamentalAlertResponse } from "../../lib/types";

let {
  alert,
  onAcknowledge,
}: {
  alert: FundamentalAlertResponse;
  onAcknowledge?: (id: string) => void;
} = $props();

const severityConfig: Record<
  AlertSeverity,
  { border: string; badge: "loss" | "warning" | "value" }
> = {
  CRITICAL: { border: "border-l-negative", badge: "loss" },
  WARNING: { border: "border-l-warning", badge: "warning" },
  MINOR: { border: "border-l-info", badge: "value" },
};

const metricLabels: Record<string, string> = {
  roe: "ROE",
  der: "DER",
  eps: "EPS",
  pbv: "PBV",
  per: "PER",
  dividend_yield: "Dividend Yield",
  payout_ratio: "Payout Ratio",
};

function formatMetricValue(metric: string, value: number): string {
  if (metric === "der" || metric === "pbv" || metric === "per") {
    return formatDecimal(value);
  }
  return `${formatDecimal(value)}%`;
}

const config = $derived(severityConfig[alert.severity]);
const metricLabel = $derived(metricLabels[alert.metric] ?? alert.metric);
const canAcknowledge = $derived(alert.status === "ACTIVE" && alert.severity === "CRITICAL");
</script>

<div
  class="flex items-center gap-4 rounded-lg border border-border-default bg-bg-elevated px-4 py-3 border-l-4 {config.border}"
>
  <div class="flex-1 min-w-0">
    <div class="flex items-center gap-2 mb-1">
      <span class="font-display font-bold text-text-primary">{alert.ticker}</span>
      <Badge variant={config.badge}>{alert.severity}</Badge>
      {#if alert.status === "ACKNOWLEDGED"}
        <Badge variant="value">{t("alerts.acknowledged")}</Badge>
      {/if}
    </div>
    <p class="text-sm text-text-primary font-mono">
      {metricLabel}: {formatMetricValue(alert.metric, alert.oldValue)} → {formatMetricValue(alert.metric, alert.newValue)}
      <span class="text-text-secondary">({alert.changePct > 0 ? "+" : ""}{formatDecimal(alert.changePct)}%)</span>
    </p>
    <p class="text-xs text-text-tertiary mt-1">
      {t("alerts.detectedAt", { date: formatRelativeTime(alert.detectedAt) })}
    </p>
  </div>

  {#if canAcknowledge && onAcknowledge}
    <Button variant="secondary" size="sm" onclick={() => onAcknowledge(alert.id)}>
      {t("alerts.acknowledge")}
    </Button>
  {/if}
</div>
