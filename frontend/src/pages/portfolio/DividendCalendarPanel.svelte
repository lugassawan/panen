<script lang="ts">
import { Calendar } from "lucide-svelte";
import { GetDividendCalendar } from "../../../wailsjs/go/backend/App";
import EmptyState from "../../lib/components/EmptyState.svelte";
import { formatRupiah } from "../../lib/format";
import type { DividendCalendarEntryResponse } from "../../lib/types";

interface Props {
  portfolioId: string;
}

let { portfolioId }: Props = $props();

let loading = $state(true);
let error = $state<string | null>(null);
let entries = $state<DividendCalendarEntryResponse[]>([]);

const MONTH_NAMES = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

type MonthGroup = {
  month: string;
  year: number;
  items: DividendCalendarEntryResponse[];
};

let grouped: MonthGroup[] = $derived.by(() => {
  const groups = new Map<string, DividendCalendarEntryResponse[]>();
  for (const entry of entries) {
    const d = new Date(entry.exDate);
    const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
    const list = groups.get(key) ?? [];
    list.push(entry);
    groups.set(key, list);
  }

  return Array.from(groups.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, items]) => {
      const [yearStr, monthStr] = key.split("-");
      return {
        month: MONTH_NAMES[Number(monthStr) - 1],
        year: Number(yearStr),
        items: items.sort((a, b) => a.exDate.localeCompare(b.exDate)),
      };
    });
});

$effect(() => {
  if (!portfolioId) return;
  loading = true;
  error = null;
  GetDividendCalendar(portfolioId)
    .then((result) => {
      entries = result ?? [];
    })
    .catch((e) => {
      error = e instanceof Error ? e.message : String(e);
    })
    .finally(() => {
      loading = false;
    });
});
</script>

<div data-testid="dividend-calendar-panel" class="rounded border border-border-default bg-bg-elevated p-4">
  <p class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Upcoming Dividends</p>

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <p class="text-sm text-text-muted">Loading calendar…</p>
    </div>
  {:else if error}
    <div class="rounded border border-border-default bg-bg-elevated p-6 text-center">
      <p class="text-sm text-loss">{error}</p>
    </div>
  {:else if entries.length === 0}
    <EmptyState icon={Calendar} title="No upcoming dividends" description="No projected dividend events for the next 12 months." />
  {:else}
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {#each grouped as group}
        <div class="rounded border border-border-default bg-bg-secondary p-3">
          <p class="mb-2 text-sm font-semibold text-text-primary">
            {group.month} {group.year}
          </p>
          <div class="space-y-1.5">
            {#each group.items as entry}
              <div class="flex items-center justify-between text-sm {entry.isProjection ? 'opacity-60' : ''}">
                <span class="font-mono font-medium {entry.isProjection ? 'border-b border-dashed border-text-muted' : ''}">
                  {entry.ticker}
                </span>
                <span class="font-mono text-text-secondary">
                  {formatRupiah(entry.amount)}
                </span>
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
