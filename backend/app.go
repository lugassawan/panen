package backend

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lugassawan/panen/backend/domain/user"
	brokerConfigLoader "github.com/lugassawan/panen/backend/infra/brokerconfig"
	"github.com/lugassawan/panen/backend/infra/database"
	"github.com/lugassawan/panen/backend/infra/github"
	"github.com/lugassawan/panen/backend/infra/platform"
	"github.com/lugassawan/panen/backend/infra/scraper"
	"github.com/lugassawan/panen/backend/infra/watchlistconfig"
	"github.com/lugassawan/panen/backend/presenter"
	"github.com/lugassawan/panen/backend/usecase"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound application controller.
// Handler methods are promoted via embedding so Wails can bind them.
type App struct {
	*presenter.StockHandler
	*presenter.PortfolioHandler
	*presenter.BrokerageHandler
	*presenter.BrokerConfigHandler
	*presenter.WatchlistHandler
	*presenter.RefreshHandler
	*presenter.ChecklistHandler
	*presenter.UpdateHandler
	*presenter.PaydayHandler
	*presenter.CrashPlaybookHandler
	*presenter.ScreenerHandler
	*presenter.DividendHandler
	*presenter.PriceHistoryHandler
	*presenter.DividendCalendarHandler
	db      *database.DB
	refresh *usecase.RefreshService
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// Startup initialises infrastructure, constructs services, and ensures a default user profile.
func (a *App) Startup(ctx context.Context) {
	dataDir, err := platform.DataDir()
	if err != nil {
		runtime.LogFatalf(ctx, "resolve data dir: %v", err)
	}
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		runtime.LogFatalf(ctx, "create data dir: %v", err)
	}

	db, err := database.Open(filepath.Join(dataDir, "panen.db"))
	if err != nil {
		runtime.LogFatalf(ctx, "open database: %v", err)
	}
	a.db = db

	if err := database.Migrate(ctx, db.Conn()); err != nil {
		runtime.LogFatalf(ctx, "migrate database: %v", err)
	}

	conn := db.Conn()
	userRepo := database.NewUserRepo(conn)
	brokerageRepo := database.NewBrokerageRepo(conn)
	portfolioRepo := database.NewPortfolioRepo(conn)
	holdingRepo := database.NewHoldingRepo(conn)
	buyTxnRepo := database.NewBuyTransactionRepo(conn)
	stockRepo := database.NewStockDataRepo(conn)
	watchlistRepo := database.NewWatchlistRepo(conn)
	watchlistItemRepo := database.NewWatchlistItemRepo(conn)
	yahoo := scraper.NewYahoo()

	stocks := usecase.NewStockService(stockRepo, yahoo)
	peakRepo := database.NewPeakRepo(conn)
	portfolios := usecase.NewPortfolioService(
		portfolioRepo,
		holdingRepo,
		buyTxnRepo,
		brokerageRepo,
		stockRepo,
		peakRepo,
	)
	brokerages := usecase.NewBrokerageService(brokerageRepo, portfolioRepo)

	profileID, err := ensureDefaultUser(ctx, userRepo)
	if err != nil {
		runtime.LogFatalf(ctx, "ensure default user: %v", err)
	}

	loader := brokerConfigLoader.NewLoader(dataDir)
	brokerConfigs := loader.Load(ctx)

	indexLoader := watchlistconfig.NewIndexLoader(dataDir)
	indexRegistry := indexLoader.Load(ctx)
	sectorRegistry := watchlistconfig.NewSectorRegistry()
	watchlistSvc := usecase.NewWatchlistService(
		watchlistRepo,
		watchlistItemRepo,
		stockRepo,
		indexRegistry,
		sectorRegistry,
	)

	screenerSvc := usecase.NewScreenerService(stockRepo, indexRegistry, sectorRegistry)
	a.ScreenerHandler = presenter.NewScreenerHandler(ctx, screenerSvc)

	priceHistoryRepo := database.NewPriceHistoryRepo(conn)
	priceHistorySvc := usecase.NewPriceHistoryService(priceHistoryRepo, yahoo)
	a.PriceHistoryHandler = presenter.NewPriceHistoryHandler(ctx, priceHistorySvc)

	divHistoryRepo := database.NewDividendHistoryRepo(conn)
	divHistorySvc := usecase.NewDividendHistoryService(divHistoryRepo, yahoo, holdingRepo, portfolioRepo, stockRepo)
	a.DividendCalendarHandler = presenter.NewDividendCalendarHandler(ctx, divHistorySvc)

	settingsRepo := database.NewSettingsRepo(conn)
	tickerCollector := database.NewTickerCollector(conn)
	wailsEmitter := presenter.NewWailsEmitter(ctx)

	refreshSvc := usecase.NewRefreshService(stockRepo, yahoo, settingsRepo, tickerCollector, wailsEmitter)
	a.refresh = refreshSvc

	dividendSvc := usecase.NewDividendService(portfolioRepo, holdingRepo, stockRepo)

	a.StockHandler = presenter.NewStockHandler(ctx, stocks)
	a.PortfolioHandler = presenter.NewPortfolioHandler(ctx, portfolios, sectorRegistry)
	a.DividendHandler = presenter.NewDividendHandler(ctx, dividendSvc)
	a.BrokerageHandler = presenter.NewBrokerageHandler(ctx, profileID, brokerages)
	a.BrokerConfigHandler = presenter.NewBrokerConfigHandler(brokerConfigs)
	a.WatchlistHandler = presenter.NewWatchlistHandler(ctx, profileID, watchlistSvc)
	a.RefreshHandler = presenter.NewRefreshHandler(ctx, refreshSvc, settingsRepo)

	checklistRepo := database.NewChecklistResultRepo(conn)
	checklistSvc := usecase.NewChecklistService(checklistRepo, portfolioRepo, holdingRepo, brokerageRepo, stockRepo)
	a.ChecklistHandler = presenter.NewChecklistHandler(ctx, checklistSvc)

	paydayRepo := database.NewPaydayRepo(conn)
	cashFlowRepo := database.NewCashFlowRepo(conn)
	paydaySvc := usecase.NewPaydayService(paydayRepo, cashFlowRepo, portfolioRepo, settingsRepo)
	a.PaydayHandler = presenter.NewPaydayHandler(ctx, paydaySvc)

	crashCapitalRepo := database.NewCrashCapitalRepo(conn)
	crashPlaybookSvc := usecase.NewCrashPlaybookService(
		stockRepo,
		yahoo,
		portfolioRepo,
		holdingRepo,
		crashCapitalRepo,
		settingsRepo,
		refreshSvc,
	)
	a.CrashPlaybookHandler = presenter.NewCrashPlaybookHandler(ctx, crashPlaybookSvc, portfolioRepo)

	ghClient := github.NewClient()
	updateChecker := &releaseCheckerAdapter{client: ghClient}
	updateSvc := usecase.NewUpdateService(updateChecker, Version())
	a.UpdateHandler = presenter.NewUpdateHandler(ctx, updateSvc, settingsRepo)

	refreshSvc.Start(ctx)

	go a.CheckForUpdateOnStartup()
}

// Shutdown stops background services and closes the database connection.
func (a *App) Shutdown(ctx context.Context) {
	if a.refresh != nil {
		a.refresh.Stop()
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			runtime.LogErrorf(ctx, "close database: %v", err)
		}
	}
}

// releaseCheckerAdapter bridges github.Client to usecase.ReleaseChecker.
type releaseCheckerAdapter struct {
	client *github.Client
}

func (a *releaseCheckerAdapter) LatestRelease(ctx context.Context) (*usecase.ReleaseInfo, error) {
	rel, err := a.client.LatestRelease(ctx)
	if err != nil {
		return nil, err
	}
	return &usecase.ReleaseInfo{
		Version:    rel.Version(),
		ReleaseURL: rel.HTMLURL,
	}, nil
}

func ensureDefaultUser(ctx context.Context, users user.Repository) (string, error) {
	profiles, err := users.List(ctx)
	if err != nil {
		return "", err
	}
	if len(profiles) > 0 {
		return profiles[0].ID, nil
	}

	p := user.NewProfile(platform.Username())
	if err := users.Create(ctx, p); err != nil {
		return "", err
	}
	return p.ID, nil
}
