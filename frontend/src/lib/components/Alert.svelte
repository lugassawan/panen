<script lang="ts">
import { X } from "lucide-svelte";
import type { Snippet } from "svelte";
import { t } from "../../i18n";

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
          class="shrink-0 rounded opacity-70 hover:opacity-100 transition-fast focus-ring"
          aria-label={t("common.dismiss")}
        >
          <X size={16} strokeWidth={2} />
        </button>
      {/if}
    </div>
  </div>
{/if}
