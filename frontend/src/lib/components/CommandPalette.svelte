<script lang="ts">
import {
  Bell,
  Bookmark,
  Briefcase,
  CalendarDays,
  Filter,
  Landmark,
  Search,
  Settings,
  Shield,
} from "lucide-svelte";
import type { Component } from "svelte";
import { t } from "../../i18n";
import { commandPalette } from "../stores/command-palette.svelte";
import { mode } from "../stores/mode.svelte";
import type { Page } from "../types";

let { onNavigate }: { onNavigate: (page: Page) => void } = $props();

let query = $state("");
let activeIndex = $state(0);
let inputEl = $state<HTMLInputElement | null>(null);

interface CommandItem {
  id: Page;
  labelKey: string;
  icon: Component;
  shortcut: string;
}

const commands: CommandItem[] = [
  { id: "lookup", labelKey: "nav.lookup", icon: Search, shortcut: "1" },
  { id: "watchlist", labelKey: "nav.watchlist", icon: Bookmark, shortcut: "2" },
  { id: "screener", labelKey: "nav.screener", icon: Filter, shortcut: "3" },
  { id: "portfolio", labelKey: "nav.portfolio", icon: Briefcase, shortcut: "4" },
  { id: "payday", labelKey: "nav.payday", icon: CalendarDays, shortcut: "5" },
  { id: "crashplaybook", labelKey: "nav.crashPlaybook", icon: Shield, shortcut: "6" },
  { id: "alerts", labelKey: "nav.alerts", icon: Bell, shortcut: "7" },
  { id: "brokerage", labelKey: "nav.brokerage", icon: Landmark, shortcut: "8" },
  { id: "settings", labelKey: "nav.settings", icon: Settings, shortcut: "9" },
];

let filtered = $derived(
  query.trim()
    ? commands.filter((c) => t(c.labelKey).toLowerCase().includes(query.toLowerCase()))
    : commands,
);

$effect(() => {
  if (commandPalette.open) {
    query = "";
    activeIndex = 0;
    // Focus input on next tick
    setTimeout(() => inputEl?.focus(), 0);
  }
});

$effect(() => {
  // Reset active index when filtered results change
  void filtered.length;
  activeIndex = 0;
});

function select(item: CommandItem) {
  commandPalette.close();
  onNavigate(item.id);
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    e.preventDefault();
    commandPalette.close();
    return;
  }

  if (e.key === "ArrowDown") {
    e.preventDefault();
    activeIndex = (activeIndex + 1) % filtered.length;
    return;
  }

  if (e.key === "ArrowUp") {
    e.preventDefault();
    activeIndex = (activeIndex - 1 + filtered.length) % filtered.length;
    return;
  }

  if (e.key === "Enter" && filtered.length > 0) {
    e.preventDefault();
    select(filtered[activeIndex]);
  }
}
</script>

{#if commandPalette.open}
  <div class="fixed inset-0 z-50 flex items-start justify-center pt-24 bg-black/50">
    <div class="fixed inset-0" role="presentation" onclick={() => commandPalette.close()}></div>
    <div
      class="relative z-10 w-full max-w-lg rounded-lg border border-border-default bg-bg-elevated shadow-lg overflow-hidden"
      role="presentation"
      onkeydown={handleKeydown}
    >
      <div class="flex items-center gap-3 border-b border-border-default px-4 py-3">
        <Search size={16} strokeWidth={2} class="text-text-muted shrink-0" aria-hidden="true" />
        <input
          bind:this={inputEl}
          bind:value={query}
          type="text"
          placeholder={t("nav.searchPages")}
          class="flex-1 bg-transparent text-sm text-text-primary placeholder:text-text-muted outline-none"
          role="combobox"
          aria-label={t("nav.searchPages")}
          aria-expanded="true"
          aria-haspopup="listbox"
          aria-autocomplete="list"
          aria-controls="command-list"
          aria-activedescendant={filtered.length > 0 ? `cmd-${filtered[activeIndex].id}` : undefined}
        />
        <kbd class="hidden sm:inline-flex items-center gap-0.5 rounded border border-border-default bg-bg-secondary px-1.5 py-0.5 text-xs text-text-muted">
          Esc
        </kbd>
      </div>

      <ul id="command-list" role="listbox" class="max-h-72 overflow-y-auto py-2">
        {#each filtered as item, i (item.id)}
          {@const Icon = item.icon}
          <li
            id="cmd-{item.id}"
            role="option"
            aria-selected={i === activeIndex}
            class="flex items-center justify-between px-4 py-2.5 text-sm cursor-pointer transition-fast {i === activeIndex
              ? mode.config.activeHighlight
              : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
            onclick={() => select(item)}
            onkeydown={(e: KeyboardEvent) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); select(item); } }}
            onmouseenter={() => { activeIndex = i; }}
          >
            <span class="flex items-center gap-3">
              <Icon size={16} strokeWidth={1.5} aria-hidden="true" />
              {t(item.labelKey)}
            </span>
            <kbd class="rounded border border-border-default bg-bg-secondary px-1.5 py-0.5 text-xs text-text-muted">
              {"\u2318"}{item.shortcut}
            </kbd>
          </li>
        {/each}
        {#if filtered.length === 0}
          <li class="px-4 py-6 text-center text-sm text-text-muted">
            {t("nav.noResults")}
          </li>
        {/if}
      </ul>
    </div>
  </div>
{/if}
