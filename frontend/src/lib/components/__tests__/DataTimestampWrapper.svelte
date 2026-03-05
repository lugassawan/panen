<script lang="ts">
import DataTimestamp from "../DataTimestamp.svelte";

let {
  date,
  label,
}: {
  date: string;
  label?: string;
} = $props();
</script>

<DataTimestamp {date} {label} />
