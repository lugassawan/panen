<script lang="ts">
import Sidebar from "./components/Sidebar.svelte";
import { theme } from "./lib/stores/theme.svelte";
import type { Page } from "./lib/types";
import BrokeragePage from "./pages/BrokeragePage.svelte";
import PortfolioPage from "./pages/PortfolioPage.svelte";
import SettingsPage from "./pages/SettingsPage.svelte";
import StockLookupPage from "./pages/StockLookupPage.svelte";
import WatchlistPage from "./pages/WatchlistPage.svelte";

let currentPage = $state<Page>("lookup");
</script>

<div class="flex h-screen" data-theme={theme.current}>
  <Sidebar {currentPage} onNavigate={(page) => currentPage = page} />

  <main class="flex-1 overflow-y-auto">
    {#if currentPage === "lookup"}
      <StockLookupPage />
    {:else if currentPage === "watchlist"}
      <WatchlistPage />
    {:else if currentPage === "portfolio"}
      <PortfolioPage />
    {:else if currentPage === "brokerage"}
      <BrokeragePage />
    {:else if currentPage === "settings"}
      <SettingsPage />
    {/if}
  </main>
</div>
