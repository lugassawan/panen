<script lang="ts">
import ConfirmDialog from "./ConfirmDialog.svelte";

let {
  title,
  body,
  confirmLabel = "Confirm",
  onConfirm,
  onCancel,
}: {
  title: string;
  body: string;
  confirmLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
} = $props();
</script>

<ConfirmDialog {title} {confirmLabel} {onConfirm} {onCancel}>
  <p>{body}</p>
</ConfirmDialog>
