<script lang="ts">
import { t } from "../../i18n";
import type { BrokerConfigResponse } from "../types";
import SearchableSelect from "./SearchableSelect.svelte";

let {
  brokerConfigs,
  value = $bindable(""),
  disabled = false,
  id,
  onselect,
}: {
  brokerConfigs: BrokerConfigResponse[];
  value?: string;
  disabled?: boolean;
  id?: string;
  onselect?: (code: string) => void;
} = $props();

function filterFn(config: BrokerConfigResponse, query: string): boolean {
  const q = query.toLowerCase();
  return config.name.toLowerCase().includes(q) || config.code.toLowerCase().includes(q);
}

function displayFn(config: BrokerConfigResponse): string {
  return config.name;
}

function keyFn(config: BrokerConfigResponse): string {
  return config.code;
}

function handleSelect(code: string) {
  onselect?.(code);
}

function selectOther(close: () => void) {
  value = "OTHER";
  close();
  onselect?.("OTHER");
}

let otherDisplay = $derived(value === "OTHER" ? t("brokerage.otherManual") : "");
</script>

<SearchableSelect
  items={brokerConfigs}
  bind:value
  {filterFn}
  {displayFn}
  {keyFn}
  placeholder={t("brokerage.searchBroker")}
  fallbackDisplay={otherDisplay}
  {disabled}
  {id}
  onselect={handleSelect}
>
  {#snippet children({ item })}
    <div>
      <span class="font-medium text-text-primary">{item.name}</span>
      <span class="text-text-muted">({item.code})</span>
    </div>
    <div class="font-mono text-xs text-text-muted">
      {t("brokerage.buy")} {item.buyFeePct}% | {t("brokerage.sell")} {item.sellFeePct}% | {t("brokerage.pph")} {item.sellTaxPct}%
    </div>
  {/snippet}
  {#snippet footer({ close })}
    <button
      type="button"
      onclick={() => selectOther(close)}
      class="w-full cursor-pointer px-3 py-2 text-left text-sm text-text-secondary hover:bg-bg-tertiary"
    >
      {t("brokerage.otherManual")}
    </button>
  {/snippet}
</SearchableSelect>
