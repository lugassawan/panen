<script lang="ts">
import { AddHolding } from "../../../wailsjs/go/backend/App";

let { portfolioId, onAdded }: { portfolioId: string; onAdded: () => void } = $props();

let ticker = $state("");
let buyPrice = $state(0);
let lots = $state(1);
let date = $state(new Date().toISOString().split("T")[0]);
let loading = $state(false);
let error = $state<string | null>(null);

async function submit() {
  error = null;
  const t = ticker.trim().toUpperCase();
  if (!t) {
    error = "Ticker is required";
    return;
  }
  if (buyPrice <= 0) {
    error = "Buy price must be greater than 0";
    return;
  }

  loading = true;
  try {
    await AddHolding(portfolioId, t, buyPrice, lots, date);
    ticker = "";
    buyPrice = 0;
    lots = 1;
    onAdded();
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
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
	<div>
		<label for="holding-ticker" class="mb-1 block text-sm text-text-secondary">
			Ticker
		</label>
		<input
			id="holding-ticker"
			bind:value={ticker}
			placeholder="e.g. BBCA"
			class="w-28 rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary uppercase placeholder:normal-case placeholder:text-text-muted outline-none focus:border-green-700 focus-ring"
		/>
	</div>

	<div>
		<label
			for="holding-buy-price"
			class="mb-1 block text-sm text-text-secondary"
		>
			Buy Price
		</label>
		<input
			id="holding-buy-price"
			type="number"
			bind:value={buyPrice}
			min="0"
			class="w-32 rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary outline-none focus:border-green-700 focus-ring"
		/>
	</div>

	<div>
		<label for="holding-lots" class="mb-1 block text-sm text-text-secondary">
			Lots
		</label>
		<input
			id="holding-lots"
			type="number"
			bind:value={lots}
			min="1"
			class="w-20 rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary outline-none focus:border-green-700 focus-ring"
		/>
	</div>

	<div>
		<label for="holding-date" class="mb-1 block text-sm text-text-secondary">
			Date
		</label>
		<input
			id="holding-date"
			type="date"
			bind:value={date}
			class="rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary outline-none focus:border-green-700 focus-ring"
		/>
	</div>

	<button
		type="submit"
		disabled={loading}
		class="rounded bg-green-700 px-5 py-2 text-sm font-medium text-text-inverse hover:bg-green-800 disabled:opacity-50 focus-ring transition-fast"
	>
		{loading ? "Adding…" : "Add Holding"}
	</button>

	{#if error}
		<div
			class="w-full rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative"
			role="alert"
		>
			{error}
		</div>
	{/if}
</form>
