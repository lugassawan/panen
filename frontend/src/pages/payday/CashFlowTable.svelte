<script lang="ts">
import { GetCashFlowSummary } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import EmptyState from "../../lib/components/EmptyState.svelte";
import { formatRupiah } from "../../lib/format";
import type { CashFlowSummaryResponse } from "../../lib/types";

let { portfolioId }: { portfolioId: string } = $props();

let summary = $state<CashFlowSummaryResponse | null>(null);
let error = $state<string | null>(null);
let loading = $state(true);

async function load() {
  try {
    loading = true;
    error = null;
    summary = await GetCashFlowSummary(portfolioId);
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  } finally {
    loading = false;
  }
}

$effect(() => {
  load();
});
</script>

{#if loading}
  <div class="py-4 text-center text-sm text-text-secondary">{t("cashFlow.loading")}</div>
{:else if error}
  <div class="py-4 text-center text-sm text-negative">{error}</div>
{:else if summary}
  <div class="mt-4">
    <div class="flex items-center gap-6 rounded-lg border border-border-default bg-bg-elevated p-4">
      <div>
        <span class="text-xs text-text-secondary">{t("cashFlow.totalInflow")}</span>
        <p class="font-mono text-sm font-medium text-text-primary">{formatRupiah(summary.totalInflow)}</p>
      </div>
      <div>
        <span class="text-xs text-text-secondary">{t("cashFlow.balance")}</span>
        <p class="font-mono text-sm font-medium text-text-primary">{formatRupiah(summary.balance)}</p>
      </div>
    </div>

    {#if summary.items && summary.items.length > 0}
      <div class="mt-3 overflow-hidden rounded-lg border border-border-default">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border-default bg-bg-secondary">
              <th class="px-4 py-2 text-left font-medium text-text-secondary">{t("cashFlow.dateHeader")}</th>
              <th class="px-4 py-2 text-left font-medium text-text-secondary">{t("cashFlow.type")}</th>
              <th class="px-4 py-2 text-right font-medium text-text-secondary">{t("cashFlow.amount")}</th>
              <th class="px-4 py-2 text-left font-medium text-text-secondary">{t("cashFlow.note")}</th>
            </tr>
          </thead>
          <tbody>
            {#each summary.items as item}
              <tr class="border-b border-border-default last:border-b-0">
                <td class="px-4 py-2 text-text-primary">{item.date}</td>
                <td class="px-4 py-2 text-text-secondary">{item.type}</td>
                <td class="px-4 py-2 text-right font-mono text-text-primary">{formatRupiah(item.amount)}</td>
                <td class="px-4 py-2 text-text-secondary">{item.note || "-"}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <EmptyState compact title={t("cashFlow.noRecords")} />
    {/if}
  </div>
{/if}
