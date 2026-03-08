<script lang="ts">
import SkeletonLine from "./SkeletonLine.svelte";

let {
  lines = 3,
  label = "Loading",
}: {
  lines?: number;
  label?: string;
} = $props();
</script>

<div class="rounded-lg border border-border-default bg-bg-elevated p-4" role="status" aria-label={label}>
  <div class="mb-3">
    <SkeletonLine height="1.25rem" width="40%" />
  </div>
  <div class="space-y-2">
    {#each Array(lines) as _, i}
      <SkeletonLine height="1rem" width={i === lines - 1 ? "75%" : "100%"} />
    {/each}
  </div>
</div>
