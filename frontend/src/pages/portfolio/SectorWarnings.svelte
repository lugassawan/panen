<script lang="ts">
import { AlertTriangle } from "lucide-svelte";
import Alert from "../../lib/components/Alert.svelte";
import { formatPercent } from "../../lib/format";
import type { SectorWeight } from "../../lib/types";

const THRESHOLD = 30;

interface Props {
  sectorWeights: SectorWeight[];
}

let { sectorWeights }: Props = $props();

let concentrated = $derived(sectorWeights.filter((s) => s.pct > THRESHOLD));
</script>

<div data-testid="sector-warnings">
  {#if concentrated.length > 0}
    <div class="mb-6 space-y-2">
      {#each concentrated as sector}
        <Alert variant="warning">
          <div class="flex items-center gap-2">
            <AlertTriangle size={16} strokeWidth={2} aria-hidden="true" />
            <span>High concentration: <strong>{sector.sector}</strong> represents {formatPercent(sector.pct)} of your portfolio (threshold: {THRESHOLD}%)</span>
          </div>
        </Alert>
      {/each}
    </div>
  {/if}
</div>
