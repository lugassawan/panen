<script lang="ts">
import { AvailableActions } from "../../wailsjs/go/backend/App";
import { ACTION_LABELS } from "../lib/action";
import type { ActionType } from "../lib/types";

let {
  portfolioId,
  ticker,
  onselect,
}: {
  portfolioId: string;
  ticker: string;
  onselect: (action: ActionType) => void;
} = $props();

let actions = $state<ActionType[]>([]);
let selected = $state<ActionType | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

async function loadActions() {
  loading = true;
  error = null;
  try {
    const result = await AvailableActions(portfolioId, ticker);
    actions = (result ?? []) as ActionType[];
    if (actions.length > 0) {
      selected = actions[0];
      onselect(actions[0]);
    }
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
  } finally {
    loading = false;
  }
}

function selectAction(action: ActionType) {
  selected = action;
  onselect(action);
}

loadActions();
</script>

<div class="flex flex-wrap gap-2" role="group" aria-label="Action types">
  {#if loading}
    <span class="text-sm text-text-secondary">Loading actions…</span>
  {:else if error}
    <span class="text-sm text-negative" role="alert">{error}</span>
  {:else}
    {#each actions as action}
      <button
        type="button"
        class="rounded-full px-3 py-1 text-sm font-medium transition-fast {selected === action
          ? 'bg-green-700 text-white'
          : 'border border-border-default bg-bg-elevated text-text-secondary hover:bg-bg-tertiary'}"
        onclick={() => selectAction(action)}
        aria-pressed={selected === action}
        data-testid="action-chip"
      >
        {ACTION_LABELS[action]}
      </button>
    {/each}
  {/if}
</div>
