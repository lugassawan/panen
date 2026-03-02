<script lang="ts">
import { CreateBrokerageAccount } from "../../wailsjs/go/backend/App";

let { onCreated }: { onCreated: () => void } = $props();

let name = $state("");
let buyFee = $state(0.15);
let sellFee = $state(0.25);
let loading = $state(false);
let error = $state<string | null>(null);

async function submit() {
  error = null;
  if (!name.trim()) {
    error = "Broker name is required";
    return;
  }

  loading = true;
  try {
    await CreateBrokerageAccount(name.trim(), buyFee, sellFee);
    onCreated();
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
	class="space-y-4"
>
	<div>
		<label
			for="broker-name"
			class="mb-1 block text-sm text-neutral-400"
		>
			Broker Name
		</label>
		<input
			id="broker-name"
			bind:value={name}
			placeholder="e.g. Ajaib, Stockbit, IPOT"
			class="w-full rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm placeholder:text-neutral-500 outline-none focus:border-amber-500"
		/>
	</div>

	<div class="grid grid-cols-2 gap-4">
		<div>
			<label
				for="buy-fee"
				class="mb-1 block text-sm text-neutral-400"
			>
				Buy Fee %
			</label>
			<input
				id="buy-fee"
				type="number"
				bind:value={buyFee}
				step="0.01"
				min="0"
				class="w-full rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
			/>
		</div>
		<div>
			<label
				for="sell-fee"
				class="mb-1 block text-sm text-neutral-400"
			>
				Sell Fee %
			</label>
			<input
				id="sell-fee"
				type="number"
				bind:value={sellFee}
				step="0.01"
				min="0"
				class="w-full rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
			/>
		</div>
	</div>

	{#if error}
		<div
			class="rounded border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400"
			role="alert"
		>
			{error}
		</div>
	{/if}

	<button
		type="submit"
		disabled={loading}
		class="rounded bg-amber-600 px-5 py-2 text-sm font-medium hover:bg-amber-500 disabled:opacity-50"
	>
		{loading ? "Creating…" : "Create Account"}
	</button>
</form>
