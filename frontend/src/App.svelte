<script lang="ts">
import Sidebar from "./components/Sidebar.svelte";
import { theme } from "./lib/stores/theme.svelte";
import type { Page } from "./lib/types";
import BrokeragePage from "./pages/brokerage/BrokeragePage.svelte";
import CrashPlaybookPage from "./pages/crashplaybook/CrashPlaybookPage.svelte";
import PaydayPage from "./pages/payday/PaydayPage.svelte";
import PortfolioPage from "./pages/portfolio/PortfolioPage.svelte";
import ScreenerPage from "./pages/screener/ScreenerPage.svelte";
import SettingsPage from "./pages/settings/SettingsPage.svelte";
import StockLookupPage from "./pages/stock/StockLookupPage.svelte";
import WatchlistPage from "./pages/watchlist/WatchlistPage.svelte";

let currentPage = $state<Page>("lookup");
</script>

<div class="flex h-screen" data-theme={theme.current}>
  <Sidebar {currentPage} onNavigate={(page) => currentPage = page} />

  <main class="flex-1 overflow-y-auto">
    {#if currentPage === "lookup"}
      <StockLookupPage />
    {:else if currentPage === "watchlist"}
      <WatchlistPage />
    {:else if currentPage === "screener"}
      <ScreenerPage />
    {:else if currentPage === "portfolio"}
      <PortfolioPage />
    {:else if currentPage === "payday"}
      <PaydayPage />
    {:else if currentPage === "crashplaybook"}
      <CrashPlaybookPage />
    {:else if currentPage === "brokerage"}
      <BrokeragePage />
    {:else if currentPage === "settings"}
      <SettingsPage />
    {/if}
  </main>
</div>
