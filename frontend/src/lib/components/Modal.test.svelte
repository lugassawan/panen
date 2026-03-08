<script lang="ts">
import type { Snippet } from "svelte";
import Modal from "./Modal.svelte";

let {
  open = true,
  title,
  "aria-label": ariaLabel,
  size,
  onClose,
  hasFooter = false,
}: {
  open?: boolean;
  title?: string;
  "aria-label"?: string;
  size?: "sm" | "md" | "lg";
  onClose: () => void;
  hasFooter?: boolean;
} = $props();
</script>

<Modal {open} {title} aria-label={ariaLabel} {size} {onClose}>
  <p>Dialog body content</p>
  <button type="button">First</button>
  <button type="button">Last</button>
  {#snippet footer()}
    {#if hasFooter}
      <button type="button">Footer action</button>
    {/if}
  {/snippet}
</Modal>
