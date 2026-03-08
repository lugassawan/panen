<script lang="ts">
import { Receipt } from "lucide-svelte";
import { ListPortfolios, ListTransactions } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import EmptyState from "../../lib/components/EmptyState.svelte";
import Input from "../../lib/components/Input.svelte";
import Select from "../../lib/components/Select.svelte";
import SkeletonCard from "../../lib/components/SkeletonCard.svelte";
import SkeletonTable from "../../lib/components/SkeletonTable.svelte";
import SortableHeader from "../../lib/components/SortableHeader.svelte";
import { formatDate, formatRupiah } from "../../lib/format";
import { mode } from "../../lib/stores/mode.svelte";
import type {
  PortfolioResponse,
  TransactionListResponse,
  TransactionRecordResponse,
  TransactionType,
} from "../../lib/types";

type FilterType = "ALL" | TransactionType;

let loading = $state(true);
let data = $state<TransactionListResponse | null>(null);
let portfolios = $state<PortfolioResponse[]>([]);

let filterPortfolio = $state("");
let filterTicker = $state("");
let filterType = $state<FilterType>("ALL");
let filterDateFrom = $state("");
let filterDateTo = $state("");

let sortField = $state("date");
let sortAsc = $state(false);

const typeFilters: { key: FilterType; labelKey: string }[] = [
  { key: "ALL", labelKey: "transactions.filterType" },
  { key: "BUY", labelKey: "transactions.buy" },
  { key: "SELL", labelKey: "transactions.sell" },
  { key: "DIVIDEND", labelKey: "transactions.dividend" },
];

function handleSort(field: string) {
  if (sortField === field) {
    sortAsc = !sortAsc;
  } else {
    sortField = field;
    sortAsc = field === "ticker" || field === "portfolio";
  }
}

function typeBadgeClass(type: TransactionType): string {
  switch (type) {
    case "BUY":
      return "bg-green-100 text-green-700";
    case "SELL":
      return "bg-negative-bg text-negative";
    case "DIVIDEND":
      return "bg-gold-100 text-gold-700";
  }
}

function typeLabel(type: TransactionType): string {
  switch (type) {
    case "BUY":
      return t("transactions.buy");
    case "SELL":
      return t("transactions.sell");
    case "DIVIDEND":
      return t("transactions.dividend");
  }
}

async function loadData() {
  loading = true;
  try {
    const txnType = filterType === "ALL" ? "" : filterType;
    data = await ListTransactions(
      filterPortfolio,
      filterTicker.trim(),
      txnType,
      filterDateFrom,
      filterDateTo,
      sortField,
      sortAsc,
    );
  } catch {
    data = null;
  } finally {
    loading = false;
  }
}

$effect(() => {
  ListPortfolios()
    .then((result: PortfolioResponse[]) => {
      portfolios = result ?? [];
    })
    .catch(() => {
      portfolios = [];
    });
});

$effect(() => {
  void filterPortfolio;
  void filterTicker;
  void filterType;
  void filterDateFrom;
  void filterDateTo;
  void sortField;
  void sortAsc;
  loadData();
});

const items = $derived<TransactionRecordResponse[]>(data?.items ?? []);
const summary = $derived(data?.summary);
</script>

