<script lang="ts">
import { CreateWatchlist } from "../../../wailsjs/go/backend/App";
import Input from "../../lib/components/Input.svelte";

let {
  onCreated,
  onCancel,
}: {
  onCreated: () => void;
  onCancel: () => void;
} = $props();

let name = $state("");
let loading = $state(false);
let error = $state<string | null>(null);

async function submit(e: Event) {
  e.preventDefault();
  const trimmed = name.trim();
  if (!trimmed) return;
  loading = true;
  error = null;
  try {
    await CreateWatchlist(trimmed);
    name = "";
    onCreated();
  } catch (err: unknown) {
    error = err instanceof Error ? err.message : String(err);
  } finally {
    loading = false;
  }
}
</script>

<form
  onsubmit={submit}
  class="mb-2 rounded border border-border-default bg-bg-elevated px-2 py-2"
>
  <Input
    bind:value={name}
    placeholder="Watchlist name"
    aria-label="New watchlist name"
    class="mb-1.5 bg-bg-primary px-2 py-1 text-xs placeholder:text-text-muted"
    disabled={loading}
  />
  {#if error}
    <p class="mb-1.5 text-xs text-negative">{error}</p>
  {/if}
  <div class="flex gap-1.5">
    <button
      type="submit"
      disabled={loading || !name.trim()}
      class="flex-1 rounded bg-green-700 px-2 py-1 text-xs font-medium text-text-inverse transition-fast focus-ring hover:bg-green-800 disabled:pointer-events-none disabled:opacity-50"
    >
      {loading ? "Adding…" : "Add"}
    </button>
    <button
      type="button"
      class="rounded px-2 py-1 text-xs text-text-secondary transition-fast focus-ring hover:bg-bg-tertiary"
      onclick={onCancel}
      disabled={loading}
    >
      Cancel
    </button>
  </div>
</form>
