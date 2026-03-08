<script lang="ts">
import { untrack } from "svelte";
import { t } from "../../i18n";
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";
import Modal from "../../lib/components/Modal.svelte";

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

let amount = $state<number>(untrack(() => expected));

function handleConfirm() {
  onConfirm(amount);
}
</script>

<Modal title={t("payday.confirmTitle")} onClose={onCancel} size="sm">
  <p class="mt-2 text-sm text-text-secondary">
    {t("payday.confirmMessage", { portfolioName })}
  </p>

  <div class="mt-4">
    <label for="confirm-amount" class="mb-1.5 block text-sm font-medium text-text-secondary">
      {t("payday.amountLabel")}
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

  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={onCancel}>{t("common.cancel")}</Button>
      <Button variant="primary" onclick={handleConfirm}>{t("common.confirm")}</Button>
    </div>
  {/snippet}
</Modal>
