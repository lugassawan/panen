<script lang="ts">
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

const riskProfiles: { value: RiskProfile; label: string; thresholds: string }[] = [
  { value: "CONSERVATIVE", label: "Conservative", thresholds: "ROE > 15%, DER < 0.8" },
  { value: "MODERATE", label: "Moderate", thresholds: "ROE > 12%, DER < 1.0" },
  { value: "AGGRESSIVE", label: "Aggressive", thresholds: "ROE > 8%, DER < 1.5" },
];
</script>

<div class="space-y-4 border-b border-border-default px-6 py-4">
  <!-- Row 1: Universe selection -->
  <div class="flex flex-wrap items-end gap-3">
    <div class="w-36">
      <label for="universe-type" class="mb-1 block text-xs font-medium text-text-secondary">Universe</label>
      <Select id="universe-type" bind:value={universeType} onchange={() => { universeName = ""; }}>
        <option value="INDEX">Index</option>
        <option value="SECTOR">Sector</option>
        <option value="CUSTOM">Custom</option>
      </Select>
    </div>

    {#if universeType === "INDEX"}
      <div class="w-44">
        <label for="universe-name" class="mb-1 block text-xs font-medium text-text-secondary">Index</label>
        <Select id="universe-name" bind:value={universeName}>
          <option value="">Select index...</option>
          {#each indices as idx}
            <option value={idx}>{idx}</option>
          {/each}
        </Select>
      </div>
    {:else if universeType === "SECTOR"}
      <div class="w-44">
        <label for="universe-name" class="mb-1 block text-xs font-medium text-text-secondary">Sector</label>
        <Select id="universe-name" bind:value={universeName}>
          <option value="">Select sector...</option>
          {#each sectors as sector}
            <option value={sector}>{sector}</option>
          {/each}
        </Select>
      </div>
    {:else}
      <div class="flex-1 min-w-48">
        <label for="custom-tickers" class="mb-1 block text-xs font-medium text-text-secondary">Tickers (comma-separated)</label>
        <Input id="custom-tickers" bind:value={customTickers} placeholder="BBCA,BMRI,TLKM" />
      </div>
    {/if}

    {#if universeType !== "SECTOR"}
      <div class="w-44">
        <label for="sector-filter" class="mb-1 block text-xs font-medium text-text-secondary">Sector Filter</label>
        <Select id="sector-filter" bind:value={sectorFilter}>
          <option value="">All sectors</option>
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
      <p class="mb-1 text-xs font-medium text-text-secondary">Risk Profile</p>
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
            title={rp.thresholds}
          >
            {rp.label}
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
      {loading ? "Screening..." : "Run Screen"}
    </button>
  </div>
</div>
