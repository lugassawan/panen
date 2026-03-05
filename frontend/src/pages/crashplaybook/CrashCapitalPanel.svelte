<script lang="ts">
import { untrack } from "svelte";
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";
import { formatRupiah } from "../../lib/format";
import type { CrashCapitalResponse, DeploymentPlanResponse } from "../../lib/types";

let {
  capital,
  plan,
  onSave,
  onOpenSettings,
}: {
  capital: CrashCapitalResponse;
  plan: DeploymentPlanResponse | null;
  onSave: (amount: number) => void;
  onOpenSettings: () => void;
} = $props();

let amountStr = $state(untrack(() => (capital.amount > 0 ? String(capital.amount) : "")));

function handleSave() {
  const amount = Number(amountStr);
  if (!Number.isNaN(amount) && amount >= 0) {
    onSave(amount);
  }
}
</script>

<div class="rounded-lg border border-border-default bg-bg-elevated p-4">
  <div class="flex items-center justify-between">
    <h3 class="font-display text-base font-semibold text-text-primary">Crash Capital</h3>
    <button
      class="text-xs text-text-secondary underline transition-fast hover:text-text-primary focus-ring rounded"
      onclick={onOpenSettings}
    >
      Deployment Settings
    </button>
  </div>

  <div class="mt-3 flex items-end gap-3">
    <label class="flex-1">
      <span class="block text-xs font-medium text-text-secondary mb-1">Reserved Amount (Rp)</span>
      <Input
        type="number"
        bind:value={amountStr}
        placeholder="e.g. 10000000"
      />
    </label>
    <Button variant="primary" size="sm" onclick={handleSave}>Save</Button>
  </div>

  {#if plan}
    <div class="mt-4 space-y-2">
      <div class="flex items-center justify-between text-sm">
        <span class="text-text-secondary">Total Reserved</span>
        <span class="font-mono font-medium text-text-primary">{formatRupiah(plan.total)}</span>
      </div>
      <div class="flex items-center justify-between text-sm">
        <span class="text-text-secondary">Deployed</span>
        <span class="font-mono font-medium text-text-primary">{formatRupiah(plan.deployed)}</span>
      </div>
      <div class="flex items-center justify-between text-sm">
        <span class="text-text-secondary">Remaining</span>
        <span class="font-mono font-medium text-profit">{formatRupiah(plan.remaining)}</span>
      </div>

      <hr class="border-border-default" />

      {#each plan.levels as level}
        <div class="flex items-center justify-between text-sm">
          <span class="text-text-secondary">{level.level.replace("_", " ")} ({level.pct}%)</span>
          <span class="font-mono text-text-primary">{formatRupiah(level.amount)}</span>
        </div>
      {/each}
    </div>
  {/if}
</div>
