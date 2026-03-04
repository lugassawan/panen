<script lang="ts">
import { LoaderCircle, Pencil, Plus, Trash2 } from "lucide-svelte";
import {
  DeleteBrokerageAccount,
  ListBrokerageAccounts,
  ListBrokerConfigs,
} from "../../../wailsjs/go/backend/App";
import BrokerageAccountForm from "../../components/BrokerageAccountForm.svelte";
import ConfirmDialog from "../../components/ConfirmDialog.svelte";
import Button from "../../lib/components/Button.svelte";
import { formatPercent } from "../../lib/format";
import type { BrokerageAccountResponse, BrokerConfigResponse } from "../../lib/types";

type PageState = "loading" | "list" | "create" | "edit" | "error";

let state = $state<PageState>("loading");
let error = $state<string | null>(null);
let accounts = $state<BrokerageAccountResponse[]>([]);
let brokerConfigs = $state<BrokerConfigResponse[]>([]);
let editingAccount = $state<BrokerageAccountResponse | null>(null);
let deletingAccount = $state<BrokerageAccountResponse | null>(null);
let deleteLoading = $state(false);
let deleteError = $state<string | null>(null);

async function load() {
  state = "loading";
  error = null;

  try {
    const [accts, configs] = await Promise.all([ListBrokerageAccounts(), ListBrokerConfigs()]);
    accounts = accts ?? [];
    brokerConfigs = configs ?? [];
    state = "list";
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

function startCreate() {
  state = "create";
}

function startEdit(acct: BrokerageAccountResponse) {
  editingAccount = acct;
  state = "edit";
}

function onSaved() {
  editingAccount = null;
  load();
}

function cancelForm() {
  editingAccount = null;
  state = "list";
}

function startDelete(acct: BrokerageAccountResponse) {
  deletingAccount = acct;
  deleteError = null;
}

async function confirmDelete() {
  if (!deletingAccount) return;
  deleteLoading = true;
  deleteError = null;
  try {
    await DeleteBrokerageAccount(deletingAccount.id);
    deletingAccount = null;
    load();
  } catch (e: unknown) {
    deleteError = e instanceof Error ? e.message : String(e);
  } finally {
    deleteLoading = false;
  }
}

function cancelDelete() {
  deletingAccount = null;
  deleteError = null;
}

load();
</script>

<div class="mx-auto max-w-4xl px-4 py-8">
  <div class="mb-6 flex items-center justify-between">
    <h2 class="text-xl font-semibold text-text-primary">Brokerage Accounts</h2>
    {#if state === "list" && accounts.length > 0}
      <Button onclick={startCreate}>
        <Plus size={16} strokeWidth={2} />
        Add Account
      </Button>
    {/if}
  </div>

  {#if state === "loading"}
    <div class="flex items-center justify-center gap-2 py-12 text-text-secondary" role="status">
      <LoaderCircle size={20} strokeWidth={2} class="animate-spin" />
      <span>Loading accounts…</span>
    </div>
  {:else if state === "error"}
    <div class="rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
      {error}
    </div>
  {:else if state === "create"}
    <div class="mx-auto max-w-lg">
      <h3 class="mb-4 text-lg font-semibold text-text-primary">New Brokerage Account</h3>
      <div class="rounded border border-border-default bg-bg-elevated p-6">
        <BrokerageAccountForm {brokerConfigs} onSaved={onSaved} onCancel={cancelForm} />
      </div>
    </div>
  {:else if state === "edit" && editingAccount}
    <div class="mx-auto max-w-lg">
      <h3 class="mb-4 text-lg font-semibold text-text-primary">Edit Brokerage Account</h3>
      <div class="rounded border border-border-default bg-bg-elevated p-6">
        <BrokerageAccountForm
          {brokerConfigs}
          existingAccount={editingAccount}
          onSaved={onSaved}
          onCancel={cancelForm}
        />
      </div>
    </div>
  {:else if state === "list"}
    {#if accounts.length === 0}
      <div class="rounded border border-border-default bg-bg-elevated px-6 py-12 text-center">
        <p class="mb-2 text-text-primary font-medium">No brokerage accounts yet</p>
        <p class="mb-6 text-sm text-text-secondary">
          Add your first brokerage account to start tracking fees and managing portfolios.
        </p>
        <Button onclick={startCreate}>
          <Plus size={16} strokeWidth={2} />
          Add Account
        </Button>
      </div>
    {:else}
      <div class="grid gap-4">
        {#each accounts as acct}
          <div
            class="flex items-center justify-between rounded border border-border-default bg-bg-elevated p-4"
            data-testid="brokerage-card"
          >
            <div>
              <p class="font-medium text-text-primary">{acct.brokerName}</p>
              {#if acct.brokerCode}
                <p class="text-xs text-text-muted">{acct.brokerCode}</p>
              {/if}
              <div class="mt-1 flex gap-4 text-sm text-text-secondary">
                <span>Buy: <span class="font-mono">{formatPercent(acct.buyFeePct)}</span></span>
                <span>Sell: <span class="font-mono">{formatPercent(acct.sellFeePct)}</span></span>
                <span>PPh: <span class="font-mono">{formatPercent(acct.sellTaxPct)}</span></span>
              </div>
            </div>
            <div class="flex gap-2">
              <Button variant="ghost" size="sm" onclick={() => startEdit(acct)}>
                <Pencil size={14} strokeWidth={2} />
                Edit
              </Button>
              <Button variant="ghost" size="sm" onclick={() => startDelete(acct)}>
                <Trash2 size={14} strokeWidth={2} />
                Delete
              </Button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>

{#if deletingAccount}
  <ConfirmDialog
    title="Delete Brokerage Account"
    confirmLabel="Delete"
    confirmVariant="danger"
    loading={deleteLoading}
    onConfirm={confirmDelete}
    onCancel={cancelDelete}
  >
    <p>Are you sure you want to delete <strong>{deletingAccount.brokerName}</strong>?</p>
    <p class="mt-1">This action cannot be undone.</p>
    {#if deleteError}
      <div class="mt-3 rounded border border-negative/20 bg-negative-bg px-3 py-2 text-sm text-negative" role="alert">
        {deleteError}
      </div>
    {/if}
  </ConfirmDialog>
{/if}
