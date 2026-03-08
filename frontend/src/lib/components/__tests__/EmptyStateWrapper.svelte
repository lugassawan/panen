<script lang="ts">
import Button from "../Button.svelte";
import EmptyState from "../EmptyState.svelte";

let {
  title = "No items",
  description,
  showAction = false,
  compact = false,
}: {
  title?: string;
  description?: string;
  showAction?: boolean;
  compact?: boolean;
} = $props();
</script>

{#if showAction}
  <EmptyState {title} {description} {compact}>
    {#snippet action()}
      <Button>Add item</Button>
    {/snippet}
  </EmptyState>
{:else}
  <EmptyState {title} {description} {compact} />
{/if}
