<script lang="ts">
import SkeletonCard from "../SkeletonCard.svelte";

let {
  lines,
  label,
}: {
  lines?: number;
  label?: string;
} = $props();
</script>

<SkeletonCard {lines} {label} />
