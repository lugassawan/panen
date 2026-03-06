<script lang="ts">
import { Bell } from "lucide-svelte";
import { t } from "../../i18n";
import { alerts } from "../../lib/stores/alerts.svelte";
import type { AlertSeverity, FundamentalAlertResponse } from "../../lib/types";
import AlertCard from "./AlertCard.svelte";

type FilterTab = "all" | "CRITICAL" | "WARNING" | "MINOR";

let activeFilter = $state<FilterTab>("all");

const filteredAlerts = $derived(
  activeFilter === "all"
    ? alerts.activeAlerts
    : alerts.activeAlerts.filter((a: FundamentalAlertResponse) => a.severity === activeFilter),
);

const severityOrder: Record<AlertSeverity, number> = {
  CRITICAL: 0,
  WARNING: 1,
  MINOR: 2,
};

const sortedAlerts = $derived(
  [...filteredAlerts].sort((a: FundamentalAlertResponse, b: FundamentalAlertResponse) => {
    const sevDiff = severityOrder[a.severity] - severityOrder[b.severity];
    if (sevDiff !== 0) return sevDiff;
    return new Date(b.detectedAt).getTime() - new Date(a.detectedAt).getTime();
  }),
);

const filterTabs: { key: FilterTab; labelKey: string }[] = [
  { key: "all", labelKey: "alerts.filterAll" },
  { key: "CRITICAL", labelKey: "alerts.filterCritical" },
  { key: "WARNING", labelKey: "alerts.filterWarning" },
  { key: "MINOR", labelKey: "alerts.filterMinor" },
];

function handleAcknowledge(id: string) {
  alerts.acknowledgeAlert(id);
}

$effect(() => {
  alerts.loadActiveAlerts();
});
</script>

<div class="mx-auto max-w-4xl p-6">
  <div class="mb-6">
    <div class="flex items-center gap-3">
      <h1 class="text-2xl font-display font-bold text-text-primary">{t("alerts.title")}</h1>
      {#if alerts.activeCount > 0}
        <span class="inline-flex items-center justify-center rounded-full bg-negative px-2 py-0.5 text-xs font-bold text-white">
          {alerts.activeCount}
        </span>
      {/if}
    </div>
  </div>

  <div class="mb-4 flex gap-1">
    {#each filterTabs as tab}
      <button
        onclick={() => (activeFilter = tab.key)}
        class="rounded-md px-3 py-1.5 text-sm font-medium transition-fast focus-ring
          {activeFilter === tab.key
          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
          : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
      >
        {t(tab.labelKey)}
      </button>
    {/each}
  </div>

  {#if alerts.loading}
    <p class="text-sm text-text-secondary">{t("common.loading")}</p>
  {:else if sortedAlerts.length === 0}
    <div class="flex flex-col items-center justify-center py-16 text-center">
      <Bell size={48} strokeWidth={1} class="text-text-tertiary mb-4" />
      <p class="text-text-secondary">{t("alerts.empty")}</p>
    </div>
  {:else}
    <div class="flex flex-col gap-3">
      {#each sortedAlerts as alert (alert.id)}
        <AlertCard {alert} onAcknowledge={handleAcknowledge} />
      {/each}
    </div>
  {/if}
</div>
