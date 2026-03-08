<script lang="ts">
import SkeletonTable from "../SkeletonTable.svelte";

let {
  rows,
  columns,
}: {
  rows?: number;
  columns?: number;
} = $props();
</script>

<SkeletonTable {rows} {columns} />
