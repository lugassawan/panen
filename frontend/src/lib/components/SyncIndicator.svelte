<script lang="ts">
import { AlertTriangle, Check, LoaderCircle } from "lucide-svelte";
import { TriggerRefresh } from "../../../wailsjs/go/backend/App";
import { sync } from "../stores/sync.svelte";

function formatRelativeTime(isoString: string): string {
  if (!isoString) return "Not synced yet";
  const date = new Date(isoString);
  const now = Date.now();
  const diffMs = now - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "just now";
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHrs = Math.floor(diffMin / 60);
  if (diffHrs < 24) return `${diffHrs}h ago`;
  const diffDays = Math.floor(diffHrs / 24);
  return `${diffDays}d ago`;
}

let retrying = $state(false);

async function retry() {
  retrying = true;
  try {
    await TriggerRefresh();
  } finally {
    retrying = false;
  }
}
</script>

<div class="px-4 py-2">
  {#if sync.isSyncing}
    <div class="flex items-center gap-2">
      <LoaderCircle size={14} class="animate-spin text-green-700 shrink-0" />
      <span class="text-xs text-text-secondary truncate">
        Syncing {sync.currentTicker ?? "..."}
        {#if sync.progress}
          <span class="font-mono">({sync.progress.index + 1}/{sync.progress.total})</span>
        {/if}
      </span>
    </div>
    {#if sync.progress}
      <div class="mt-1 h-0.5 rounded-full bg-bg-tertiary overflow-hidden">
        <div
          class="h-full rounded-full bg-green-700 transition-all duration-300"
          style="width: {sync.progressPercent}%"
        ></div>
      </div>
    {/if}
  {:else if sync.hasError}
    <div class="flex items-center gap-2">
      <AlertTriangle size={14} class="text-negative shrink-0" />
      <span class="text-xs text-negative truncate">{sync.errorMessage ?? "Sync failed"}</span>
    </div>
    <button
      onclick={retry}
      disabled={retrying}
      class="mt-1 text-xs text-green-700 hover:text-green-800 transition-fast focus-ring rounded"
    >
      {retrying ? "Retrying..." : "Retry"}
    </button>
  {:else}
    <div class="flex items-center gap-2">
      <Check size={14} class="text-green-700 shrink-0" />
      <span class="text-xs text-text-muted font-mono">
        {formatRelativeTime(sync.lastRefresh)}
      </span>
    </div>
  {/if}
</div>
