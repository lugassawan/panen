<script lang="ts">
import type { Snippet } from "svelte";
import { t } from "../i18n";
import Button from "../lib/components/Button.svelte";

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

let dialogEl = $state<HTMLDivElement | null>(null);

$effect(() => {
  dialogEl?.focus();
});

function trapFocus(e: KeyboardEvent) {
  if (e.key === "Escape") {
    onCancel();
    return;
  }
  if (e.key !== "Tab" || !dialogEl) return;

  const focusable = dialogEl.querySelectorAll<HTMLElement>(
    'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
  );
  if (focusable.length === 0) return;

  const first = focusable[0];
  const last = focusable[focusable.length - 1];

  if (e.shiftKey && document.activeElement === first) {
    e.preventDefault();
    last.focus();
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault();
    first.focus();
  }
}
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
  <div class="fixed inset-0" role="presentation" onclick={onCancel}></div>
  <div
    bind:this={dialogEl}
    class="relative z-10 w-full max-w-md rounded-lg border border-border-default bg-bg-elevated p-6 shadow-lg"
    role="dialog"
    aria-modal="true"
    aria-labelledby="confirm-dialog-title"
    tabindex="-1"
    onkeydown={trapFocus}
  >
    <h3 id="confirm-dialog-title" class="mb-2 text-lg font-semibold text-text-primary">{title}</h3>
    <div class="mb-6 text-sm text-text-secondary">
      {@render children()}
    </div>
    <div class="flex justify-end gap-3">
      <Button variant="secondary" onclick={onCancel} disabled={loading}>{t("common.cancel")}</Button>
      <Button variant={confirmVariant} onclick={onConfirm} {loading}>{confirmLabel}</Button>
    </div>
  </div>
</div>
