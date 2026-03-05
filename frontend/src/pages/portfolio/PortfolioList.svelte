<script lang="ts">
import { Pencil, Plus, Trash2 } from "lucide-svelte";
import Button from "../../lib/components/Button.svelte";
import { formatRupiah } from "../../lib/format";
import type { PortfolioResponse } from "../../lib/types";

interface Props {
  portfolios: PortfolioResponse[];
  onView: (portfolio: PortfolioResponse) => void;
  onEdit: (portfolio: PortfolioResponse) => void;
  onDelete: (portfolio: PortfolioResponse) => void;
  onCreate: () => void;
}

let { portfolios, onView, onEdit, onDelete, onCreate }: Props = $props();

const MODE_BADGE: Record<string, string> = {
  VALUE: "bg-green-100 text-green-700",
  DIVIDEND: "bg-gold-100 text-gold-700",
};
</script>

<div class="mb-6 flex items-center justify-between">
  <h2 class="text-xl font-semibold text-text-primary">Portfolios</h2>
  {#if portfolios.length < 2}
    <Button onclick={onCreate}>
      <Plus size={16} strokeWidth={2} />
      New Portfolio
    </Button>
  {/if}
</div>

<div class="grid gap-4">
  {#each portfolios as portfolio}
    <div
      class="flex items-center justify-between rounded border border-border-default bg-bg-elevated p-4"
      data-testid="portfolio-card"
    >
      <button
        type="button"
        class="flex-1 text-left"
        onclick={() => onView(portfolio)}
      >
        <div class="flex items-center gap-2">
          <p class="font-medium text-text-primary">{portfolio.name}</p>
          <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {MODE_BADGE[portfolio.mode]}" data-testid="mode-badge">
            {portfolio.mode === "VALUE" ? "Value" : "Dividend"}
          </span>
        </div>
        <div class="mt-1 flex gap-4 text-sm text-text-secondary">
          <span>Risk: {portfolio.riskProfile.charAt(0) + portfolio.riskProfile.slice(1).toLowerCase()}</span>
          <span>Capital: <span class="font-mono">{formatRupiah(portfolio.capital)}</span></span>
        </div>
      </button>
      <div class="flex gap-2">
        <Button variant="ghost" size="sm" onclick={() => onEdit(portfolio)}>
          <Pencil size={14} strokeWidth={2} />
          Edit
        </Button>
        <Button variant="ghost" size="sm" onclick={() => onDelete(portfolio)}>
          <Trash2 size={14} strokeWidth={2} />
          Delete
        </Button>
      </div>
    </div>
  {/each}
</div>
