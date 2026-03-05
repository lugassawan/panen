<script lang="ts">
import { untrack } from "svelte";
import { CreatePortfolio, UpdatePortfolio } from "../../../wailsjs/go/backend/App";
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";
import type { Mode, PortfolioResponse, RiskProfile } from "../../lib/types";

let {
  brokerageAcctId = "",
  existingPortfolio = null,
  onSaved,
  onCancel,
}: {
  brokerageAcctId?: string;
  existingPortfolio?: PortfolioResponse | null;
  onSaved: () => void;
  onCancel?: () => void;
} = $props();

const isEdit = untrack(() => existingPortfolio != null);
let name = $state(untrack(() => existingPortfolio?.name ?? ""));
let mode = $state<Mode>(untrack(() => existingPortfolio?.mode ?? "VALUE"));
let riskProfile = $state<RiskProfile>(untrack(() => existingPortfolio?.riskProfile ?? "MODERATE"));
let capital = $state(untrack(() => existingPortfolio?.capital ?? 0));
let monthlyAddition = $state(untrack(() => existingPortfolio?.monthlyAddition ?? 0));
let maxStocks = $state(untrack(() => existingPortfolio?.maxStocks ?? 10));
let loading = $state(false);
let error = $state<string | null>(null);

const modeOptions: {
  value: Mode;
  label: string;
  description: string;
  selectedClass: string;
  accentClass: string;
}[] = [
  {
    value: "VALUE",
    label: "Value",
    description: "Focus on undervalued stocks with margin of safety.",
    selectedClass: "border-green-700/50 bg-green-50 dark:bg-green-900/20",
    accentClass: "accent-green-700",
  },
  {
    value: "DIVIDEND",
    label: "Dividend",
    description: "Focus on consistent dividend-paying stocks.",
    selectedClass: "border-gold-500/50 bg-gold-50 dark:bg-gold-500/10",
    accentClass: "accent-gold-500",
  },
];

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

let riskSelectedClass = $derived(
  mode === "VALUE"
    ? "border-green-700/50 bg-green-50 dark:bg-green-900/20"
    : "border-gold-500/50 bg-gold-50 dark:bg-gold-500/10",
);

let riskAccentClass = $derived(mode === "VALUE" ? "accent-green-700" : "accent-gold-500");
let buttonVariant = $derived<"primary" | "gold">(mode === "VALUE" ? "primary" : "gold");

async function submit() {
  error = null;
  if (!name.trim()) {
    error = "Portfolio name is required";
    return;
  }

  loading = true;
  try {
    if (isEdit && existingPortfolio) {
      await UpdatePortfolio(
        existingPortfolio.id,
        name.trim(),
        riskProfile,
        capital,
        monthlyAddition,
        maxStocks,
      );
    } else {
      await CreatePortfolio(
        brokerageAcctId,
        name.trim(),
        mode,
        riskProfile,
        capital,
        monthlyAddition,
        maxStocks,
      );
    }
    onSaved();
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
			class="mb-1 block text-sm text-text-secondary"
		>
			Portfolio Name
		</label>
		<Input
			id="portfolio-name"
			bind:value={name}
			placeholder="e.g. Value Portfolio"
			class="placeholder:text-text-muted"
		/>
	</div>

	<fieldset>
		<legend class="mb-2 text-sm text-text-secondary">Mode</legend>
		<div class="grid grid-cols-2 gap-3">
			{#each modeOptions as option}
				<label
					class="flex cursor-pointer items-start gap-3 rounded border px-3 py-2 transition-colors {mode ===
					option.value
						? option.selectedClass
						: 'border-border-default hover:border-border-strong'} {isEdit ? 'opacity-60 cursor-not-allowed' : ''}"
				>
					<input
						type="radio"
						bind:group={mode}
						value={option.value}
						disabled={isEdit}
						class="mt-0.5 {option.accentClass}"
					/>
					<div>
						<span class="text-sm font-medium">{option.label}</span>
						<p class="text-xs text-text-muted">
							{option.description}
						</p>
					</div>
				</label>
			{/each}
		</div>
		{#if isEdit}
			<p class="mt-1 text-xs text-text-muted">Mode cannot be changed after creation.</p>
		{/if}
	</fieldset>

	<fieldset>
		<legend class="mb-2 text-sm text-text-secondary">Risk Profile</legend>
		<div class="space-y-2">
			{#each riskOptions as option}
				<label
					class="flex cursor-pointer items-start gap-3 rounded border px-3 py-2 transition-colors {riskProfile ===
					option.value
						? riskSelectedClass
						: 'border-border-default hover:border-border-strong'}"
				>
					<input
						type="radio"
						bind:group={riskProfile}
						value={option.value}
						class="mt-0.5 {riskAccentClass}"
					/>
					<div>
						<span class="text-sm font-medium">{option.label}</span>
						<p class="text-xs text-text-muted">
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
				class="mb-1 block text-sm text-text-secondary"
			>
				Capital
			</label>
			<Input
				id="capital"
				type="number"
				bind:value={capital}
				min="0"
			/>
		</div>
		<div>
			<label
				for="monthly-addition"
				class="mb-1 block text-sm text-text-secondary"
			>
				Monthly Addition
			</label>
			<Input
				id="monthly-addition"
				type="number"
				bind:value={monthlyAddition}
				min="0"
			/>
		</div>
		<div>
			<label
				for="max-stocks"
				class="mb-1 block text-sm text-text-secondary"
			>
				Max Stocks
			</label>
			<Input
				id="max-stocks"
				type="number"
				bind:value={maxStocks}
				min="1"
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

	<div class="flex gap-3">
		{#if onCancel}
			<Button variant="secondary" onclick={onCancel}>Cancel</Button>
		{/if}
		<Button variant={buttonVariant} type="submit" loading={loading}>
			{#if loading}
				{isEdit ? "Saving…" : "Creating…"}
			{:else}
				{isEdit ? "Save Changes" : "Create Portfolio"}
			{/if}
		</Button>
	</div>
</form>
