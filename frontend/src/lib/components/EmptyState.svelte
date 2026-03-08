<script lang="ts">
import type { Component, Snippet } from "svelte";

let {
  icon,
  title,
  description,
  action,
  compact = false,
}: {
  icon?: Component<{ size: number; class: string }>;
  title: string;
  description?: string;
  action?: Snippet;
  compact?: boolean;
} = $props();
</script>

<div class="flex flex-col items-center justify-center {compact ? 'py-4' : 'py-12'} text-center">
  {#if icon}
    {@const Icon = icon}
    <div class="mb-3">
      <Icon size={48} class="text-text-muted" />
    </div>
  {/if}
  <h3 class="{compact ? 'text-sm' : 'text-lg'} font-medium text-text-primary">{title}</h3>
  {#if description}
    <p class="mt-1 text-sm text-text-secondary max-w-sm">{description}</p>
  {/if}
  {#if action}
    <div class="mt-4">
      {@render action()}
    </div>
  {/if}
</div>
