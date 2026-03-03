<script lang="ts">
import { CreateBrokerageAccount } from "../../wailsjs/go/backend/App";
import type { BrokerageAccountResponse } from "../lib/types";

let { onCreated }: { onCreated: (acct: BrokerageAccountResponse) => void } = $props();

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
    const acct = await CreateBrokerageAccount(name.trim(), buyFee, sellFee);
    onCreated(acct);
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
			class="mb-1 block text-sm text-text-secondary"
		>
			Broker Name
		</label>
		<input
			id="broker-name"
			bind:value={name}
			placeholder="e.g. Ajaib, Stockbit, IPOT"
			class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary placeholder:text-text-muted outline-none focus:border-green-700 focus-ring"
		/>
	</div>

	<div class="grid grid-cols-2 gap-4">
		<div>
			<label
				for="buy-fee"
				class="mb-1 block text-sm text-text-secondary"
			>
				Buy Fee %
			</label>
			<input
				id="buy-fee"
				type="number"
				bind:value={buyFee}
				step="0.01"
				min="0"
				class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary outline-none focus:border-green-700 focus-ring"
			/>
		</div>
		<div>
			<label
				for="sell-fee"
				class="mb-1 block text-sm text-text-secondary"
			>
				Sell Fee %
			</label>
			<input
				id="sell-fee"
				type="number"
				bind:value={sellFee}
				step="0.01"
				min="0"
				class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary outline-none focus:border-green-700 focus-ring"
			/>
		</div>
	</div>

	{#if error}
		<div
			class="rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative"
			role="alert"
		>
			{error}
		</div>
	{/if}

	<button
		type="submit"
		disabled={loading}
		class="rounded bg-green-700 px-5 py-2 text-sm font-medium text-text-inverse hover:bg-green-800 disabled:opacity-50 focus-ring transition-fast"
	>
		{loading ? "Creating…" : "Create Account"}
	</button>
</form>
