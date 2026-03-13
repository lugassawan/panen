<script lang="ts">
import { Pencil, Plus, Trash2 } from "lucide-svelte";
import { t } from "../../i18n";
import Button from "../../lib/components/Button.svelte";
import { formatRupiah } from "../../lib/format";
import { MODE_BADGE } from "../../lib/mode-styles";
import type { PortfolioResponse } from "../../lib/types";

interface Props {
  portfolios: PortfolioResponse[];
  brokerNameMap?: Record<string, string>;
  onView: (portfolio: PortfolioResponse) => void;
  onEdit: (portfolio: PortfolioResponse) => void;
  onDelete: (portfolio: PortfolioResponse) => void;
  onCreate: () => void;
}

let { portfolios, brokerNameMap = {}, onView, onEdit, onDelete, onCreate }: Props = $props();
</script>

<div class="mb-6 flex items-center justify-between">
  <h1 class="text-2xl font-display font-bold text-text-primary">{t("portfolio.title")}</h1>
  {#if portfolios.length < 2}
    <Button onclick={onCreate}>
      <Plus size={16} strokeWidth={2} />
      {t("portfolio.newPortfolio")}
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
        class="flex-1 rounded text-left focus-ring"
        onclick={() => onView(portfolio)}
      >
        <div class="flex items-center gap-2">
          <p class="font-medium text-text-primary">{portfolio.name}</p>
          <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {MODE_BADGE[portfolio.mode]}" data-testid="mode-badge">
            {portfolio.mode === "VALUE" ? t("mode.value") : t("mode.dividend")}
          </span>
        </div>
        <div class="mt-1 flex gap-4 text-sm text-text-secondary">
          <span>{t("portfolio.risk")} {portfolio.riskProfile.charAt(0) + portfolio.riskProfile.slice(1).toLowerCase()}</span>
          <span>{t("portfolio.capital")} <span class="font-mono">{formatRupiah(portfolio.capital)}</span></span>
          {#if brokerNameMap[portfolio.brokerageAcctId]}
            <span>{brokerNameMap[portfolio.brokerageAcctId]}</span>
          {/if}
        </div>
      </button>
      <div class="flex gap-2">
        <Button variant="ghost" size="sm" onclick={() => onEdit(portfolio)}>
          <Pencil size={14} strokeWidth={2} />
          {t("common.edit")}
        </Button>
        <Button variant="ghost" size="sm" onclick={() => onDelete(portfolio)}>
          <Trash2 size={14} strokeWidth={2} />
          {t("common.delete")}
        </Button>
      </div>
    </div>
  {/each}
</div>
