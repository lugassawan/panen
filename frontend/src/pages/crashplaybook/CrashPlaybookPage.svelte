<script lang="ts">
import { LoaderCircle } from "lucide-svelte";
import {
  GetCrashCapital,
  GetDeploymentPlan,
  GetDeploymentSettings,
  GetDiagnostic,
  GetPortfolioPlaybook,
  ListAllPortfolios,
  SaveCrashCapital,
  SaveDeploymentSettings,
} from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Button from "../../lib/components/Button.svelte";
import Select from "../../lib/components/Select.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import type {
  CrashCapitalResponse,
  DeploymentPlanResponse,
  DeploymentSettingsResponse,
  DiagnosticResponse,
  PortfolioPlaybookResponse,
  PortfolioResponse,
} from "../../lib/types";
import CrashCapitalPanel from "./CrashCapitalPanel.svelte";
import DeploymentSettings from "./DeploymentSettings.svelte";
import FallingKnifeDialog from "./FallingKnifeDialog.svelte";
import MarketStatusBanner from "./MarketStatusBanner.svelte";
import StockPlaybookCard from "./StockPlaybookCard.svelte";

type PageState = "loading" | "ready" | "empty" | "error";

let state = $state<PageState>("loading");
let error = $state<string | null>(null);
let portfolios = $state<PortfolioResponse[]>([]);
let selectedPortfolioId = $state<string>("");
let playbook = $state<PortfolioPlaybookResponse | null>(null);
let capital = $state<CrashCapitalResponse | null>(null);
let deploymentPlan = $state<DeploymentPlanResponse | null>(null);
let deploymentSettings = $state<DeploymentSettingsResponse | null>(null);

let diagnosticTicker = $state<string | null>(null);
let diagnosticResult = $state<DiagnosticResponse | null>(null);
let showSettings = $state(false);

async function loadPortfolios() {
  try {
    state = "loading";
    error = null;
    const list = await ListAllPortfolios();
    portfolios = list ?? [];
    if (portfolios.length === 0) {
      state = "empty";
      return;
    }
    if (!selectedPortfolioId || !portfolios.find((p) => p.id === selectedPortfolioId)) {
      selectedPortfolioId = portfolios[0].id;
    }
    await loadPlaybook();
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

async function loadPlaybook() {
  try {
    state = "loading";
    error = null;
    const [pb, cc, dp] = await Promise.all([
      GetPortfolioPlaybook(selectedPortfolioId),
      GetCrashCapital(selectedPortfolioId),
      GetDeploymentPlan(selectedPortfolioId),
    ]);
    playbook = pb;
    capital = cc;
    deploymentPlan = dp;
    state = "ready";
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

async function handlePortfolioChange(e: Event & { currentTarget: HTMLSelectElement }) {
  selectedPortfolioId = e.currentTarget.value;
  await loadPlaybook();
}

async function handleSaveCapital(amount: number) {
  try {
    await SaveCrashCapital(selectedPortfolioId, amount);
    capital = await GetCrashCapital(selectedPortfolioId);
    deploymentPlan = await GetDeploymentPlan(selectedPortfolioId);
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

async function handleDiagnostic(ticker: string) {
  try {
    diagnosticTicker = ticker;
    diagnosticResult = await GetDiagnostic(ticker, selectedPortfolioId, null, null);
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

async function handleDiagnosticUpdate(
  companyBadNews: boolean | null,
  fundamentalsOK: boolean | null,
) {
  if (!diagnosticTicker) return;
  try {
    diagnosticResult = await GetDiagnostic(
      diagnosticTicker,
      selectedPortfolioId,
      companyBadNews,
      fundamentalsOK,
    );
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

async function handleOpenSettings() {
  try {
    deploymentSettings = await GetDeploymentSettings();
    showSettings = true;
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

async function handleSaveSettings(normal: number, crash: number, extreme: number) {
  try {
    await SaveDeploymentSettings(normal, crash, extreme);
    showSettings = false;
    await loadPlaybook();
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }
}

$effect(() => {
  loadPortfolios();
});
</script>

<div class="p-6">
  <h1 class="font-display text-2xl font-bold text-text-primary">{t("crashPlaybook.title")}</h1>
  <p class="mt-1 text-sm text-text-secondary">{t("crashPlaybook.subtitle")}</p>

  {#if state === "loading"}
    <div class="flex items-center justify-center gap-2 py-16 text-text-secondary" role="status">
      <LoaderCircle size={20} strokeWidth={2} class="animate-spin" />
      <span class="text-sm">{t("crashPlaybook.loading")}</span>
    </div>
  {:else if state === "error"}
    <div class="mt-6 rounded-lg border border-negative bg-negative-bg p-4">
      <p class="text-sm text-negative">{error}</p>
      <div class="mt-3">
        <Button variant="secondary" size="sm" onclick={loadPortfolios}>{t("common.retry")}</Button>
      </div>
    </div>
  {:else if state === "empty"}
    <div class="mt-6 text-center text-sm text-text-secondary">
      {t("crashPlaybook.noPortfolios")}
    </div>
  {:else if state === "ready" && playbook}
    <div class="mt-4">
      <Select value={selectedPortfolioId} onchange={handlePortfolioChange} aria-label="Select portfolio">
        {#each portfolios as p}
          <option value={p.id}>{p.name}</option>
        {/each}
      </Select>
    </div>

    <div class="mt-4">
      <MarketStatusBanner market={playbook.market} />
    </div>

    {#if playbook.market.condition !== "NORMAL"}
      <div class="mt-2 text-xs text-text-secondary">
        {t("crashPlaybook.refreshInterval", { hours: String(Math.floor(playbook.refreshMin / 60)) })}
      </div>
    {/if}

    {#if playbook.stocks.length > 0}
      <div class="mt-4 grid gap-4 md:grid-cols-2">
        {#each playbook.stocks as stock}
          <StockPlaybookCard {stock} onDiagnostic={handleDiagnostic} />
        {/each}
      </div>
    {:else}
      <div class="mt-6 text-center text-sm text-text-secondary">
        {t("crashPlaybook.noHoldings")}
      </div>
    {/if}

    {#if capital}
      <div class="mt-6">
        <CrashCapitalPanel
          {capital}
          plan={deploymentPlan}
          onSave={handleSaveCapital}
          onOpenSettings={handleOpenSettings}
        />
      </div>
    {/if}
  {/if}
</div>

{#if diagnosticTicker && diagnosticResult}
  <FallingKnifeDialog
    ticker={diagnosticTicker}
    diagnostic={diagnosticResult}
    onUpdate={handleDiagnosticUpdate}
    onClose={() => { diagnosticTicker = null; diagnosticResult = null; }}
  />
{/if}

{#if showSettings && deploymentSettings}
  <DeploymentSettings
    settings={deploymentSettings}
    onSave={handleSaveSettings}
    onClose={() => { showSettings = false; }}
  />
{/if}
