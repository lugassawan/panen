<script lang="ts">
import { t } from "../../i18n";
import { SHORTCUT_REGISTRY, type ShortcutCategory } from "../shortcuts-registry";
import Modal from "./Modal.svelte";

let { open, onClose }: { open: boolean; onClose: () => void } = $props();

const categories: { key: ShortcutCategory; label: string }[] = [
  { key: "global", label: "shortcuts.global" },
  { key: "navigation", label: "shortcuts.navigation" },
  { key: "action", label: "shortcuts.actions" },
];

const grouped = $derived.by(() => {
  const map = new Map<ShortcutCategory, typeof SHORTCUT_REGISTRY>();
  for (const cat of categories) {
    map.set(cat.key, []);
  }
  for (const shortcut of SHORTCUT_REGISTRY) {
    map.get(shortcut.category)?.push(shortcut);
  }
  return map;
});
</script>

<Modal {open} title={t("shortcuts.title")} size="md" onClose={onClose}>
  <div class="space-y-5">
    {#each categories as cat}
      {@const items = grouped.get(cat.key) ?? []}
      {#if items.length > 0}
        <div>
          <h4 class="mb-2 text-xs font-semibold uppercase tracking-wider text-text-muted">{t(cat.label)}</h4>
          <ul class="space-y-1">
            {#each items as shortcut}
              <li class="flex items-center justify-between py-1.5">
                <span class="text-sm text-text-secondary">{t(shortcut.label)}</span>
                <kbd class="rounded border border-border-default bg-bg-tertiary px-2 py-0.5 font-mono text-xs text-text-primary">{shortcut.keys}</kbd>
              </li>
            {/each}
          </ul>
        </div>
      {/if}
    {/each}
  </div>
</Modal>
