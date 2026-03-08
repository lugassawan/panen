<script lang="ts">
import { ArrowLeft } from "lucide-svelte";
import {
  DeletePortfolio,
  GetPortfolio,
  ListBrokerageAccounts,
  ListBrokerConfigs,
  ListPortfolios,
} from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import ConfirmDialog from "../../lib/components/ConfirmDialog.svelte";
import SkeletonTable from "../../lib/components/SkeletonTable.svelte";
import { toastStore } from "../../lib/stores/toast.svelte";
import type {
  ActionType,
  BrokerConfigResponse,
  PortfolioDetailResponse,
  PortfolioResponse,
} from "../../lib/types";
import ActionSelector from "./ActionSelector.svelte";
import ChecklistPanel from "./ChecklistPanel.svelte";
import PortfolioDetail from "./PortfolioDetail.svelte";
import PortfolioForm from "./PortfolioForm.svelte";
import PortfolioList from "./PortfolioList.svelte";
import PortfolioOnboarding from "./PortfolioOnboarding.svelte";

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

async function load() {
  state = "loading";
  error = null;

  try {
    const [accounts, configs] = await Promise.all([ListBrokerageAccounts(), ListBrokerConfigs()]);
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

async function confirmDelete() {
  if (!deletingPortfolio) return;
  deleteLoading = true;
  deleteError = null;
  try {
    await DeletePortfolio(deletingPortfolio.id);
    toastStore.add("Portfolio deleted", "success");
    deletingPortfolio = null;
    await load();
  } catch (e: unknown) {
    deleteError = e instanceof Error ? e.message : String(e);
  } finally {
    deleteLoading = false;
  }
}

load();
</script>

<div class="mx-auto max-w-4xl px-4 py-8">
  {#if state === "loading"}
    <SkeletonTable rows={5} columns={7} label={t("portfolio.loading")} />
  {:else if state === "error"}
    <div class="rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
      {error}
    </div>
  {:else if state === "onboarding"}
    <PortfolioOnboarding
      {brokerConfigs}
      {brokerageAcctId}
      showPortfolioForm={onboardingStep === 2}
      onBrokerageCreated={(acct) => {
        brokerageAcctId = acct.id;
        onboardingStep = 2;
      }}
      onPortfolioCreated={() => load()}
    />
  {:else if state === "create-portfolio"}
    <div class="mx-auto max-w-lg">
      <h2 class="mb-6 text-xl font-semibold text-text-primary">{t("portfolio.createPortfolio")}</h2>
      <div class="rounded border border-border-default bg-bg-elevated p-6">
        <PortfolioForm
          brokerageAcctId={brokerageAcctId ?? ""}
          onSaved={() => load()}
        />
      </div>
    </div>
  {:else if state === "list"}
    <PortfolioList
      {portfolios}
      onView={viewPortfolio}
      onEdit={(portfolio) => {
        editingPortfolio = portfolio;
        state = "edit-portfolio";
      }}
      onDelete={(portfolio) => {
        deletingPortfolio = portfolio;
        deleteError = null;
      }}
      onCreate={() => { state = "create-portfolio"; }}
    />
  {:else if state === "edit-portfolio" && editingPortfolio}
    <div class="mx-auto max-w-lg">
      <h3 class="mb-4 text-lg font-semibold text-text-primary">{t("portfolio.editPortfolio")}</h3>
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
    <PortfolioDetail
      {detail}
      onBack={() => load()}
      onChecklist={(ticker) => {
        checklistTicker = ticker;
        checklistAction = null;
        state = "checklist";
      }}
      onHoldingAdded={() => viewPortfolio(detail!.portfolio)}
    />
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
        aria-label={t("portfolio.backToPortfolio")}
      >
        <ArrowLeft size={20} strokeWidth={2} />
      </button>
      <h2 class="text-xl font-semibold text-text-primary">
        {t("portfolio.checklist", { ticker: checklistTicker })}
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
    title={t("portfolio.deletePortfolio")}
    confirmLabel={t("common.delete")}
    confirmVariant="danger"
    loading={deleteLoading}
    onConfirm={confirmDelete}
    onCancel={() => {
      deletingPortfolio = null;
      deleteError = null;
    }}
  >
    <p>Are you sure you want to delete <strong>{deletingPortfolio.name}</strong>?</p>
    <p class="mt-1">{t("common.cannotUndo")}</p>
    {#if deleteError}
      <div class="mt-3 rounded border border-negative/20 bg-negative-bg px-3 py-2 text-sm text-negative" role="alert">
        {deleteError}
      </div>
    {/if}
  </ConfirmDialog>
{/if}
