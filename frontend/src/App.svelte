<script lang="ts">
import Sidebar from "./components/Sidebar.svelte";
import type { Page } from "./lib/types";
import PortfolioPage from "./pages/PortfolioPage.svelte";
import SettingsPage from "./pages/SettingsPage.svelte";
import StockLookupPage from "./pages/StockLookupPage.svelte";

let currentPage = $state<Page>("lookup");
</script>

<div class="flex h-screen bg-neutral-950 text-white">
  <Sidebar {currentPage} onNavigate={(page) => currentPage = page} />

  <main class="flex-1 overflow-y-auto">
    {#if currentPage === "lookup"}
      <StockLookupPage />
    {:else if currentPage === "portfolio"}
      <PortfolioPage />
    {:else if currentPage === "settings"}
      <SettingsPage />
    {/if}
  </main>
</div>
