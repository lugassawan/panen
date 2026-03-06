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
import { t } from "../i18n";
import SyncIndicator from "../lib/components/SyncIndicator.svelte";
import { alerts } from "../lib/stores/alerts.svelte";
import type { Page } from "../lib/types";

let { currentPage, onNavigate }: { currentPage: Page; onNavigate: (page: Page) => void } = $props();

const navItems: { page: Page; labelKey: string; icon: Component }[] = [
  { page: "lookup", labelKey: "nav.lookup", icon: Search },
  { page: "watchlist", labelKey: "nav.watchlist", icon: Bookmark },
  { page: "screener", labelKey: "nav.screener", icon: Filter },
  { page: "portfolio", labelKey: "nav.portfolio", icon: Briefcase },
  { page: "payday", labelKey: "nav.payday", icon: CalendarDays },
  { page: "crashplaybook", labelKey: "nav.crashPlaybook", icon: Shield },
  { page: "alerts", labelKey: "nav.alerts", icon: Bell },
  { page: "brokerage", labelKey: "nav.brokerage", icon: Landmark },
  { page: "settings", labelKey: "nav.settings", icon: Settings },
];

$effect(() => {
  alerts.loadCount();
});
</script>

<nav class="flex h-full w-sidebar flex-col border-r border-border-default bg-bg-secondary" aria-label="Main navigation">
  <div class="flex items-center gap-2.5 p-4">
    <img src="/favicon.svg" alt="" class="h-7 w-7" aria-hidden="true" />
    <h1 class="text-lg font-bold tracking-tight text-green-700">Panen</h1>
  </div>

  <ul class="flex flex-1 flex-col" role="list">
    {#each navItems as item}
      {@const isSettings = item.page === "settings"}
      {#if isSettings}
        <li class="mt-auto border-t border-border-default">
          <SyncIndicator />
        </li>
      {/if}
      <li>
        <button
          onclick={() => onNavigate(item.page)}
          class="flex w-full items-center gap-3 rounded-md px-4 py-3 text-sm font-medium transition-fast focus-ring {currentPage === item.page
            ? 'bg-green-100 text-green-800 font-semibold dark:bg-green-900/30 dark:text-green-400'
            : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
          aria-current={currentPage === item.page ? "page" : undefined}
        >
          <item.icon size={20} strokeWidth={1.5} class="shrink-0" />
          {t(item.labelKey)}
          {#if item.page === "alerts" && alerts.activeCount > 0}
            <span class="ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-negative px-1.5 text-xs font-bold text-white">
              {alerts.activeCount}
            </span>
          {/if}
        </button>
      </li>
    {/each}
  </ul>
</nav>
