<script lang="ts">
import { LoaderCircle } from "lucide-svelte";
import { GetPortfolio, ListBrokerageAccounts, ListPortfolios } from "../../wailsjs/go/backend/App";
import AddHoldingForm from "../components/AddHoldingForm.svelte";
import BrokerageAccountForm from "../components/BrokerageAccountForm.svelte";
import PortfolioForm from "../components/PortfolioForm.svelte";
import { formatPercent, formatRupiah } from "../lib/format";
import type {
  BrokerageAccountResponse,
  HoldingDetailResponse,
  PortfolioDetailResponse,
} from "../lib/types";
import { getVerdictDisplay } from "../lib/verdict";

type PageState = "loading" | "onboarding" | "create-portfolio" | "view" | "error";

let state = $state<PageState>("loading");
let error = $state<string | null>(null);
let brokerageAcctId = $state<string | null>(null);
let detail = $state<PortfolioDetailResponse | null>(null);
let onboardingStep = $state<1 | 2>(1);

async function load() {
  state = "loading";
  error = null;

  try {
    const accounts: BrokerageAccountResponse[] = await ListBrokerageAccounts();
    if (!accounts || accounts.length === 0) {
      state = "onboarding";
      onboardingStep = 1;
      return;
    }

    brokerageAcctId = accounts[0].id;
    const portfolios = await ListPortfolios(brokerageAcctId);
    if (!portfolios || portfolios.length === 0) {
      state = "create-portfolio";
      return;
    }

    detail = await GetPortfolio(portfolios[0].id);
    state = "view";
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

function getSignal(h: HoldingDetailResponse): string {
  if (
    h.verdict === "OVERVALUED" &&
    h.currentPrice != null &&
    h.exitTarget != null &&
    h.currentPrice > h.exitTarget
  ) {
    return "Consider Selling";
  }
  if (h.verdict === "UNDERVALUED") {
    return "Hold / Add";
  }
  return "Hold";
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
            onCreated={(acct) => {
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
            onCreated={() => load()}
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
          onCreated={() => load()}
        />
      </div>
    </div>
  {:else if state === "view" && detail}
    <h2 class="mb-6 text-xl font-semibold text-text-primary">{detail.portfolio.name}</h2>

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
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Signal</th>
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
              <td class="px-4 py-3 text-sm">
                {#if holding.verdict}
                  {getSignal(holding)}
                {:else}
                  <span class="text-text-muted">&mdash;</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Add Holding -->
    <div class="rounded border border-border-default bg-bg-elevated p-4">
      <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Add Holding</h3>
      <AddHoldingForm portfolioId={detail.portfolio.id} onAdded={() => load()} />
    </div>
  {/if}
</div>
