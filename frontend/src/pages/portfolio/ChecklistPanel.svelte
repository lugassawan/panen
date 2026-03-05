<script lang="ts">
import { Check, LoaderCircle, X } from "lucide-svelte";
import {
  EvaluateChecklist,
  ResetChecklist,
  ToggleManualCheck,
} from "../../../wailsjs/go/backend/App";
import Button from "../../lib/components/Button.svelte";
import type { ActionType, ChecklistEvaluationResponse, CheckResultResponse } from "../../lib/types";
import SuggestionCard from "./SuggestionCard.svelte";

let {
  portfolioId,
  ticker,
  action,
}: {
  portfolioId: string;
  ticker: string;
  action: ActionType;
} = $props();

let evaluation = $state<ChecklistEvaluationResponse | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);
let toggling = $state(false);

async function loadChecklist() {
  loading = true;
  error = null;
  try {
    evaluation = await EvaluateChecklist(portfolioId, ticker, action);
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
  } finally {
    loading = false;
  }
}

async function handleToggle(check: CheckResultResponse, completed: boolean) {
  toggling = true;
  try {
    await ToggleManualCheck(portfolioId, ticker, action, check.key, completed);
    await loadChecklist();
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
  } finally {
    toggling = false;
  }
}

async function handleReset() {
  try {
    await ResetChecklist(portfolioId, ticker, action);
    await loadChecklist();
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
  }
}

$effect(() => {
  void action;
  loadChecklist();
});
</script>

<div class="space-y-4">
  {#if loading}
    <div
      class="flex items-center justify-center gap-2 py-8 text-text-secondary"
      role="status"
    >
      <LoaderCircle size={20} strokeWidth={2} class="animate-spin" />
      <span>Evaluating checklist…</span>
    </div>
  {:else if error}
    <div
      class="rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative"
      role="alert"
    >
      {error}
    </div>
  {:else if evaluation}
    <!-- Auto Checks -->
    <div>
      <h4
        class="mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted"
      >
        Auto Checks
      </h4>
      <div class="space-y-2">
        {#each evaluation.checks.filter((c) => c.type === "AUTO") as check}
          <div
            class="flex items-start gap-3 rounded border border-border-default bg-bg-elevated px-4 py-3"
            data-testid="auto-check"
          >
            <span class="mt-0.5 flex-shrink-0">
              {#if check.status === "PASS"}
                <Check size={16} strokeWidth={2} class="text-profit" aria-hidden="true" />
              {:else}
                <X size={16} strokeWidth={2} class="text-loss" aria-hidden="true" />
              {/if}
            </span>
            <div class="flex-1">
              <p class="text-sm font-medium text-text-primary">{check.label}</p>
              <p class="text-xs text-text-secondary font-mono">{check.detail}</p>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Manual Checks -->
    {#if evaluation.checks.filter((c) => c.type === "MANUAL").length > 0}
      <div>
        <h4
          class="mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted"
        >
          Manual Checks
        </h4>
        <div class="space-y-2">
          {#each evaluation.checks.filter((c) => c.type === "MANUAL") as check}
            <label
              class="flex items-start gap-3 rounded border border-border-default bg-bg-elevated px-4 py-3 cursor-pointer"
              data-testid="manual-check"
            >
              <input
                type="checkbox"
                class="mt-1 h-4 w-4 rounded border-border-default text-green-700 focus-ring"
                checked={check.status === "PASS"}
                disabled={toggling}
                onchange={(e) => handleToggle(check, e.currentTarget.checked)}
              />
              <div class="flex-1">
                <p class="text-sm font-medium text-text-primary">
                  {check.label}
                </p>
              </div>
            </label>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Status + Reset -->
    <div class="flex items-center justify-between">
      <p
        class="text-sm {evaluation.allPassed
          ? 'text-profit font-medium'
          : 'text-text-secondary'}"
      >
        {#if evaluation.allPassed}
          All checks passed
        {:else}
          {evaluation.checks.filter((c) => c.status === "PASS").length} / {evaluation
            .checks.length} checks passed
        {/if}
      </p>
      <Button variant="ghost" size="sm" onclick={handleReset}>
        Reset
      </Button>
    </div>

    <!-- Suggestion Card -->
    {#if evaluation.allPassed && evaluation.suggestion}
      <SuggestionCard suggestion={evaluation.suggestion} />
    {/if}
  {/if}
</div>
