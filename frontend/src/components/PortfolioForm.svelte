<script lang="ts">
import { CreatePortfolio } from "../../wailsjs/go/backend/App";
import type { RiskProfile } from "../lib/types";

let { brokerageAcctId, onCreated }: { brokerageAcctId: string; onCreated: () => void } = $props();

let name = $state("");
let riskProfile = $state<RiskProfile>("MODERATE");
let capital = $state(0);
let monthlyAddition = $state(0);
let maxStocks = $state(10);
let loading = $state(false);
let error = $state<string | null>(null);

const riskOptions: {
  value: RiskProfile;
  label: string;
  description: string;
}[] = [
  {
    value: "CONSERVATIVE",
    label: "Conservative",
    description: "Stricter margin of safety. Best for preserving capital.",
  },
  {
    value: "MODERATE",
    label: "Moderate",
    description: "Balanced approach. Best for long-term wealth building.",
  },
  {
    value: "AGGRESSIVE",
    label: "Aggressive",
    description: "Lower margin of safety threshold. Best for growth-focused investors.",
  },
];

async function submit() {
  error = null;
  if (!name.trim()) {
    error = "Portfolio name is required";
    return;
  }

  loading = true;
  try {
    await CreatePortfolio(
      brokerageAcctId,
      name.trim(),
      "VALUE",
      riskProfile,
      capital,
      monthlyAddition,
      maxStocks,
    );
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
			for="portfolio-name"
			class="mb-1 block text-sm text-neutral-400"
		>
			Portfolio Name
		</label>
		<input
			id="portfolio-name"
			bind:value={name}
			placeholder="e.g. Value Portfolio"
			class="w-full rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm placeholder:text-neutral-500 outline-none focus:border-amber-500"
		/>
	</div>

	<fieldset>
		<legend class="mb-2 text-sm text-neutral-400">Risk Profile</legend>
		<div class="space-y-2">
			{#each riskOptions as option}
				<label
					class="flex cursor-pointer items-start gap-3 rounded border px-3 py-2 transition-colors {riskProfile ===
					option.value
						? 'border-amber-500/50 bg-amber-500/5'
						: 'border-neutral-700 hover:border-neutral-600'}"
				>
					<input
						type="radio"
						bind:group={riskProfile}
						value={option.value}
						class="mt-0.5 accent-amber-500"
					/>
					<div>
						<span class="text-sm font-medium">{option.label}</span>
						<p class="text-xs text-neutral-500">
							{option.description}
						</p>
					</div>
				</label>
			{/each}
		</div>
	</fieldset>

	<div class="grid grid-cols-3 gap-4">
		<div>
			<label
				for="capital"
				class="mb-1 block text-sm text-neutral-400"
			>
				Capital
			</label>
			<input
				id="capital"
				type="number"
				bind:value={capital}
				min="0"
				class="w-full rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
			/>
		</div>
		<div>
			<label
				for="monthly-addition"
				class="mb-1 block text-sm text-neutral-400"
			>
				Monthly Addition
			</label>
			<input
				id="monthly-addition"
				type="number"
				bind:value={monthlyAddition}
				min="0"
				class="w-full rounded border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-amber-500"
			/>
		</div>
		<div>
			<label
				for="max-stocks"
				class="mb-1 block text-sm text-neutral-400"
			>
				Max Stocks
			</label>
			<input
				id="max-stocks"
				type="number"
				bind:value={maxStocks}
				min="1"
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
		{loading ? "Creating…" : "Create Portfolio"}
	</button>
</form>
