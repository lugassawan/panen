<script lang="ts">
import type { HoldingWeight, Mode, SectorWeight } from "../../lib/types";

interface Props {
  holdingWeights: HoldingWeight[];
  sectorWeights: SectorWeight[];
  portfolioMode: Mode;
}

let { holdingWeights, sectorWeights, portfolioMode }: Props = $props();
</script>

<div data-testid="composition-chart">
  <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Composition</p>
</div>
