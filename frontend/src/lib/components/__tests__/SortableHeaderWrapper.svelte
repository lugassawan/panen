<script lang="ts">
import SortableHeader from "../SortableHeader.svelte";

let {
  label = "Name",
  field = "name",
  currentSort = "",
  ascending = true,
  onclick = () => {},
}: {
  label?: string;
  field?: string;
  currentSort?: string;
  ascending?: boolean;
  onclick?: (field: string) => void;
} = $props();
</script>

<table>
  <thead>
    <tr>
      <SortableHeader {label} {field} {currentSort} {ascending} {onclick} />
    </tr>
  </thead>
</table>
