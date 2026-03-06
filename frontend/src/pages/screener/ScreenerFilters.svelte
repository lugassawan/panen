<script lang="ts">
import { t } from "../../i18n";
import Input from "../../lib/components/Input.svelte";
import Select from "../../lib/components/Select.svelte";
import type { RiskProfile } from "../../lib/types";

let {
  universeType = $bindable<string>("INDEX"),
  universeName = $bindable<string>(""),
  riskProfile = $bindable<RiskProfile>("MODERATE"),
  sectorFilter = $bindable<string>(""),
  customTickers = $bindable<string>(""),
  indices,
  sectors,
  loading = false,
  onrun,
}: {
  universeType: string;
  universeName: string;
  riskProfile: RiskProfile;
  sectorFilter: string;
  customTickers: string;
  indices: string[];
  sectors: string[];
  loading?: boolean;
  onrun: () => void;
} = $props();

const riskProfiles: { value: RiskProfile; labelKey: string; thresholdsKey: string }[] = [
  {
    value: "CONSERVATIVE",
    labelKey: "screener.conservative",
    thresholdsKey: "screener.conservativeThresholds",
  },
  {
    value: "MODERATE",
    labelKey: "screener.moderate",
    thresholdsKey: "screener.moderateThresholds",
  },
  {
    value: "AGGRESSIVE",
    labelKey: "screener.aggressive",
    thresholdsKey: "screener.aggressiveThresholds",
  },
];
</script>

<div class="space-y-4 border-b border-border-default px-6 py-4">
  <!-- Row 1: Universe selection -->
  <div class="flex flex-wrap items-end gap-3">
    <div class="w-36">
      <label for="universe-type" class="mb-1 block text-xs font-medium text-text-secondary">{t("screener.universe")}</label>
      <Select id="universe-type" bind:value={universeType} onchange={() => { universeName = ""; }}>
        <option value="INDEX">{t("screener.index")}</option>
        <option value="SECTOR">{t("screener.sector")}</option>
        <option value="CUSTOM">{t("screener.custom")}</option>
      </Select>
    </div>

    {#if universeType === "INDEX"}
      <div class="w-44">
        <label for="universe-name" class="mb-1 block text-xs font-medium text-text-secondary">{t("screener.index")}</label>
        <Select id="universe-name" bind:value={universeName}>
          <option value="">{t("screener.selectIndex")}</option>
          {#each indices as idx}
            <option value={idx}>{idx}</option>
          {/each}
        </Select>
      </div>
    {:else if universeType === "SECTOR"}
      <div class="w-44">
        <label for="universe-name" class="mb-1 block text-xs font-medium text-text-secondary">{t("screener.sector")}</label>
        <Select id="universe-name" bind:value={universeName}>
          <option value="">{t("screener.selectSector")}</option>
          {#each sectors as sector}
            <option value={sector}>{sector}</option>
          {/each}
        </Select>
      </div>
    {:else}
      <div class="flex-1 min-w-48">
        <label for="custom-tickers" class="mb-1 block text-xs font-medium text-text-secondary">{t("screener.tickersLabel")}</label>
        <Input id="custom-tickers" bind:value={customTickers} placeholder={t("screener.tickersPlaceholder")} />
      </div>
    {/if}

    {#if universeType !== "SECTOR"}
      <div class="w-44">
        <label for="sector-filter" class="mb-1 block text-xs font-medium text-text-secondary">{t("screener.sectorFilter")}</label>
        <Select id="sector-filter" bind:value={sectorFilter}>
          <option value="">{t("screener.allSectors")}</option>
          {#each sectors as sector}
            <option value={sector}>{sector}</option>
          {/each}
        </Select>
      </div>
    {/if}
  </div>

  <!-- Row 2: Risk profile + Run button -->
  <div class="flex items-end gap-4">
    <div>
      <p class="mb-1 text-xs font-medium text-text-secondary">{t("screener.riskProfile")}</p>
      <div class="flex gap-1" role="radiogroup" aria-label="Risk profile">
        {#each riskProfiles as rp}
          <button
            type="button"
            role="radio"
            aria-checked={riskProfile === rp.value}
            class="rounded border px-3 py-1.5 text-xs font-medium transition-fast focus-ring {riskProfile === rp.value
              ? 'border-green-700 bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
              : 'border-border-default bg-bg-elevated text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
            onclick={() => { riskProfile = rp.value; }}
            title={t(rp.thresholdsKey)}
          >
            {t(rp.labelKey)}
          </button>
        {/each}
      </div>
    </div>

    <button
      type="button"
      class="rounded bg-green-700 px-5 py-1.5 text-sm font-medium text-white transition-fast focus-ring hover:bg-green-800 disabled:opacity-50"
      disabled={loading || (universeType !== "CUSTOM" && !universeName)}
      onclick={onrun}
    >
      {loading ? t("screener.screening") : t("screener.runScreen")}
    </button>
  </div>
</div>
