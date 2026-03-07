<script lang="ts">
import { untrack } from "svelte";
import { CreateBrokerageAccount, UpdateBrokerageAccount } from "../../wailsjs/go/backend/App";
import { t } from "../i18n";
import BrokerPicker from "../lib/components/BrokerPicker.svelte";
import Button from "../lib/components/Button.svelte";
import Input from "../lib/components/Input.svelte";
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

let name = $state(untrack(() => existingAccount?.brokerName ?? ""));
let brokerCode = $state(untrack(() => existingAccount?.brokerCode ?? ""));
let buyFee = $state(untrack(() => existingAccount?.buyFeePct ?? 0.15));
let sellFee = $state(untrack(() => existingAccount?.sellFeePct ?? 0.15));
let sellTax = $state(untrack(() => existingAccount?.sellTaxPct ?? 0.1));
let isManualFee = $state(untrack(() => existingAccount?.isManualFee ?? false));
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
    error = t("brokerage.brokerRequired");
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
      <label for="broker-select" class="mb-1 block text-sm text-text-secondary">{t("brokerage.broker")}</label>
      <BrokerPicker
        {brokerConfigs}
        id="broker-select"
        bind:value={brokerCode}
        onselect={onBrokerSelect}
      />
    </div>
  {/if}

  <div>
    <label for="broker-name" class="mb-1 block text-sm text-text-secondary">{t("brokerage.brokerName")}</label>
    <Input
      id="broker-name"
      bind:value={name}
      placeholder={t("brokerage.brokerPlaceholder")}
      disabled={brokerCode !== "" && brokerCode !== "OTHER"}
      class="placeholder:text-text-muted disabled:opacity-60"
    />
  </div>

  <div class="grid grid-cols-3 gap-4">
    <div>
      <label for="buy-fee" class="mb-1 block text-sm text-text-secondary">{t("brokerage.buyFee")}</label>
      <Input
        id="buy-fee"
        type="number"
        bind:value={buyFee}
        step="0.01"
        min="0"
        disabled={feesDisabled}
        class="font-mono disabled:opacity-60"
      />
    </div>
    <div>
      <label for="sell-fee" class="mb-1 block text-sm text-text-secondary">{t("brokerage.sellFee")}</label>
      <Input
        id="sell-fee"
        type="number"
        bind:value={sellFee}
        step="0.01"
        min="0"
        disabled={feesDisabled}
        class="font-mono disabled:opacity-60"
      />
    </div>
    <div>
      <label for="sell-tax" class="mb-1 block text-sm text-text-secondary">{t("brokerage.sellTax")}</label>
      <Input
        id="sell-tax"
        type="number"
        bind:value={sellTax}
        step="0.01"
        min="0"
        disabled={feesDisabled}
        class="font-mono disabled:opacity-60"
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
      {t("brokerage.customFees")}
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
      <Button variant="secondary" onclick={onCancel}>{t("common.cancel")}</Button>
    {/if}
    <Button type="submit" {loading}>
      {#if loading}
        {isEdit ? t("brokerage.saving") : t("brokerage.creating")}
      {:else}
        {isEdit ? t("brokerage.saveChanges") : t("brokerage.createAccount")}
      {/if}
    </Button>
  </div>
</form>
