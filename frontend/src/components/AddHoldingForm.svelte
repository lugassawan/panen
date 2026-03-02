<script lang="ts">
import { AddHolding } from "../../wailsjs/go/backend/App";

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
		<label for="holding-ticker" class="mb-1 block text-sm text-neutral-400">
			Ticker
		</label>
		<input
			id="holding-ticker"
			bind:value={ticker}
			placeholder="e.g. BBCA"
			class="w-28 rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm uppercase placeholder:normal-case placeholder:text-neutral-500 outline-none focus:border-amber-500"
		/>
	</div>

	<div>
		<label
			for="holding-buy-price"
			class="mb-1 block text-sm text-neutral-400"
		>
			Buy Price
		</label>
		<input
			id="holding-buy-price"
			type="number"
			bind:value={buyPrice}
			min="0"
			class="w-32 rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
		/>
	</div>

	<div>
		<label for="holding-lots" class="mb-1 block text-sm text-neutral-400">
			Lots
		</label>
		<input
			id="holding-lots"
			type="number"
			bind:value={lots}
			min="1"
			class="w-20 rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
		/>
	</div>

	<div>
		<label for="holding-date" class="mb-1 block text-sm text-neutral-400">
			Date
		</label>
		<input
			id="holding-date"
			type="date"
			bind:value={date}
			class="rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
		/>
	</div>

	<button
		type="submit"
		disabled={loading}
		class="rounded bg-amber-600 px-5 py-2 text-sm font-medium hover:bg-amber-500 disabled:opacity-50"
	>
		{loading ? "Adding…" : "Add Holding"}
	</button>

	{#if error}
		<div
			class="w-full rounded border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400"
			role="alert"
		>
			{error}
		</div>
	{/if}
</form>
