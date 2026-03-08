<script lang="ts">
import { untrack } from "svelte";
import { t } from "../../i18n";
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";
import Modal from "../../lib/components/Modal.svelte";
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

let normal = $state(untrack(() => String(settings.normal)));
let crash = $state(untrack(() => String(settings.crash)));
let extreme = $state(untrack(() => String(settings.extreme)));

const sum = $derived(Number(normal) + Number(crash) + Number(extreme));
const isValid = $derived(sum === 100);

function handleSave() {
  if (isValid) {
    onSave(Number(normal), Number(crash), Number(extreme));
  }
}
</script>

<Modal title={t("crashPlaybook.deploymentTitle")} onClose={onClose} size="sm">
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

  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" size="sm" onclick={onClose}>{t("common.cancel")}</Button>
      <Button variant="primary" size="sm" onclick={handleSave} disabled={!isValid}>{t("common.save")}</Button>
    </div>
  {/snippet}
</Modal>
