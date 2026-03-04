<script lang="ts">
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

<select
  {id}
  bind:value
  {disabled}
  aria-label={ariaLabel}
  {onchange}
  class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary outline-none focus:border-green-700 disabled:opacity-60 focus-ring {extraClass}"
>
  {@render children()}
</select>
