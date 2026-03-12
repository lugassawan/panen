<script lang="ts">
import { SellHolding } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Alert from "../../lib/components/Alert.svelte";
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";
import Modal from "../../lib/components/Modal.svelte";
import { formatError } from "../../lib/error";
import { formatRupiah } from "../../lib/format";

interface Props {
  portfolioId: string;
  holdingId: string;
  ticker: string;
  maxLots: number;
  avgBuyPrice: number;
  onSold: () => void;
  onClose: () => void;
}

let { portfolioId, holdingId, ticker, maxLots, avgBuyPrice, onSold, onClose }: Props = $props();

let sellPrice = $state(0);
let lots = $state(1);
let date = $state(new Date().toISOString().split("T")[0]);
let loading = $state(false);
let error = $state<string | null>(null);

async function submit() {
  error = null;
  if (sellPrice <= 0) {
    error = t("holding.sellPriceError");
    return;
  }
  if (lots < 1 || lots > maxLots) {
    error = t("holding.lotsError", { max: String(maxLots) });
    return;
  }

  loading = true;
  try {
    const tx = await SellHolding(portfolioId, holdingId, sellPrice, lots, date);
    onSold();
    if (tx) {
      // Toast is handled by the parent after onSold refreshes.
    }
  } catch (e: unknown) {
    error = formatError(e instanceof Error ? e.message : String(e));
  } finally {
    loading = false;
  }
}
</script>

<Modal title="{t('holding.sellHolding')}: {ticker}" onClose={onClose}>
  <form
    onsubmit={(e) => {
      e.preventDefault();
      submit();
    }}
    class="space-y-4"
  >
    <div class="rounded border border-border-default bg-bg-secondary p-3">
      <div class="flex justify-between text-sm">
        <span class="text-text-secondary">{t("holding.avgBuyPrice")}</span>
        <span class="font-mono">{formatRupiah(avgBuyPrice)}</span>
      </div>
      <div class="mt-1 flex justify-between text-sm">
        <span class="text-text-secondary">{t("holding.lots")}</span>
        <span>{maxLots}</span>
      </div>
    </div>

    <div>
      <label for="sell-price" class="mb-1 block text-sm text-text-secondary">
        {t("holding.sellPrice")}
      </label>
      <Input id="sell-price" type="number" bind:value={sellPrice} min="0" />
    </div>

    <div>
      <label for="sell-lots" class="mb-1 block text-sm text-text-secondary">
        {t("holding.lots")}
      </label>
      <Input id="sell-lots" type="number" bind:value={lots} min="1" />
    </div>

    <div>
      <label for="sell-date" class="mb-1 block text-sm text-text-secondary">
        {t("holding.date")}
      </label>
      <Input id="sell-date" type="date" bind:value={date} />
    </div>

    {#if error}
      <Alert variant="negative">{error}</Alert>
    {/if}

    <div class="flex justify-end gap-2">
      <Button variant="ghost" type="button" onclick={onClose}>
        {t("common.cancel")}
      </Button>
      <Button type="submit" disabled={loading}>
        {loading ? t("holding.selling") : t("holding.sell")}
      </Button>
    </div>
  </form>
</Modal>
