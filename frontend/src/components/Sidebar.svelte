<script lang="ts">
import type { Page } from "../lib/types";
import Icon from "./Icon.svelte";

let { currentPage, onNavigate }: { currentPage: Page; onNavigate: (page: Page) => void } = $props();

const navItems: { page: Page; label: string; icon: string }[] = [
  {
    page: "lookup",
    label: "Stock Lookup",
    icon: "M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z",
  },
  {
    page: "portfolio",
    label: "Portfolio",
    icon: "M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z",
  },
  {
    page: "settings",
    label: "Settings",
    icon: "M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z M15 12a3 3 0 11-6 0 3 3 0 016 0z",
  },
];
</script>

<nav class="flex h-full w-sidebar flex-col border-r border-border-default bg-bg-secondary" aria-label="Main navigation">
  <div class="flex items-center gap-2.5 p-4">
    <img src="/favicon.svg" alt="" class="h-7 w-7" aria-hidden="true" />
    <h1 class="text-lg font-bold tracking-tight text-green-700">Panen</h1>
  </div>

  <ul class="flex flex-1 flex-col" role="list">
    {#each navItems as item}
      {@const isSettings = item.page === "settings"}
      <li class={isSettings ? "mt-auto" : ""}>
        <button
          onclick={() => onNavigate(item.page)}
          class="flex w-full items-center gap-3 rounded-md px-4 py-3 text-sm font-medium transition-fast focus-ring {currentPage === item.page
            ? 'bg-green-100 text-green-800 font-semibold dark:bg-green-900/30 dark:text-green-400'
            : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
          aria-current={currentPage === item.page ? "page" : undefined}
        >
          <Icon path={item.icon} />
          {item.label}
        </button>
      </li>
    {/each}
  </ul>
</nav>
