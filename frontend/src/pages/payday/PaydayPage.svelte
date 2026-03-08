<script lang="ts">
import {
  ConfirmPayday,
  DeferPayday,
  GetCurrentMonthStatus,
  GetPaydayDay,
  SavePaydayDay,
  SkipPayday,
} from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Alert from "../../lib/components/Alert.svelte";
import Badge from "../../lib/components/Badge.svelte";
import Button from "../../lib/components/Button.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import LoadingState from "../../lib/components/LoadingState.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import { formatRupiah } from "../../lib/format";
import { toastStore } from "../../lib/stores/toast.svelte";
import type {
  MonthlyPaydayResponse,
  PaydayStatus,
  PortfolioPaydayItemResponse,
} from "../../lib/types";
import CashFlowTable from "./CashFlowTable.svelte";
import ConfirmDialog from "./ConfirmDialog.svelte";
import PaydaySetup from "./PaydaySetup.svelte";

type PageState = "loading" | "setup" | "dashboard" | "error";

let state = $state<PageState>("loading");
let error = $state<string | null>(null);
let monthStatus = $state<MonthlyPaydayResponse | null>(null);
let confirmingPortfolio = $state<PortfolioPaydayItemResponse | null>(null);
let deferringPortfolio = $state<PortfolioPaydayItemResponse | null>(null);
let deferDate = $state<string>("");
let expandedCashFlow = $state<string | null>(null);

function statusBadgeVariant(
  status: PaydayStatus,
): "value" | "dividend" | "profit" | "loss" | "warning" {
  switch (status) {
    case "CONFIRMED":
      return "profit";
    case "PENDING":
      return "value";
    case "DEFERRED":
      return "warning";
    case "SKIPPED":
      return "loss";
    default:
      return "dividend";
  }
}

function modeLabel(mode: string): string {
  return mode === "VALUE" ? t("payday.warChest") : t("payday.dcaFund");
}

