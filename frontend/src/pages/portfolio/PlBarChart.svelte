<script lang="ts">
import type { HoldingDetailResponse } from "../../lib/types";

interface Props {
  holdings: HoldingDetailResponse[];
}

let { holdings }: Props = $props();
</script>

<div data-testid="pl-bar-chart">
  <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">P/L by Holding</p>
</div>
