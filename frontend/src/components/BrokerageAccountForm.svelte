<script lang="ts">
import { CreateBrokerageAccount, UpdateBrokerageAccount } from "../../wailsjs/go/backend/App";
import Button from "../lib/components/Button.svelte";
import type { BrokerageAccountResponse, BrokerConfigResponse } from "../lib/types";

let {
  brokerConfigs = [],
  existingAccount = null,
  onSaved,
  onCancel,
}: {
  brokerConfigs?: BrokerConfigResponse[];
  existingAccount?: BrokerageAccountResponse | null;
  onSaved: (acct: BrokerageAccountResponse) => void;
  onCancel?: () => void;
} = $props();

let isEdit = $derived(existingAccount != null);

let name = $state(existingAccount?.brokerName ?? "");
let brokerCode = $state(existingAccount?.brokerCode ?? "");
let buyFee = $state(existingAccount?.buyFeePct ?? 0.15);
let sellFee = $state(existingAccount?.sellFeePct ?? 0.15);
let sellTax = $state(existingAccount?.sellTaxPct ?? 0.1);
let isManualFee = $state(existingAccount?.isManualFee ?? false);
let loading = $state(false);
let error = $state<string | null>(null);

function onBrokerSelect(code: string) {
  brokerCode = code;
  if (code && code !== "OTHER") {
    const config = brokerConfigs.find((c) => c.code === code);
    if (config) {
      name = config.name;
      if (!isManualFee) {
        buyFee = config.buyFeePct;
        sellFee = config.sellFeePct;
        sellTax = config.sellTaxPct;
      }
    }
  } else if (code === "OTHER") {
    name = "";
    isManualFee = true;
  }
}

function onManualFeeToggle(checked: boolean) {
  isManualFee = checked;
  if (!checked && brokerCode && brokerCode !== "OTHER") {
    const config = brokerConfigs.find((c) => c.code === brokerCode);
    if (config) {
      buyFee = config.buyFeePct;
      sellFee = config.sellFeePct;
      sellTax = config.sellTaxPct;
    }
  }
}

let feesDisabled = $derived(!isManualFee && brokerCode !== "" && brokerCode !== "OTHER");

async function submit() {
  error = null;
  if (!name.trim()) {
    error = "Broker name is required";
    return;
  }

  loading = true;
  try {
    let acct: BrokerageAccountResponse;
    if (isEdit && existingAccount) {
      acct = await UpdateBrokerageAccount(
        existingAccount.id,
        name.trim(),
        brokerCode,
        buyFee,
        sellFee,
        sellTax,
        isManualFee,
      );
    } else {
      acct = await CreateBrokerageAccount(
        name.trim(),
        brokerCode,
        buyFee,
        sellFee,
        sellTax,
        isManualFee,
      );
    }
    onSaved(acct);
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
  } finally {
    loading = false;
  }
}
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    submit();
  }}
  class="space-y-4"
>
  {#if brokerConfigs.length > 0}
    <div>
      <label for="broker-select" class="mb-1 block text-sm text-text-secondary">Broker</label>
      <select
        id="broker-select"
        value={brokerCode}
        onchange={(e) => onBrokerSelect(e.currentTarget.value)}
        class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary outline-none focus:border-green-700 focus-ring"
      >
        <option value="">Select a broker…</option>
        {#each brokerConfigs as config}
          <option value={config.code}>{config.name} ({config.code})</option>
        {/each}
        <option value="OTHER">Other (manual)</option>
      </select>
    </div>
  {/if}

  <div>
    <label for="broker-name" class="mb-1 block text-sm text-text-secondary">Broker Name</label>
    <input
      id="broker-name"
      bind:value={name}
      placeholder="e.g. Ajaib, Stockbit, IPOT"
      disabled={brokerCode !== "" && brokerCode !== "OTHER"}
      class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary placeholder:text-text-muted outline-none focus:border-green-700 focus-ring disabled:opacity-60"
    />
  </div>

  <div class="grid grid-cols-3 gap-4">
    <div>
      <label for="buy-fee" class="mb-1 block text-sm text-text-secondary">Buy Fee %</label>
      <input
        id="buy-fee"
        type="number"
        bind:value={buyFee}
        step="0.01"
        min="0"
        disabled={feesDisabled}
        class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm font-mono text-text-primary outline-none focus:border-green-700 focus-ring disabled:opacity-60"
      />
    </div>
    <div>
      <label for="sell-fee" class="mb-1 block text-sm text-text-secondary">Sell Fee %</label>
      <input
        id="sell-fee"
        type="number"
        bind:value={sellFee}
        step="0.01"
        min="0"
        disabled={feesDisabled}
        class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm font-mono text-text-primary outline-none focus:border-green-700 focus-ring disabled:opacity-60"
      />
    </div>
    <div>
      <label for="sell-tax" class="mb-1 block text-sm text-text-secondary">Sell Tax %</label>
      <input
        id="sell-tax"
        type="number"
        bind:value={sellTax}
        step="0.01"
        min="0"
        disabled={feesDisabled}
        class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm font-mono text-text-primary outline-none focus:border-green-700 focus-ring disabled:opacity-60"
      />
    </div>
  </div>

  {#if brokerCode && brokerCode !== "OTHER"}
    <label class="flex items-center gap-2 text-sm text-text-secondary">
      <input
        type="checkbox"
        checked={isManualFee}
        onchange={(e) => onManualFeeToggle(e.currentTarget.checked)}
        class="rounded border-border-default focus-ring"
      />
      Custom fees
    </label>
  {/if}

  {#if error}
    <div
      class="rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative"
      role="alert"
    >
      {error}
    </div>
  {/if}

  <div class="flex gap-3">
    {#if onCancel}
      <Button variant="secondary" onclick={onCancel}>Cancel</Button>
    {/if}
    <Button type="submit" {loading}>
      {#if loading}
        {isEdit ? "Saving…" : "Creating…"}
      {:else}
        {isEdit ? "Save Changes" : "Create Account"}
      {/if}
    </Button>
  </div>
</form>
