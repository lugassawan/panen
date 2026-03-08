<script lang="ts">
import SkeletonCard from "../SkeletonCard.svelte";

let {
  lines,
}: {
  lines?: number;
} = $props();
</script>

<SkeletonCard {lines} />
