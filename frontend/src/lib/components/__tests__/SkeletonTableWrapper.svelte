<script lang="ts">
import SkeletonTable from "../SkeletonTable.svelte";

let {
  rows,
  columns,
  label,
}: {
  rows?: number;
  columns?: number;
  label?: string;
} = $props();
</script>

<SkeletonTable {rows} {columns} {label} />
