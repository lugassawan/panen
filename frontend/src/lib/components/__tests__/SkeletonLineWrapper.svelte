<script lang="ts">
import SkeletonLine from "../SkeletonLine.svelte";

let {
  width,
  height,
}: {
  width?: string;
  height?: string;
} = $props();
</script>

<SkeletonLine {width} {height} />
