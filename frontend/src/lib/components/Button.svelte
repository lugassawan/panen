<script lang="ts">
import type { Snippet } from "svelte";

let {
  variant = "primary",
  size = "md",
  disabled = false,
  type = "button",
  loading = false,
  onclick,
  children,
}: {
  variant?: "primary" | "secondary" | "ghost" | "danger" | "gold";
  size?: "sm" | "md" | "lg";
  disabled?: boolean;
  type?: "button" | "submit" | "reset";
  loading?: boolean;
  onclick?: (e: MouseEvent) => void;
  children: Snippet;
} = $props();

const sizeClasses: Record<string, string> = {
  sm: "px-2.5 py-1 text-xs",
  md: "px-4 py-2 text-sm",
  lg: "px-6 py-3 text-base",
};

const variantClasses: Record<string, string> = {
  primary: "bg-green-700 text-text-inverse hover:bg-green-800",
  secondary: "border border-border-default text-text-primary hover:bg-bg-tertiary",
  ghost: "text-text-secondary hover:bg-bg-tertiary hover:text-text-primary",
  danger: "bg-negative text-text-inverse hover:opacity-90",
  gold: "bg-gold-500 text-text-inverse hover:bg-gold-600",
};
</script>

<button
  {type}
  disabled={disabled || loading}
  {onclick}
  class="inline-flex items-center justify-center gap-2 rounded-md font-medium focus-ring transition-fast disabled:opacity-50 disabled:pointer-events-none {sizeClasses[size]} {variantClasses[variant]}"
>
  {#if loading}
    <svg class="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
      <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25"></circle>
      <path d="M4 12a8 8 0 018-8" stroke="currentColor" stroke-width="4" stroke-linecap="round" class="opacity-75"></path>
    </svg>
  {/if}
  {@render children()}
</button>
