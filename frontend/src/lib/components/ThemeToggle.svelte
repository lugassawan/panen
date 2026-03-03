<script lang="ts">
import { Monitor, Moon, Sun } from "lucide-svelte";
import { theme } from "../stores/theme.svelte";
</script>

<button
  type="button"
  onclick={() => theme.toggle()}
  class="rounded-md p-2 text-text-secondary hover:bg-bg-tertiary hover:text-text-primary focus-ring transition-fast"
  aria-label="Toggle theme ({theme.preference})"
  title="Theme: {theme.preference}"
>
  {#if theme.preference === "light"}
    <Sun size={20} strokeWidth={1.5} />
  {:else if theme.preference === "dark"}
    <Moon size={20} strokeWidth={1.5} />
  {:else}
    <Monitor size={20} strokeWidth={1.5} />
  {/if}
</button>
