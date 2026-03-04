<script lang="ts">
import { ArrowLeft, LoaderCircle, Pencil, Plus, Trash2 } from "lucide-svelte";
import {
  DeletePortfolio,
  GetPortfolio,
  ListBrokerageAccounts,
  ListBrokerConfigs,
  ListPortfolios,
} from "../../../wailsjs/go/backend/App";
import BrokerageAccountForm from "../../components/BrokerageAccountForm.svelte";
import ConfirmDialog from "../../components/ConfirmDialog.svelte";
import Button from "../../lib/components/Button.svelte";
import { formatPercent, formatRupiah } from "../../lib/format";
import type {
  ActionType,
  BrokerageAccountResponse,
  BrokerConfigResponse,
  HoldingDetailResponse,
  PortfolioDetailResponse,
  PortfolioResponse,
} from "../../lib/types";
import { getVerdictDisplay } from "../../lib/verdict";
import ActionSelector from "./ActionSelector.svelte";
import AddHoldingForm from "./AddHoldingForm.svelte";
import ChecklistPanel from "./ChecklistPanel.svelte";
import PortfolioForm from "./PortfolioForm.svelte";

type PageState =
  | "loading"
  | "onboarding"
  | "create-portfolio"
  | "list"
  | "view"
  | "edit-portfolio"
  | "checklist"
  | "error";

let state = $state<PageState>("loading");
let error = $state<string | null>(null);
let brokerageAcctId = $state<string | null>(null);
let detail = $state<PortfolioDetailResponse | null>(null);
let onboardingStep = $state<1 | 2>(1);
let brokerConfigs = $state<BrokerConfigResponse[]>([]);
let portfolios = $state<PortfolioResponse[]>([]);
let editingPortfolio = $state<PortfolioResponse | null>(null);
let deletingPortfolio = $state<PortfolioResponse | null>(null);
let deleteLoading = $state(false);
let deleteError = $state<string | null>(null);
let checklistTicker = $state<string | null>(null);
let checklistAction = $state<ActionType | null>(null);

const MODE_BADGE: Record<string, string> = {
  VALUE: "bg-green-100 text-green-700",
  DIVIDEND: "bg-gold-100 text-gold-700",
};

