<script lang="ts">
import { ArrowLeft } from "lucide-svelte";
import {
  ClearHoldings,
  DeletePortfolio,
  GetPortfolio,
  ListBrokerageAccounts,
  ListBrokerConfigs,
  ListPortfolios,
  RemoveHolding,
} from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Alert from "../../lib/components/Alert.svelte";
import ConfirmDialog from "../../lib/components/ConfirmDialog.svelte";
import SkeletonTable from "../../lib/components/SkeletonTable.svelte";
import { formatError } from "../../lib/error";
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
let removingHolding = $state<{ id: string; ticker: string } | null>(null);
let removeLoading = $state(false);
let confirmingClearAll = $state(false);
let clearAllLoading = $state(false);
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
    error = formatError(e instanceof Error ? e.message : String(e));
    state = "error";
  }
}

async function viewPortfolio(portfolio: PortfolioResponse) {
  state = "loading";
  try {
    detail = await GetPortfolio(portfolio.id);
    state = "view";
  } catch (e: unknown) {
    error = formatError(e instanceof Error ? e.message : String(e));
    state = "error";
  }
}

async function confirmDelete() {
  if (!deletingPortfolio) return;
  deleteLoading = true;
  deleteError = null;
  try {
    await DeletePortfolio(deletingPortfolio.id);
    toastStore.add(t("common.portfolioDeleted"), "success");
    deletingPortfolio = null;
    await load();
  } catch (e: unknown) {
    deleteError = formatError(e instanceof Error ? e.message : String(e));
  } finally {
    deleteLoading = false;
  }
}

async function confirmRemoveHolding() {
  if (!removingHolding || !detail) return;
  removeLoading = true;
  try {
    await RemoveHolding(detail.portfolio.id, removingHolding.id);
    toastStore.add(t("holding.holdingRemoved", { ticker: removingHolding.ticker }), "success");
    removingHolding = null;
    await viewPortfolio(detail.portfolio);
  } catch (e: unknown) {
    toastStore.add(e instanceof Error ? e.message : String(e), "error");
    removingHolding = null;
  } finally {
    removeLoading = false;
  }
}

async function confirmClearAll() {
  if (!detail) return;
  clearAllLoading = true;
  try {
    await ClearHoldings(detail.portfolio.id);
    toastStore.add(t("holding.holdingsCleared"), "success");
    confirmingClearAll = false;
    await viewPortfolio(detail.portfolio);
  } catch (e: unknown) {
    toastStore.add(e instanceof Error ? e.message : String(e), "error");
    confirmingClearAll = false;
  } finally {
    clearAllLoading = false;
  }
}

load();
</script>

<div class="mx-auto max-w-4xl px-4 py-8">
  {#if state === "loading"}
    <SkeletonTable rows={5} columns={7} label={t("portfolio.loading")} />
  {:else if state === "error"}
    <Alert variant="negative">{error}</Alert>
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
      onRemove={(holdingId, ticker) => {
        removingHolding = { id: holdingId, ticker };
      }}
      onClearAll={() => {
        confirmingClearAll = true;
      }}
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

{#if removingHolding}
  <ConfirmDialog
    title={t("holding.removeHolding")}
    confirmLabel={t("common.remove")}
    confirmVariant="danger"
    loading={removeLoading}
    onConfirm={confirmRemoveHolding}
    onCancel={() => { removingHolding = null; }}
  >
    <p>{t("holding.confirmRemoveMessage", { ticker: removingHolding.ticker })}</p>
    <p class="mt-1">{t("common.cannotUndo")}</p>
  </ConfirmDialog>
{/if}

{#if confirmingClearAll}
  <ConfirmDialog
    title={t("holding.clearAllHoldings")}
    confirmLabel={t("common.delete")}
    confirmVariant="danger"
    loading={clearAllLoading}
    onConfirm={confirmClearAll}
    onCancel={() => { confirmingClearAll = false; }}
  >
    <p>{t("holding.confirmClearMessage")}</p>
    <p class="mt-1">{t("common.cannotUndo")}</p>
  </ConfirmDialog>
{/if}

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
    <p>{t("common.confirmDeleteMessage", { name: deletingPortfolio.name })}</p>
    <p class="mt-1">{t("common.cannotUndo")}</p>
    {#if deleteError}
      <div class="mt-3">
        <Alert variant="negative">{deleteError}</Alert>
      </div>
    {/if}
  </ConfirmDialog>
{/if}
