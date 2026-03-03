<script lang="ts">
import ThemeToggle from "../lib/components/ThemeToggle.svelte";
import { theme } from "../lib/stores/theme.svelte";
</script>

<div class="mx-auto max-w-lg px-4 py-8">
  <h2 class="mb-6 text-xl font-semibold text-text-primary">Settings</h2>

  <div class="space-y-6">
    <div>
      <label class="mb-1 block text-sm text-text-secondary" for="language">Language</label>
      <select
        id="language"
        disabled
        class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary opacity-60"
      >
        <option>English</option>
        <option>Bahasa Indonesia</option>
      </select>
    </div>

    <div>
      <p class="mb-1 text-sm text-text-secondary">Theme</p>
      <div class="flex items-center gap-3">
        <ThemeToggle />
        <span class="text-sm text-text-tertiary capitalize">{theme.preference}</span>
      </div>
    </div>

    <div>
      <p class="mb-1 text-sm text-text-secondary">Data Refresh</p>
      <p class="text-sm text-text-tertiary">
        Stock data is fetched on demand from Yahoo Finance. Automatic refresh is not yet available.
      </p>
    </div>
  </div>

  <p class="mt-6 text-xs text-text-muted">Language selection coming in a future update</p>
</div>