async function load() {
  state = "loading";
  error = null;

  try {
    const [accounts, configs]: [BrokerageAccountResponse[], BrokerConfigResponse[]] =
      await Promise.all([ListBrokerageAccounts(), ListBrokerConfigs()]);
    brokerConfigs = configs ?? [];
    if (!accounts || accounts.length === 0) {
      state = "onboarding";
      onboardingStep = 1;
      return;
    }

    brokerageAcctId = accounts[0].id;
    const result = await ListPortfolios(brokerageAcctId);
    portfolios = result ?? [];
    if (portfolios.length === 0) {
      state = "create-portfolio";
      return;
    }

    state = "list";
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

async function viewPortfolio(portfolio: PortfolioResponse) {
  state = "loading";
  try {
    detail = await GetPortfolio(portfolio.id);
    state = "view";
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

function startEdit(portfolio: PortfolioResponse) {
  editingPortfolio = portfolio;
  state = "edit-portfolio";
}

function startDelete(portfolio: PortfolioResponse) {
  deletingPortfolio = portfolio;
  deleteError = null;
}

async function confirmDelete() {
  if (!deletingPortfolio) return;
  deleteLoading = true;
  deleteError = null;
  try {
    await DeletePortfolio(deletingPortfolio.id);
    deletingPortfolio = null;
    await load();
  } catch (e: unknown) {
    deleteError = e instanceof Error ? e.message : String(e);
  } finally {
    deleteLoading = false;
  }
}

function cancelDelete() {
  deletingPortfolio = null;
  deleteError = null;
}

function calcPL(h: HoldingDetailResponse): number | null {
  if (h.currentPrice == null) return null;
  return ((h.currentPrice - h.avgBuyPrice) / h.avgBuyPrice) * 100;
}

let totalInvested = $derived(
  detail ? detail.holdings.reduce((sum, h) => sum + h.avgBuyPrice * h.lots * 100, 0) : 0,
);

let currentValue = $derived(
  detail
    ? detail.holdings.reduce((sum, h) => sum + (h.currentPrice ?? h.avgBuyPrice) * h.lots * 100, 0)
    : 0,
);

let overallPL = $derived(
  totalInvested > 0 ? ((currentValue - totalInvested) / totalInvested) * 100 : 0,
);

load();
</script>

<div class="mx-auto max-w-4xl px-4 py-8">
  {#if state === "loading"}
    <div class="flex items-center justify-center gap-2 py-12 text-text-secondary" role="status">
      <LoaderCircle size={20} strokeWidth={2} class="animate-spin" />
      <span>Loading portfolio…</span>
    </div>
  {:else if state === "error"}
    <div class="rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
      {error}
    </div>
  {:else if state === "onboarding"}
    <div class="mx-auto max-w-lg">
      {#if onboardingStep === 1}
        <h2 class="mb-6 text-xl font-semibold text-text-primary">Set Up Your Brokerage</h2>
        <div class="rounded border border-border-default bg-bg-elevated p-6">
          <BrokerageAccountForm
            {brokerConfigs}
            onSaved={(acct) => {
              brokerageAcctId = acct.id;
              onboardingStep = 2;
            }}
          />
        </div>
      {:else}
        <h2 class="mb-6 text-xl font-semibold text-text-primary">Create Your Portfolio</h2>
        <div class="rounded border border-border-default bg-bg-elevated p-6">
          <PortfolioForm
            brokerageAcctId={brokerageAcctId ?? ""}
            onSaved={() => load()}
          />
        </div>
      {/if}
    </div>
  {:else if state === "create-portfolio"}
    <div class="mx-auto max-w-lg">
      <h2 class="mb-6 text-xl font-semibold text-text-primary">Create Your Portfolio</h2>
      <div class="rounded border border-border-default bg-bg-elevated p-6">
        <PortfolioForm
          brokerageAcctId={brokerageAcctId ?? ""}
          onSaved={() => load()}
        />
      </div>
    </div>
  {:else if state === "list"}
    <div class="mb-6 flex items-center justify-between">
      <h2 class="text-xl font-semibold text-text-primary">Portfolios</h2>
      {#if portfolios.length < 2}
        <Button onclick={() => { state = "create-portfolio"; }}>
          <Plus size={16} strokeWidth={2} />
          New Portfolio
        </Button>
      {/if}
    </div>

    <div class="grid gap-4">
      {#each portfolios as portfolio}
        <div
          class="flex items-center justify-between rounded border border-border-default bg-bg-elevated p-4"
          data-testid="portfolio-card"
        >
          <button
            type="button"
            class="flex-1 text-left"
            onclick={() => viewPortfolio(portfolio)}
          >
            <div class="flex items-center gap-2">
              <p class="font-medium text-text-primary">{portfolio.name}</p>
              <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {MODE_BADGE[portfolio.mode]}" data-testid="mode-badge">
                {portfolio.mode === "VALUE" ? "Value" : "Dividend"}
              </span>
            </div>
            <div class="mt-1 flex gap-4 text-sm text-text-secondary">
              <span>Risk: {portfolio.riskProfile.charAt(0) + portfolio.riskProfile.slice(1).toLowerCase()}</span>
              <span>Capital: <span class="font-mono">{formatRupiah(portfolio.capital)}</span></span>
            </div>
          </button>
          <div class="flex gap-2">
            <Button variant="ghost" size="sm" onclick={() => startEdit(portfolio)}>
              <Pencil size={14} strokeWidth={2} />
              Edit
            </Button>
            <Button variant="ghost" size="sm" onclick={() => startDelete(portfolio)}>
              <Trash2 size={14} strokeWidth={2} />
              Delete
            </Button>
          </div>
        </div>
      {/each}
    </div>
  {:else if state === "edit-portfolio" && editingPortfolio}
    <div class="mx-auto max-w-lg">
      <h3 class="mb-4 text-lg font-semibold text-text-primary">Edit Portfolio</h3>
      <div class="rounded border border-border-default bg-bg-elevated p-6">
        <PortfolioForm
          existingPortfolio={editingPortfolio}
          onSaved={() => {
            editingPortfolio = null;
            load();
          }}
          onCancel={() => {
            editingPortfolio = null;
            state = "list";
          }}
        />
      </div>
    </div>
  {:else if state === "view" && detail}
    <div class="mb-6 flex items-center gap-3">
      <button
        type="button"
        class="rounded p-1 text-text-secondary hover:bg-bg-tertiary hover:text-text-primary focus-ring transition-fast"
        onclick={() => load()}
        aria-label="Back to list"
      >
        <ArrowLeft size={20} strokeWidth={2} />
      </button>
      <h2 class="text-xl font-semibold text-text-primary">{detail.portfolio.name}</h2>
      <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {MODE_BADGE[detail.portfolio.mode]}">
        {detail.portfolio.mode === "VALUE" ? "Value" : "Dividend"}
      </span>
    </div>

    <!-- Summary Bar -->
    <div class="mb-6 grid grid-cols-3 gap-4">
      <div class="rounded border border-border-default bg-bg-elevated p-4" data-testid="total-invested">
        <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Total Invested</p>
        <p class="mt-1 text-lg font-medium">{formatRupiah(totalInvested)}</p>
      </div>
      <div class="rounded border border-border-default bg-bg-elevated p-4" data-testid="current-value">
        <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Current Value</p>
        <p class="mt-1 text-lg font-medium">{formatRupiah(currentValue)}</p>
      </div>
      <div class="rounded border border-border-default bg-bg-elevated p-4" data-testid="overall-pl">
        <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Overall P/L</p>
        <p class="mt-1 text-lg font-medium font-mono {overallPL >= 0 ? 'text-profit' : 'text-loss'}">
          {overallPL >= 0 ? "+" : ""}{formatPercent(overallPL)}
        </p>
      </div>
    </div>

    <!-- Holdings Table -->
    <div class="mb-6 overflow-x-auto rounded border border-border-default">
      <table class="w-full text-sm" aria-label="Holdings">
        <thead class="border-b border-border-default bg-bg-secondary">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Ticker</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Avg Buy Price</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Lots</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Current Price</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">P/L %</th>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Verdict</th>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border-default">
          {#each detail.holdings as holding}
            {@const pl = calcPL(holding)}
            {@const verdict = holding.verdict ? getVerdictDisplay(holding.verdict) : null}
            <tr class="hover:bg-bg-tertiary">
              <td class="px-4 py-3 font-medium">{holding.ticker}</td>
              <td class="px-4 py-3 text-right font-mono text-text-secondary">{formatRupiah(holding.avgBuyPrice)}</td>
              <td class="px-4 py-3 text-right text-text-secondary">{holding.lots}</td>
              <td class="px-4 py-3 text-right font-mono text-text-secondary">
                {holding.currentPrice != null ? formatRupiah(holding.currentPrice) : "\u2014"}
              </td>
              <td
                class="px-4 py-3 text-right font-mono {pl != null && pl >= 0 ? 'text-profit' : ''} {pl != null && pl < 0 ? 'text-loss' : ''}"
                data-testid="pl-{holding.ticker}"
              >
                {#if pl != null}
                  {pl >= 0 ? "+" : ""}{formatPercent(pl)}
                {:else}
                  &mdash;
                {/if}
              </td>
              <td class="px-4 py-3">
                {#if verdict}
                  <span class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium {verdict.bgClass} {verdict.colorClass}">
                    <span aria-hidden="true">{verdict.icon}</span>
                    {verdict.label}
                  </span>
                {:else}
                  <span class="text-text-muted">&mdash;</span>
                {/if}
              </td>
              <td class="px-4 py-3">
                <Button
                  variant="ghost"
                  size="sm"
                  onclick={() => {
                    checklistTicker = holding.ticker;
                    checklistAction = null;
                    state = "checklist";
                  }}
                >
                  Checklist
                </Button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Add Holding -->
    <div class="rounded border border-border-default bg-bg-elevated p-4">
      <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Add Holding</h3>
      <AddHoldingForm portfolioId={detail.portfolio.id} onAdded={() => viewPortfolio(detail!.portfolio)} />
    </div>
  {:else if state === "checklist" && detail && checklistTicker}
    <div class="mb-6 flex items-center gap-3">
      <button
        type="button"
        class="rounded p-1 text-text-secondary hover:bg-bg-tertiary hover:text-text-primary focus-ring transition-fast"
        onclick={() => {
          checklistTicker = null;
          checklistAction = null;
          state = "view";
        }}
        aria-label="Back to portfolio"
      >
        <ArrowLeft size={20} strokeWidth={2} />
      </button>
      <h2 class="text-xl font-semibold text-text-primary">
        {checklistTicker} Checklist
      </h2>
    </div>

    <div class="space-y-6">
      <ActionSelector
        portfolioId={detail.portfolio.id}
        ticker={checklistTicker}
        onselect={(action) => { checklistAction = action; }}
      />

      {#if checklistAction}
        <ChecklistPanel
          portfolioId={detail.portfolio.id}
          ticker={checklistTicker}
          action={checklistAction}
        />
      {/if}
    </div>
  {/if}
</div>

{#if deletingPortfolio}
  <ConfirmDialog
    title="Delete Portfolio"
    confirmLabel="Delete"
    confirmVariant="danger"
    loading={deleteLoading}
    onConfirm={confirmDelete}
    onCancel={cancelDelete}
  >
    <p>Are you sure you want to delete <strong>{deletingPortfolio.name}</strong>?</p>
    <p class="mt-1">This action cannot be undone.</p>
    {#if deleteError}
      <div class="mt-3 rounded border border-negative/20 bg-negative-bg px-3 py-2 text-sm text-negative" role="alert">
        {deleteError}
      </div>
    {/if}
  </ConfirmDialog>
{/if}
