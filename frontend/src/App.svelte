<script lang="ts">
import Sidebar from "./components/Sidebar.svelte";
import { t } from "./i18n";
import CommandPalette from "./lib/components/CommandPalette.svelte";
import ToastContainer from "./lib/components/ToastContainer.svelte";
import UpdateDialog from "./lib/components/UpdateDialog.svelte";
import { handleGlobalShortcut } from "./lib/shortcuts";
import { commandPalette } from "./lib/stores/command-palette.svelte";
import { theme } from "./lib/stores/theme.svelte";
import type { Page } from "./lib/types";
import AlertsPage from "./pages/alerts/AlertsPage.svelte";
import BrokeragePage from "./pages/brokerage/BrokeragePage.svelte";
import ComparisonPage from "./pages/comparison/ComparisonPage.svelte";
import CrashPlaybookPage from "./pages/crashplaybook/CrashPlaybookPage.svelte";
import DashboardPage from "./pages/dashboard/DashboardPage.svelte";
import PaydayPage from "./pages/payday/PaydayPage.svelte";
import PortfolioPage from "./pages/portfolio/PortfolioPage.svelte";
import ScreenerPage from "./pages/screener/ScreenerPage.svelte";
import SettingsPage from "./pages/settings/SettingsPage.svelte";
import StockLookupPage from "./pages/stock/StockLookupPage.svelte";
import TransactionHistoryPage from "./pages/transactions/TransactionHistoryPage.svelte";
import WatchlistPage from "./pages/watchlist/WatchlistPage.svelte";

let currentPage = $state<Page>("dashboard");

function navigateTo(page: Page) {
  currentPage = page;
}
</script>

<svelte:window onkeydown={(e) => handleGlobalShortcut(e, { onNavigate: navigateTo, onToggleCommandPalette: () => commandPalette.toggle() })} />

<div class="flex h-screen" data-theme={theme.current}>
  <a href="#main-content" class="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[100] focus:bg-bg-elevated focus:text-text-primary focus:px-4 focus:py-2 focus:rounded-lg focus:shadow-lg focus:ring-2 focus:ring-accent">
    {t("a11y.skipToContent")}
  </a>
  <Sidebar {currentPage} onNavigate={navigateTo} />

  <main id="main-content" tabindex="-1" class="flex-1 overflow-y-auto">
    {#if currentPage === "dashboard"}
      <DashboardPage onNavigate={navigateTo} />
    {:else if currentPage === "lookup"}
      <StockLookupPage />
    {:else if currentPage === "watchlist"}
      <WatchlistPage />
    {:else if currentPage === "screener"}
      <ScreenerPage />
    {:else if currentPage === "comparison"}
      <ComparisonPage />
    {:else if currentPage === "portfolio"}
      <PortfolioPage />
    {:else if currentPage === "payday"}
      <PaydayPage />
    {:else if currentPage === "crashplaybook"}
      <CrashPlaybookPage />
    {:else if currentPage === "transactions"}
      <TransactionHistoryPage />
    {:else if currentPage === "alerts"}
      <AlertsPage />
    {:else if currentPage === "brokerage"}
      <BrokeragePage />
    {:else if currentPage === "settings"}
      <SettingsPage />
    {/if}
  </main>
</div>

<ToastContainer />
<CommandPalette onNavigate={navigateTo} />
<UpdateDialog />
