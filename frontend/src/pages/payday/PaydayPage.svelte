<script lang="ts">
import {
  ConfirmPayday,
  DeferPayday,
  GetCurrentMonthStatus,
  GetPaydayDay,
  SavePaydayDay,
  SkipPayday,
} from "../../../wailsjs/go/backend/App";
import Badge from "../../lib/components/Badge.svelte";
import Button from "../../lib/components/Button.svelte";
import { formatRupiah } from "../../lib/format";
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
  return mode === "VALUE" ? "War Chest" : "DCA Fund";
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
    confirmingPortfolio = null;
    await load();
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

async function handleDefer(portfolioId: string) {
  try {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 7);
    const deferDate = tomorrow.toISOString().split("T")[0];
    await DeferPayday(portfolioId, deferDate);
    await load();
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

async function handleSkip(portfolioId: string) {
  try {
    await SkipPayday(portfolioId);
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
  <h1 class="text-2xl font-bold text-text-primary font-display">Payday</h1>
  <p class="mt-1 text-sm text-text-secondary">Track your monthly investment schedule.</p>

  {#if state === "loading"}
    <div class="flex items-center justify-center py-16">
      <p class="text-sm text-text-secondary">Loading...</p>
    </div>
  {:else if state === "error"}
    <div class="mt-6 rounded-lg border border-negative bg-negative-bg p-4">
      <p class="text-sm text-negative">{error}</p>
      <div class="mt-3">
        <Button variant="secondary" size="sm" onclick={load}>Retry</Button>
      </div>
    </div>
  {:else if state === "setup"}
    <PaydaySetup onSave={handleSaveDay} />
  {:else if state === "dashboard" && monthStatus}
    <div class="mt-6">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-text-primary font-display">{monthStatus.month}</h2>
        <span class="font-mono text-sm text-text-secondary">
          Total Expected: {formatRupiah(monthStatus.totalExpected)}
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
              <span class="text-xs text-text-secondary">{modeLabel(portfolio.mode)}</span>
              {#if portfolio.status === "DEFERRED" && portfolio.deferUntil}
                <span class="text-xs text-warning">Deferred until {portfolio.deferUntil}</span>
              {/if}
            </div>

            {#if portfolio.status === "PENDING"}
              <div class="mt-3 flex items-center gap-2">
                <Button variant="primary" size="sm" onclick={() => { confirmingPortfolio = portfolio; }}>
                  Confirm
                </Button>
                <Button variant="secondary" size="sm" onclick={() => handleDefer(portfolio.portfolioId)}>
                  Defer
                </Button>
                <Button variant="danger" size="sm" onclick={() => handleSkip(portfolio.portfolioId)}>
                  Skip
                </Button>
              </div>
            {/if}

            <div class="mt-3">
              <button
                class="text-xs text-text-secondary underline transition-fast hover:text-text-primary focus-ring rounded"
                onclick={() => toggleCashFlow(portfolio.portfolioId)}
              >
                {expandedCashFlow === portfolio.portfolioId ? "Hide" : "Show"} Cash Flows
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
    <div class="mt-6 text-center text-sm text-text-secondary">
      No portfolios found. Create a portfolio first to track your payday schedule.
    </div>
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
