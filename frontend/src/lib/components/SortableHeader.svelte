<script lang="ts">
import { ChevronDown, ChevronUp } from "lucide-svelte";

let {
  label,
  field,
  currentSort,
  ascending,
  onclick,
}: {
  label: string;
  field: string;
  currentSort: string;
  ascending: boolean;
  onclick: (field: string) => void;
} = $props();

let isActive = $derived(currentSort === field);
let ariaSortValue = $derived<"ascending" | "descending" | "none">(
  isActive ? (ascending ? "ascending" : "descending") : "none",
);
</script>

<th aria-sort={ariaSortValue}>
  <button
    type="button"
    class="inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wider text-text-muted hover:text-text-primary transition-fast focus-ring rounded-sm"
    onclick={() => onclick(field)}
  >
    {label}
    {#if isActive}
      {#if ascending}
        <ChevronUp size={14} aria-hidden="true" />
      {:else}
        <ChevronDown size={14} aria-hidden="true" />
      {/if}
    {/if}
  </button>
</th>
