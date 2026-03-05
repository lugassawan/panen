<script lang="ts">
import { X } from "lucide-svelte";
import type { ToastVariant } from "../stores/toast.svelte";

let {
  message,
  variant,
  onDismiss,
}: {
  message: string;
  variant: ToastVariant;
  onDismiss: () => void;
} = $props();

const variantClasses: Record<ToastVariant, string> = {
  success: "bg-positive-bg text-positive border-positive/20",
  error: "bg-negative-bg text-negative border-negative/20",
  warning: "bg-warning-bg text-warning border-warning/20",
  info: "bg-info-bg text-info border-info/20",
};
</script>

<div
  role="status"
  aria-live={variant === "error" ? "assertive" : "polite"}
  class="flex items-center justify-between gap-3 rounded-lg border px-4 py-3 text-sm shadow-md transition-fast {variantClasses[variant]}"
>
  <span>{message}</span>
  <button
    type="button"
    onclick={onDismiss}
    class="shrink-0 opacity-70 hover:opacity-100 transition-fast focus-ring rounded-sm"
    aria-label="Dismiss notification"
  >
    <X size={14} strokeWidth={2} aria-hidden="true" />
  </button>
</div>