<div class="mx-auto max-w-6xl p-6">
  <div class="mb-6">
    <h1 class="text-2xl font-display font-bold text-text-primary">{t("transactions.title")}</h1>
    <p class="mt-1 text-sm text-text-secondary">{t("transactions.subtitle")}</p>
  </div>

  {#if summary && !loading}
    <div class="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
      <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
        <p class="text-xs font-medium uppercase tracking-wider text-text-muted">{t("transactions.totalInvested")}</p>
        <p class="mt-1 text-lg font-bold font-mono text-text-primary">{formatRupiah(summary.totalBuyAmount)}</p>
      </div>
      <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
        <p class="text-xs font-medium uppercase tracking-wider text-text-muted">{t("transactions.totalDividends")}</p>
        <p class="mt-1 text-lg font-bold font-mono text-profit">{formatRupiah(summary.totalDividendAmount)}</p>
      </div>
      <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
        <p class="text-xs font-medium uppercase tracking-wider text-text-muted">{t("transactions.totalFees")}</p>
        <p class="mt-1 text-lg font-bold font-mono text-text-primary">{formatRupiah(summary.totalFees)}</p>
      </div>
    </div>
  {/if}

  <div class="mb-4 flex flex-wrap items-end gap-3">
    <div class="w-44">
      <Select bind:value={filterPortfolio} aria-label={t("transactions.filterPortfolio")}>
        <option value="">{t("transactions.filterPortfolio")}</option>
        {#each portfolios as p}
          <option value={p.id}>{p.name}</option>
        {/each}
      </Select>
    </div>

    <div class="w-40">
      <Input
        type="text"
        placeholder={t("transactions.filterTicker")}
        bind:value={filterTicker}
        aria-label={t("transactions.filterTicker")}
      />
    </div>

    <div class="flex gap-1">
      {#each typeFilters as tf}
        <button
          onclick={() => (filterType = tf.key)}
          class="rounded-md px-3 py-2 text-sm font-medium transition-fast focus-ring
            {filterType === tf.key
            ? mode.config.activeHighlight
            : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
        >
          {t(tf.labelKey)}
        </button>
      {/each}
    </div>

    <div class="flex items-center gap-2">
      <Input
        type="date"
        bind:value={filterDateFrom}
        aria-label={t("transactions.filterDateFrom")}
        class="w-36"
      />
      <span class="text-text-muted text-sm">&ndash;</span>
      <Input
        type="date"
        bind:value={filterDateTo}
        aria-label={t("transactions.filterDateTo")}
        class="w-36"
      />
    </div>
  </div>

  {#if loading}
    <div class="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
      {#each Array(3) as _}
        <SkeletonCard lines={2} />
      {/each}
    </div>
    <SkeletonTable rows={5} columns={8} />
  {:else if items.length === 0}
    <EmptyState icon={Receipt} title={t("transactions.empty")} description={t("transactions.emptyDesc")} />
  {:else}
    <div class="overflow-x-auto rounded-lg border border-border-default">
      <table class="w-full text-sm">
        <thead class="border-b border-border-default bg-bg-secondary">
          <tr>
            <SortableHeader label={t("transactions.date")} field="date" currentSort={sortField} ascending={sortAsc} onclick={handleSort} />
            <SortableHeader label={t("transactions.type")} field="type" currentSort={sortField} ascending={sortAsc} onclick={handleSort} />
            <SortableHeader label={t("transactions.ticker")} field="ticker" currentSort={sortField} ascending={sortAsc} onclick={handleSort} />
            <SortableHeader label={t("transactions.portfolio")} field="portfolio" currentSort={sortField} ascending={sortAsc} onclick={handleSort} />
            <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-text-muted">{t("transactions.lots")}</th>
            <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-text-muted">{t("transactions.price")}</th>
            <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-text-muted">{t("transactions.fee")}</th>
            <SortableHeader label={t("transactions.total")} field="total" currentSort={sortField} ascending={sortAsc} onclick={handleSort} />
          </tr>
        </thead>
        <tbody class="divide-y divide-border-default">
          {#each items as txn (txn.id)}
            <tr class="hover:bg-bg-tertiary transition-fast">
              <td class="whitespace-nowrap px-4 py-3 font-mono text-text-secondary">{formatDate(txn.date)}</td>
              <td class="whitespace-nowrap px-4 py-3">
                <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {typeBadgeClass(txn.type)}">
                  {typeLabel(txn.type)}
                </span>
              </td>
              <td class="whitespace-nowrap px-4 py-3 font-display font-semibold text-text-primary">{txn.ticker}</td>
              <td class="whitespace-nowrap px-4 py-3 text-text-secondary">{txn.portfolioName}</td>
              <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-text-primary">{txn.lots}</td>
              <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-text-primary">{formatRupiah(txn.price)}</td>
              <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-text-secondary">{formatRupiah(txn.fee + txn.tax)}</td>
              <td class="whitespace-nowrap px-4 py-3 text-right font-mono font-semibold text-text-primary">{formatRupiah(txn.total)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    {#if summary}
      <div class="mt-3 text-right text-xs text-text-muted">
        {t("transactions.transactionCount")}: <span class="font-mono">{summary.transactionCount}</span>
      </div>
    {/if}
  {/if}
</div>
