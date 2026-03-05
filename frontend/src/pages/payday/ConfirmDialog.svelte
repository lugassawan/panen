<script lang="ts">
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";

let {
  expected,
  portfolioName,
  onConfirm,
  onCancel,
}: {
  expected: number;
  portfolioName: string;
  onConfirm: (amount: number) => void;
  onCancel: () => void;
} = $props();

let initialAmount = expected;
let amount = $state<number>(initialAmount);

function handleConfirm() {
  onConfirm(amount);
}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
  class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
  onclick={(e) => { if (e.target === e.currentTarget) onCancel(); }}
>
  <div class="w-full max-w-sm rounded-lg border border-border-default bg-bg-elevated p-6 shadow-lg">
    <h3 class="text-lg font-semibold text-text-primary font-display">Confirm Payday</h3>
    <p class="mt-2 text-sm text-text-secondary">
      Confirm the payday amount for <span class="font-medium text-text-primary">{portfolioName}</span>.
    </p>

    <div class="mt-4">
      <label for="confirm-amount" class="mb-1.5 block text-sm font-medium text-text-secondary">
        Amount (IDR)
      </label>
      <Input
        id="confirm-amount"
        type="number"
        bind:value={amount}
        min={0}
        aria-label="Payday amount"
        class="font-mono"
      />
    </div>

    <div class="mt-6 flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={onCancel}>Cancel</Button>
      <Button variant="primary" onclick={handleConfirm}>Confirm</Button>
    </div>
  </div>
</div>
