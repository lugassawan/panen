<script lang="ts">
import SkeletonLine from "./SkeletonLine.svelte";

let {
  rows = 5,
  columns = 4,
}: {
  rows?: number;
  columns?: number;
} = $props();

const widthPattern = ["30%", "80%", "70%", "50%"];
</script>

<div role="status" aria-label="Loading">
  <div class="overflow-x-auto rounded border border-border-default">
    <table class="w-full">
      <thead class="border-b border-border-default bg-bg-secondary">
        <tr>
          {#each Array(columns) as _}
            <th class="px-4 py-3">
              <SkeletonLine height="0.75rem" width="60%" />
            </th>
          {/each}
        </tr>
      </thead>
      <tbody class="divide-y divide-border-default">
        {#each Array(rows) as _r, rowIndex}
          <tr>
            {#each Array(columns) as _c, colIndex}
              <td class="px-4 py-3">
                <SkeletonLine height="1rem" width={widthPattern[colIndex % widthPattern.length]} />
              </td>
            {/each}
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
