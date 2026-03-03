<script lang="ts">
import type { Snippet } from "svelte";

let {
  variant = "info",
  dismissible = false,
  children,
}: {
  variant?: "positive" | "warning" | "negative" | "info";
  dismissible?: boolean;
  children: Snippet;
} = $props();

let visible = $state(true);

const variantClasses: Record<string, string> = {
  positive: "bg-positive-bg text-positive border-positive/20",
  warning: "bg-warning-bg text-warning border-warning/20",
  negative: "bg-negative-bg text-negative border-negative/20",
  info: "bg-info-bg text-info border-info/20",
};
</script>

{#if visible}
  <div
    role="alert"
    class="rounded-lg border px-4 py-3 text-sm {variantClasses[variant]}"
  >
    <div class="flex items-start justify-between gap-2">
      <div>{@render children()}</div>
      {#if dismissible}
        <button
          type="button"
          onclick={() => (visible = false)}
          class="shrink-0 opacity-70 hover:opacity-100 transition-fast"
          aria-label="Dismiss"
        >
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      {/if}
    </div>
  </div>
{/if}
