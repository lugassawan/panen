<script lang="ts">
import type { InvestmentMode } from "../stores/mode.svelte";
import { mode } from "../stores/mode.svelte";

const tabs: { id: InvestmentMode; label: string; emoji: string }[] = [
  { id: "value", label: "Value", emoji: "\u{1F4C8}" },
  { id: "dividend", label: "Dividend", emoji: "\u{1F4B0}" },
];
</script>

<div role="tablist" class="flex gap-1 rounded-lg bg-bg-tertiary p-1">
  {#each tabs as tab}
    <button
      role="tab"
      type="button"
      aria-selected={mode.current === tab.id}
      onclick={() => mode.set(tab.id)}
      class="flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-fast focus-ring
        {mode.current === tab.id
          ? mode.config.tabActiveClass
          : 'text-text-secondary hover:text-text-primary'}"
    >
      <span aria-hidden="true">{tab.emoji}</span> {tab.label}
    </button>
  {/each}
</div>
