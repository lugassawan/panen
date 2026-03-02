<script lang="ts">
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
    h.currentPrice &&
    h.exitTarget &&
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
    ? detail.holdings.reduce(
        (sum, h) => sum + (h.currentPrice != null ? h.currentPrice * h.lots * 100 : 0),
        0,
      )
    : 0,
);

let overallPL = $derived(
  totalInvested > 0 ? ((currentValue - totalInvested) / totalInvested) * 100 : 0,
);

load();
</script>

<div class="mx-auto max-w-4xl px-4 py-8">
  {#if state === "loading"}
    <div class="flex items-center justify-center gap-2 py-12 text-neutral-400" role="status">
      <svg class="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25"></circle>
        <path d="M4 12a8 8 0 018-8" stroke="currentColor" stroke-width="4" stroke-linecap="round" class="opacity-75"></path>
      </svg>
      <span>Loading portfolio…</span>
    </div>
  {:else if state === "error"}
    <div class="rounded border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400" role="alert">
      {error}
    </div>
  {:else if state === "onboarding"}
    <div class="mx-auto max-w-lg">
      {#if onboardingStep === 1}
        <h2 class="mb-6 text-xl font-semibold text-neutral-200">Set Up Your Brokerage</h2>
        <div class="rounded border border-neutral-800 bg-neutral-900 p-6">
          <BrokerageAccountForm
            onCreated={(acct) => {
              brokerageAcctId = acct.id;
              onboardingStep = 2;
            }}
          />
        </div>
      {:else}
        <h2 class="mb-6 text-xl font-semibold text-neutral-200">Create Your Portfolio</h2>
        <div class="rounded border border-neutral-800 bg-neutral-900 p-6">
          <PortfolioForm
            brokerageAcctId={brokerageAcctId ?? ""}
            onCreated={() => load()}
          />
        </div>
      {/if}
    </div>
  {:else if state === "create-portfolio"}
    <div class="mx-auto max-w-lg">
      <h2 class="mb-6 text-xl font-semibold text-neutral-200">Create Your Portfolio</h2>
      <div class="rounded border border-neutral-800 bg-neutral-900 p-6">
        <PortfolioForm
          brokerageAcctId={brokerageAcctId ?? ""}
          onCreated={() => load()}
        />
      </div>
    </div>
  {:else if state === "view" && detail}
    <h2 class="mb-6 text-xl font-semibold text-neutral-200">{detail.portfolio.name}</h2>

    <!-- Summary Bar -->
    <div class="mb-6 grid grid-cols-3 gap-4">
      <div class="rounded border border-neutral-800 bg-neutral-900 p-4" data-testid="total-invested">
        <p class="text-xs font-semibold uppercase tracking-wider text-neutral-500">Total Invested</p>
        <p class="mt-1 text-lg font-medium">{formatRupiah(totalInvested)}</p>
      </div>
      <div class="rounded border border-neutral-800 bg-neutral-900 p-4" data-testid="current-value">
        <p class="text-xs font-semibold uppercase tracking-wider text-neutral-500">Current Value</p>
        <p class="mt-1 text-lg font-medium">{formatRupiah(currentValue)}</p>
      </div>
      <div class="rounded border border-neutral-800 bg-neutral-900 p-4" data-testid="overall-pl">
        <p class="text-xs font-semibold uppercase tracking-wider text-neutral-500">Overall P/L</p>
        <p class="mt-1 text-lg font-medium {overallPL >= 0 ? 'text-emerald-400' : 'text-red-400'}">
          {overallPL >= 0 ? "+" : ""}{formatPercent(overallPL)}
        </p>
      </div>
    </div>

    <!-- Holdings Table -->
    <div class="mb-6 overflow-x-auto rounded border border-neutral-800">
      <table class="w-full text-sm">
        <thead class="border-b border-neutral-800 bg-neutral-900">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-neutral-500">Ticker</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-neutral-500">Avg Buy Price</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-neutral-500">Lots</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-neutral-500">Current Price</th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-neutral-500">P/L %</th>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-neutral-500">Verdict</th>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-neutral-500">Signal</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-neutral-800">
          {#each detail.holdings as holding}
            {@const pl = calcPL(holding)}
            {@const verdict = holding.verdict ? getVerdictDisplay(holding.verdict) : null}
            <tr class="hover:bg-neutral-900/50">
              <td class="px-4 py-3 font-medium">{holding.ticker}</td>
              <td class="px-4 py-3 text-right text-neutral-300">{formatRupiah(holding.avgBuyPrice)}</td>
              <td class="px-4 py-3 text-right text-neutral-300">{holding.lots}</td>
              <td class="px-4 py-3 text-right text-neutral-300">
                {holding.currentPrice != null ? formatRupiah(holding.currentPrice) : "\u2014"}
              </td>
              <td
                class="px-4 py-3 text-right {pl != null && pl >= 0 ? 'text-emerald-400' : ''} {pl != null && pl < 0 ? 'text-red-400' : ''}"
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
                  <span class="text-neutral-500">&mdash;</span>
                {/if}
              </td>
              <td class="px-4 py-3 text-sm">
                {#if holding.verdict}
                  {getSignal(holding)}
                {:else}
                  <span class="text-neutral-500">&mdash;</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Add Holding -->
    <div class="rounded border border-neutral-800 bg-neutral-900 p-4">
      <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-neutral-500">Add Holding</h3>
      <AddHoldingForm portfolioId={detail.portfolio.id} onAdded={() => load()} />
    </div>
  {/if}
</div>
