<script lang="ts">
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";
import type { DeploymentSettingsResponse } from "../../lib/types";

let {
  settings,
  onSave,
  onClose,
}: {
  settings: DeploymentSettingsResponse;
  onSave: (normal: number, crash: number, extreme: number) => void;
  onClose: () => void;
} = $props();

let normal = $state(String(settings.normal));
let crash = $state(String(settings.crash));
let extreme = $state(String(settings.extreme));

const sum = $derived(Number(normal) + Number(crash) + Number(extreme));
const isValid = $derived(sum === 100);

function handleSave() {
  if (isValid) {
    onSave(Number(normal), Number(crash), Number(extreme));
  }
}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions, a11y_click_events_have_key_events -->
<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={onClose}>
  <!-- svelte-ignore a11y_no_static_element_interactions, a11y_click_events_have_key_events -->
  <div class="w-full max-w-sm rounded-lg border border-border-default bg-bg-elevated p-6" onclick={(e) => e.stopPropagation()}>
    <h3 class="font-display text-lg font-semibold text-text-primary">Deployment Settings</h3>
    <p class="mt-1 text-sm text-text-secondary">Set capital deployment % per crash level. Must sum to 100%.</p>

    <div class="mt-4 space-y-3">
      <label class="block">
        <span class="block text-xs font-medium text-text-secondary mb-1">Normal Dip (%)</span>
        <Input type="number" bind:value={normal} />
      </label>
      <label class="block">
        <span class="block text-xs font-medium text-text-secondary mb-1">Crash (%)</span>
        <Input type="number" bind:value={crash} />
      </label>
      <label class="block">
        <span class="block text-xs font-medium text-text-secondary mb-1">Extreme (%)</span>
        <Input type="number" bind:value={extreme} />
      </label>
    </div>

    <div class="mt-3 text-sm {isValid ? 'text-profit' : 'text-loss'}">
      Total: {sum}% {isValid ? "" : "(must be 100%)"}
    </div>

    <div class="mt-4 flex items-center justify-end gap-3">
      <Button variant="secondary" size="sm" onclick={onClose}>Cancel</Button>
      <Button variant="primary" size="sm" onclick={handleSave} disabled={!isValid}>Save</Button>
    </div>
  </div>
</div>
