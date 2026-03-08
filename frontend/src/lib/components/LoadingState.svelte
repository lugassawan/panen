<script lang="ts">
import { LoaderCircle } from "lucide-svelte";

let {
  message,
  size = "md",
  class: className = "",
}: {
  message?: string;
  size?: "sm" | "md";
  class?: string;
} = $props();
</script>

<div
  class="flex items-center justify-center gap-2 text-text-secondary {className}"
  role="status"
>
  <LoaderCircle
    size={size === "sm" ? 16 : 20}
    strokeWidth={2}
    class="animate-spin"
  />
  {#if message}
    <span class={size === "sm" ? "text-sm" : ""}>{message}</span>
  {/if}
</div>
