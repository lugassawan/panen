<script lang="ts">
import type { SectorWeight } from "../../lib/types";

interface Props {
  sectorWeights: SectorWeight[];
}

let { sectorWeights }: Props = $props();
</script>

<div data-testid="sector-warnings"></div>
