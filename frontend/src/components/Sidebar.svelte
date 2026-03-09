<script lang="ts">
import {
  ArrowLeftRight,
  Bell,
  Bookmark,
  Briefcase,
  CalendarDays,
  Filter,
  Landmark,
  LayoutDashboard,
  Receipt,
  Search,
  Settings,
  Shield,
} from "lucide-svelte";
import type { Component } from "svelte";
import { t } from "../i18n";
import SyncIndicator from "../lib/components/SyncIndicator.svelte";
import { alerts } from "../lib/stores/alerts.svelte";
import { mode } from "../lib/stores/mode.svelte";
import type { Page } from "../lib/types";

let { currentPage, onNavigate }: { currentPage: Page; onNavigate: (page: Page) => void } = $props();

type NavGroup = "overview" | "research" | "portfolio" | "account";

const navItems: { page: Page; labelKey: string; icon: Component; group: NavGroup }[] = [
  { page: "dashboard", labelKey: "nav.dashboard", icon: LayoutDashboard, group: "overview" },
  { page: "lookup", labelKey: "nav.lookup", icon: Search, group: "research" },
  { page: "watchlist", labelKey: "nav.watchlist", icon: Bookmark, group: "research" },
  { page: "screener", labelKey: "nav.screener", icon: Filter, group: "research" },
  { page: "comparison", labelKey: "nav.comparison", icon: ArrowLeftRight, group: "research" },
  { page: "portfolio", labelKey: "nav.portfolio", icon: Briefcase, group: "portfolio" },
  { page: "payday", labelKey: "nav.payday", icon: CalendarDays, group: "portfolio" },
  { page: "crashplaybook", labelKey: "nav.crashPlaybook", icon: Shield, group: "portfolio" },
  { page: "transactions", labelKey: "nav.transactions", icon: Receipt, group: "portfolio" },
  { page: "alerts", labelKey: "nav.alerts", icon: Bell, group: "account" },
  { page: "brokerage", labelKey: "nav.brokerage", icon: Landmark, group: "account" },
];

const groupOrder: NavGroup[] = ["overview", "research", "portfolio", "account"];

const groupedItems = $derived.by(() => {
  const groups = new Map<NavGroup, typeof navItems>();
  for (const group of groupOrder) {
    groups.set(group, []);
  }
  for (const item of navItems) {
    groups.get(item.group)?.push(item);
  }
  return groups;
});

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
    {#each groupOrder as group}
      <li>
        <span class="block px-3 pt-4 pb-1 text-[11px] font-semibold uppercase tracking-wider text-text-tertiary">
          {t(`nav.group.${group}`)}
        </span>
        <ul role="list">
          {#each groupedItems.get(group) ?? [] as item}
            <li>
              <button
                onclick={() => onNavigate(item.page)}
                class="flex w-full items-center gap-3 rounded-md px-4 py-3 text-sm font-medium transition-fast focus-ring {currentPage === item.page
                  ? `${mode.config.activeHighlight} font-semibold`
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
      </li>
    {/each}
    <li class="mt-auto border-t border-border-default">
      <SyncIndicator />
    </li>
    <li>
      <button
        onclick={() => onNavigate("settings")}
        class="flex w-full items-center gap-3 rounded-md px-4 py-3 text-sm font-medium transition-fast focus-ring {currentPage === 'settings'
          ? `${mode.config.activeHighlight} font-semibold`
          : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
        aria-current={currentPage === "settings" ? "page" : undefined}
      >
        <Settings size={20} strokeWidth={1.5} class="shrink-0" />
        {t("nav.settings")}
      </button>
    </li>
  </ul>
</nav>