async function load() {
  try {
    state = "loading";
    error = null;
    const day = await GetPaydayDay();
    if (day === 0) {
      state = "setup";
      return;
    }
    monthStatus = await GetCurrentMonthStatus();
    state = "dashboard";
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

async function handleSaveDay(day: number) {
  try {
    await SavePaydayDay(day);
    await load();
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

async function handleConfirm(portfolioId: string, amount: number) {
  try {
    await ConfirmPayday(portfolioId, amount);
    toastStore.add(t("common.paydayConfirmed"), "success");
    confirmingPortfolio = null;
    await load();
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

function openDeferDialog(portfolio: PortfolioPaydayItemResponse) {
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 7);
  deferDate = tomorrow.toISOString().split("T")[0];
  deferringPortfolio = portfolio;
}

async function handleDefer(portfolioId: string) {
  try {
    await DeferPayday(portfolioId, deferDate);
    toastStore.add(t("common.paydayDeferred"), "success");
    deferringPortfolio = null;
    await load();
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

async function handleSkip(portfolioId: string) {
  try {
    await SkipPayday(portfolioId);
    toastStore.add(t("common.paydaySkipped"), "info");
    await load();
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

function toggleCashFlow(portfolioId: string) {
  expandedCashFlow = expandedCashFlow === portfolioId ? null : portfolioId;
}

$effect(() => {
  load();
});
</script>

<div class="p-6">
  <h1 class="text-2xl font-display font-bold text-text-primary">{t("payday.title")}</h1>
  <p class="mt-1 text-sm text-text-secondary">{t("payday.subtitle")}</p>

  {#if state === "loading"}
    <LoadingState message={t("payday.loading")} class="py-16" />
  {:else if state === "error"}
    <div class="mt-6">
      <Alert variant="negative">{error}</Alert>
      <div class="mt-3">
        <Button variant="secondary" size="sm" onclick={load}>{t("common.retry")}</Button>
      </div>
    </div>
  {:else if state === "setup"}
    <PaydaySetup onSave={handleSaveDay} />
  {:else if state === "dashboard" && monthStatus}
    <div class="mt-6">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-text-primary font-display">{monthStatus.month}</h2>
        <span class="font-mono text-sm text-text-secondary">
          {t("payday.totalExpected")} {formatRupiah(monthStatus.totalExpected)}
        </span>
      </div>

      <div class="mt-4 grid gap-4">
        {#each monthStatus.portfolios as portfolio}
          <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <h3 class="font-medium text-text-primary">{portfolio.portfolioName}</h3>
                <Badge variant={portfolio.mode === "VALUE" ? "value" : "dividend"}>
                  {portfolio.mode}
                </Badge>
                <Badge variant={statusBadgeVariant(portfolio.status)}>
                  {portfolio.status}
                </Badge>
              </div>
              <span class="font-mono text-sm font-medium text-text-primary">
                {formatRupiah(portfolio.expected)}
              </span>
            </div>

            <div class="mt-2 flex items-center justify-between">
              <Tooltip text={portfolio.mode === "VALUE" ? t("payday.warChestTooltip") : t("payday.dcaFundTooltip")}>
                <span class="text-xs text-text-secondary underline decoration-dotted cursor-help">{modeLabel(portfolio.mode)}</span>
              </Tooltip>
              {#if portfolio.status === "DEFERRED" && portfolio.deferUntil}
                <span class="text-xs text-warning">{t("payday.deferredUntil", { date: portfolio.deferUntil })}</span>
              {/if}
            </div>

            {#if portfolio.status === "PENDING"}
              <div class="mt-3 flex items-center gap-2">
                <Button variant="primary" size="sm" onclick={() => { confirmingPortfolio = portfolio; }}>
                  {t("common.confirm")}
                </Button>
                <Button variant="secondary" size="sm" onclick={() => openDeferDialog(portfolio)}>
                  {t("payday.defer")}
                </Button>
                <Button variant="danger" size="sm" onclick={() => handleSkip(portfolio.portfolioId)}>
                  {t("payday.skip")}
                </Button>
              </div>
            {/if}

            <div class="mt-3">
              <button
                class="text-xs text-text-secondary underline transition-fast hover:text-text-primary focus-ring rounded"
                onclick={() => toggleCashFlow(portfolio.portfolioId)}
              >
                {expandedCashFlow === portfolio.portfolioId ? t("common.hide") : t("common.show")} {t("payday.cashFlows")}
              </button>
              {#if expandedCashFlow === portfolio.portfolioId}
                <CashFlowTable portfolioId={portfolio.portfolioId} />
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>
  {:else if state === "dashboard"}
    <EmptyState title={t("payday.noPortfolios")} />
  {/if}
</div>

{#if confirmingPortfolio}
  <ConfirmDialog
    expected={confirmingPortfolio.expected}
    portfolioName={confirmingPortfolio.portfolioName}
    onConfirm={(amount) => handleConfirm(confirmingPortfolio!.portfolioId, amount)}
    onCancel={() => { confirmingPortfolio = null; }}
  />
{/if}

{#if deferringPortfolio}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="fixed inset-0" role="presentation" onclick={() => { deferringPortfolio = null; }}></div>
    <div
      class="relative z-10 w-full max-w-md rounded-lg border border-border-default bg-bg-elevated p-6"
      role="dialog"
      aria-modal="true"
      aria-labelledby="defer-dialog-title"
      tabindex="-1"
      onkeydown={(e) => { if (e.key === "Escape") deferringPortfolio = null; }}
    >
      <h3 id="defer-dialog-title" class="text-lg font-semibold text-text-primary font-display">{t("payday.defer")} {t("payday.title")}</h3>
      <p class="mt-1 text-sm text-text-secondary">
        {t("common.chooseDeferDate", { name: deferringPortfolio.portfolioName })}
      </p>
      <label class="mt-4 block text-sm font-medium text-text-secondary">
        {t("common.deferDate")}
        <input
          type="date"
          class="mt-1 block w-full rounded-md border border-border-default bg-bg-primary px-3 py-2 text-sm text-text-primary focus-ring"
          bind:value={deferDate}
          min={new Date(Date.now() + 86400000).toISOString().split("T")[0]}
        />
      </label>
      <div class="mt-4 flex items-center justify-end gap-3">
        <Button variant="secondary" size="sm" onclick={() => { deferringPortfolio = null; }}>{t("common.cancel")}</Button>
        <Button variant="primary" size="sm" onclick={() => handleDefer(deferringPortfolio!.portfolioId)}>
          {t("payday.defer")}
        </Button>
      </div>
    </div>
  </div>
{/if}
