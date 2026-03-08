<script lang="ts">
import type { Snippet } from "svelte";

let {
  open = true,
  title,
  "aria-label": ariaLabel,
  size = "md",
  onClose,
  children,
  footer,
}: {
  open?: boolean;
  title?: string;
  "aria-label"?: string;
  size?: "sm" | "md" | "lg";
  onClose: () => void;
  children: Snippet;
  footer?: Snippet;
} = $props();

const sizeClass: Record<string, string> = {
  sm: "max-w-sm",
  md: "max-w-md",
  lg: "max-w-lg",
};

let dialogEl = $state<HTMLDivElement | null>(null);

$effect(() => {
  if (open) {
    dialogEl?.focus();
  }
});

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    onClose();
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

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="fixed inset-0" role="presentation" onclick={onClose}></div>
    <div
      bind:this={dialogEl}
      class="relative z-10 w-full {sizeClass[size]} rounded-lg border border-border-default bg-bg-elevated p-6 shadow-lg"
      role="dialog"
      aria-modal="true"
      aria-labelledby={title ? "modal-title" : undefined}
      aria-label={title ? undefined : ariaLabel}
      tabindex="-1"
      onkeydown={handleKeydown}
    >
      {#if title}
        <h3 id="modal-title" class="mb-2 font-display text-lg font-semibold text-text-primary">{title}</h3>
      {/if}
      {@render children()}
      {#if footer}
        <div class="mt-6">
          {@render footer()}
        </div>
      {/if}
    </div>
  </div>
{/if}
