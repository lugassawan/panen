<script lang="ts">
import type { Snippet } from "svelte";
import { t } from "../../i18n";
import Button from "./Button.svelte";
import Modal from "./Modal.svelte";

let {
  title,
  confirmLabel = t("common.confirm"),
  confirmVariant = "danger",
  loading = false,
  onConfirm,
  onCancel,
  children,
}: {
  title: string;
  confirmLabel?: string;
  confirmVariant?: "primary" | "danger";
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  children: Snippet;
} = $props();
</script>

<Modal {title} onClose={onCancel} size="md">
  <div class="mb-6 text-sm text-text-secondary">
    {@render children()}
  </div>
  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <Button variant="secondary" onclick={onCancel} disabled={loading}>{t("common.cancel")}</Button>
      <Button variant={confirmVariant} onclick={onConfirm} {loading}>{confirmLabel}</Button>
    </div>
  {/snippet}
</Modal>
