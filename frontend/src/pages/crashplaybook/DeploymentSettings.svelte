<script lang="ts">
import { untrack } from "svelte";
import { t } from "../../i18n";
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

let dialogEl = $state<HTMLDivElement | null>(null);
let normal = $state(untrack(() => String(settings.normal)));
let crash = $state(untrack(() => String(settings.crash)));
let extreme = $state(untrack(() => String(settings.extreme)));

$effect(() => {
  dialogEl?.focus();
});

const sum = $derived(Number(normal) + Number(crash) + Number(extreme));
const isValid = $derived(sum === 100);

function handleSave() {
  if (isValid) {
    onSave(Number(normal), Number(crash), Number(extreme));
  }
}
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
  <div class="fixed inset-0" role="presentation" onclick={onClose}></div>
  <div
    bind:this={dialogEl}
    class="relative z-10 w-full max-w-sm rounded-lg border border-border-default bg-bg-elevated p-6"
    role="dialog"
    aria-modal="true"
    aria-labelledby="deploy-settings-title"
    tabindex="-1"
    onkeydown={(e) => { if (e.key === "Escape") onClose(); }}
  >
    <h3 id="deploy-settings-title" class="font-display text-lg font-semibold text-text-primary">{t("crashPlaybook.deploymentTitle")}</h3>
    <p class="mt-1 text-sm text-text-secondary">{t("crashPlaybook.deploymentDesc")}</p>

    <div class="mt-4 space-y-3">
      <label class="block">
        <span class="block text-xs font-medium text-text-secondary mb-1">{t("crashPlaybook.normalDipPercent")}</span>
        <Input type="number" bind:value={normal} />
      </label>
      <label class="block">
        <span class="block text-xs font-medium text-text-secondary mb-1">{t("crashPlaybook.crashPercent")}</span>
        <Input type="number" bind:value={crash} />
      </label>
      <label class="block">
        <span class="block text-xs font-medium text-text-secondary mb-1">{t("crashPlaybook.extremePercent")}</span>
        <Input type="number" bind:value={extreme} />
      </label>
    </div>

    <div class="mt-3 text-sm {isValid ? 'text-profit' : 'text-loss'}">
      {isValid ? `Total: ${sum}%` : t("crashPlaybook.deploymentTotal", { sum: String(sum) })}
    </div>

    <div class="mt-4 flex items-center justify-end gap-3">
      <Button variant="secondary" size="sm" onclick={onClose}>{t("common.cancel")}</Button>
      <Button variant="primary" size="sm" onclick={handleSave} disabled={!isValid}>{t("common.save")}</Button>
    </div>
  </div>
</div>
