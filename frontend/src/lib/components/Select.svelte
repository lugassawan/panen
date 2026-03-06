<script lang="ts">
import { ChevronDown } from "lucide-svelte";
import type { Snippet } from "svelte";

let {
  id,
  value = $bindable<string | number>(),
  disabled = false,
  "aria-label": ariaLabel,
  onchange,
  class: extraClass = "",
  children,
}: {
  id?: string;
  value?: string | number;
  disabled?: boolean;
  "aria-label"?: string;
  onchange?: (e: Event & { currentTarget: HTMLSelectElement }) => void;
  class?: string;
  children: Snippet;
} = $props();
</script>

<div class="relative">
  <select
    {id}
    bind:value
    {disabled}
    aria-label={ariaLabel}
    {onchange}
    class="w-full appearance-none rounded border border-border-default bg-bg-elevated py-2 pr-8 pl-3 text-sm text-text-primary outline-none transition-colors transition-fast hover:border-border-strong focus-border disabled:opacity-60 focus-ring {extraClass}"
  >
    {@render children()}
  </select>
  <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2 text-text-tertiary">
    <ChevronDown size={16} strokeWidth={2} />
  </div>
</div>
