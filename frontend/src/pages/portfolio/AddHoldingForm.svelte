<script lang="ts">
import { AddHolding } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Alert from "../../lib/components/Alert.svelte";
import { formatError } from "../../lib/error";
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";

let { portfolioId, onAdded }: { portfolioId: string; onAdded: () => void } = $props();

let ticker = $state("");
let buyPrice = $state(0);
let lots = $state(1);
let date = $state(new Date().toISOString().split("T")[0]);
let loading = $state(false);
let error = $state<string | null>(null);

async function submit() {
  error = null;
  const normalizedTicker = ticker.trim().toUpperCase();
  if (!normalizedTicker) {
    error = t("holding.tickerRequired");
    return;
  }
  if (buyPrice <= 0) {
    error = t("holding.buyPriceError");
    return;
  }

  loading = true;
  try {
    await AddHolding(portfolioId, normalizedTicker, buyPrice, lots, date);
    ticker = "";
    buyPrice = 0;
    lots = 1;
    onAdded();
  } catch (e: unknown) {
    error = formatError(e instanceof Error ? e.message : String(e));
  } finally {
    loading = false;
  }
}
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		submit();
	}}
	class="flex flex-wrap items-end gap-3"
>
	<div class="w-28">
		<label for="holding-ticker" class="mb-1 block text-sm text-text-secondary">
			{t("holding.ticker")}
		</label>
		<Input
			id="holding-ticker"
			bind:value={ticker}
			placeholder={t("holding.tickerPlaceholder")}
			class="uppercase placeholder:normal-case placeholder:text-text-muted"
		/>
	</div>

	<div class="w-32">
		<label
			for="holding-buy-price"
			class="mb-1 block text-sm text-text-secondary"
		>
			{t("holding.buyPrice")}
		</label>
		<Input
			id="holding-buy-price"
			type="number"
			bind:value={buyPrice}
			min="0"
		/>
	</div>

	<div class="w-20">
		<label for="holding-lots" class="mb-1 block text-sm text-text-secondary">
			{t("holding.lots")}
		</label>
		<Input
			id="holding-lots"
			type="number"
			bind:value={lots}
			min="1"
		/>
	</div>

	<div>
		<label for="holding-date" class="mb-1 block text-sm text-text-secondary">
			{t("holding.date")}
		</label>
		<Input
			id="holding-date"
			type="date"
			bind:value={date}
		/>
	</div>

	<Button type="submit" disabled={loading}>
		{loading ? t("holding.adding") : t("holding.addHolding")}
	</Button>

	{#if error}
		<Alert variant="negative">{error}</Alert>
	{/if}
</form>
