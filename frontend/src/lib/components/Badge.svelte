<script lang="ts">
import type { Snippet } from "svelte";

let {
  variant = "value",
  children,
}: {
  variant?: "value" | "dividend" | "profit" | "loss" | "warning";
  children: Snippet;
} = $props();

const variantClasses: Record<string, string> = {
  value: "bg-green-100 text-green-700",
  dividend: "bg-gold-100 text-gold-700",
  profit: "bg-positive-bg text-positive",
  loss: "bg-negative-bg text-negative",
  warning: "bg-warning-bg text-warning",
};
</script>

<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium {variantClasses[variant]}">
  {@render children()}
</span>
