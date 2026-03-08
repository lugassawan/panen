package backend

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lugassawan/panen/backend/domain/user"
	"github.com/lugassawan/panen/backend/infra/applog"
	"github.com/lugassawan/panen/backend/infra/backup"
	brokerConfigLoader "github.com/lugassawan/panen/backend/infra/brokerconfig"
	"github.com/lugassawan/panen/backend/infra/database"
	"github.com/lugassawan/panen/backend/infra/github"
	"github.com/lugassawan/panen/backend/infra/liveconfig"
	"github.com/lugassawan/panen/backend/infra/platform"
	"github.com/lugassawan/panen/backend/infra/scraper"
	"github.com/lugassawan/panen/backend/infra/updater"
	"github.com/lugassawan/panen/backend/infra/watchlistconfig"
	"github.com/lugassawan/panen/backend/presenter"
	"github.com/lugassawan/panen/backend/usecase"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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
	*presenter.AlertHandler
	*presenter.BackupHandler
	*presenter.LogHandler
	*presenter.LiveConfigHandler
	*presenter.TransactionHandler
	*presenter.DashboardHandler
	db        *database.DB
	backup    *backup.BackupService
	dbPath    string
	backupDir string
	logDir    string
	refresh   *usecase.RefreshService
}

// repos groups all database repository instances.
type repos struct {
	user            user.Repository
	brokerage       *database.BrokerageRepo
	portfolio       *database.PortfolioRepo
	holding         *database.HoldingRepo
	buyTxn          *database.BuyTransactionRepo
	stock           *database.StockDataRepo
	watchlist       *database.WatchlistRepo
	watchlistItem   *database.WatchlistItemRepo
	peak            *database.PeakRepo
	settings        *database.SettingsRepo
	tickerCollector *database.TickerCollector
	priceHistory    *database.PriceHistoryRepo
	divHistory      *database.DividendHistoryRepo
	snapshot        *database.SnapshotRepo
	alert           *database.AlertRepo
	checklistResult *database.ChecklistResultRepo
	payday          *database.PaydayRepo
	cashFlow        *database.CashFlowRepo
	txnHistory      *database.TransactionHistoryRepo
	crashCapital    *database.CrashCapitalRepo
}

// services groups all application service instances.
type services struct {
	stocks         *usecase.StockService
	portfolios     *usecase.PortfolioService
	brokerages     *usecase.BrokerageService
	watchlists     *usecase.WatchlistService
	screener       *usecase.ScreenerService
	priceHistory   *usecase.PriceHistoryService
	divHistory     *usecase.DividendHistoryService
	refresh        *usecase.RefreshService
	dividends      *usecase.DividendService
	alerts         *usecase.AlertService
	checklists     *usecase.ChecklistService
	payday         *usecase.PaydayService
	transactions   *usecase.TransactionService
	dashboard      *usecase.DashboardService
	crashPlaybook  *usecase.CrashPlaybookService
	update         *usecase.UpdateService
	selfUpdate     *usecase.SelfUpdateService
}

// NewApp creates all handlers upfront so embedded pointers are never nil.
// Dependencies are wired later in Startup via Bind calls.
func NewApp() *App {
	return &App{
		StockHandler:            &presenter.StockHandler{},
		PortfolioHandler:        &presenter.PortfolioHandler{},
		BrokerageHandler:        &presenter.BrokerageHandler{},
		BrokerConfigHandler:     &presenter.BrokerConfigHandler{},
		WatchlistHandler:        &presenter.WatchlistHandler{},
		RefreshHandler:          &presenter.RefreshHandler{},
		ChecklistHandler:        &presenter.ChecklistHandler{},
		UpdateHandler:           &presenter.UpdateHandler{},
		PaydayHandler:           &presenter.PaydayHandler{},
		CrashPlaybookHandler:    &presenter.CrashPlaybookHandler{},
		ScreenerHandler:         &presenter.ScreenerHandler{},
		DividendHandler:         &presenter.DividendHandler{},
		PriceHistoryHandler:     &presenter.PriceHistoryHandler{},
		DividendCalendarHandler: &presenter.DividendCalendarHandler{},
		AlertHandler:            &presenter.AlertHandler{},
		BackupHandler:           &presenter.BackupHandler{},
		LogHandler:              &presenter.LogHandler{},
		LiveConfigHandler:       &presenter.LiveConfigHandler{},
		TransactionHandler:      &presenter.TransactionHandler{},
		DashboardHandler:        &presenter.DashboardHandler{},
		backup:                  backup.NewBackupService(),
	}
}

// Startup initialises infrastructure, constructs services, and ensures a default user profile.
func (a *App) Startup(ctx context.Context) {
	a.initLogging(ctx)

	dataDir := a.initDataDir(ctx)

	db, err := database.Open(a.dbPath)
	if err != nil {
		wailsRuntime.LogFatalf(ctx, "open database: %v", err)
	}
	a.db = db

	if err := database.Migrate(ctx, db.Conn()); err != nil {
		wailsRuntime.LogFatalf(ctx, "migrate database: %v", err)
	}

	if err := a.backup.RunDaily(a.dbPath, a.backupDir); err != nil {
		wailsRuntime.LogWarningf(ctx, "daily backup: %v", err)
	}
	if err := a.backup.Cleanup(a.backupDir, 7); err != nil {
		wailsRuntime.LogWarningf(ctx, "backup cleanup: %v", err)
	}

	r := a.initRepos(db.Conn())

	profileID, err := ensureDefaultUser(ctx, r.user)
	if err != nil {
		wailsRuntime.LogFatalf(ctx, "ensure default user: %v", err)
	}

	a.initDebugLogging(ctx, r.settings)

	wailsEmitter := presenter.NewWailsEmitter(ctx)
	yahoo := scraper.NewYahoo()
	sectorRegistry := watchlistconfig.NewSectorRegistry()
	liveDeps := liveconfig.Deps{
		Settings: r.settings,
		Emitter:  wailsEmitter,
	}

	brokerLoader := brokerConfigLoader.NewLoader(dataDir, liveDeps)
	brokerResult := brokerLoader.Load(ctx)

	indexLoader := watchlistconfig.NewIndexLoader(dataDir, liveDeps)
	indexResult := indexLoader.Load(ctx)
	swappableIndexReg := watchlistconfig.NewSwappableIndexRegistry(indexResult.Data)

	svc := a.initServices(r, yahoo, wailsEmitter, sectorRegistry, swappableIndexReg)

	a.bindHandlers(ctx, svc, r, profileID, sectorRegistry)

	a.Init(ctx)
	a.RegisterLoader("brokers", brokerLoader, func(_ context.Context) {
		configs := brokerLoader.LastResult().Data
		a.BrokerConfigHandler.Bind(configs)
		if _, err := svc.brokerages.SyncFeesFromConfig(ctx, profileID, configs); err != nil {
			applog.Warn("broker fee sync", err, nil)
		}
	})
	a.RegisterLoader("indices", indexLoader, func(_ context.Context) {
		swappableIndexReg.Swap(indexLoader.LastResult().Data)
	})

	a.BrokerConfigHandler.Bind(brokerResult.Data)

	a.BackupHandler.Bind(ctx, a.backup, a.dbPath, a.backupDir)
	a.PortfolioHandler.BindBackup(a.backup, a.dbPath, a.backupDir)

	svc.refresh.Start(ctx)

	go a.CheckForUpdateOnStartup()
	go svc.selfUpdate.CleanupPreviousUpdate()
}

// Shutdown stops background services and closes the database connection.
func (a *App) Shutdown(ctx context.Context) {
	if a.refresh != nil {
		a.refresh.Stop()
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			wailsRuntime.LogErrorf(ctx, "close database: %v", err)
		}
	}
	applog.Shutdown()
}

// initLogging sets up file-based structured logging.
func (a *App) initLogging(ctx context.Context) {
	logDir, err := platform.LogDir()
	if err != nil {
		wailsRuntime.LogFatalf(ctx, "resolve log dir: %v", err)
	}
	if err := applog.Init(logDir); err != nil {
		wailsRuntime.LogFatalf(ctx, "init logging: %v", err)
	}
	a.logDir = logDir
}

// initDataDir resolves and creates the data/backup directories,
// attempts database recovery if needed, and sets a.dbPath and a.backupDir.
func (a *App) initDataDir(ctx context.Context) string {
	dataDir, err := platform.DataDir()
	if err != nil {
		wailsRuntime.LogFatalf(ctx, "resolve data dir: %v", err)
	}
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		wailsRuntime.LogFatalf(ctx, "create data dir: %v", err)
	}

	backupDir, err := platform.BackupDir()
	if err != nil {
		wailsRuntime.LogFatalf(ctx, "resolve backup dir: %v", err)
	}
	a.backupDir = backupDir
	a.dbPath = filepath.Join(dataDir, "panen.db")

	if restored, err := backup.TryRecover(dataDir, backupDir); err != nil {
		wailsRuntime.LogWarningf(ctx, "recovery check: %v", err)
	} else if restored != "" {
		wailsRuntime.LogInfof(ctx, "database restored from backup: %s", restored)
	}

	return dataDir
}

// initRepos creates all database repository instances from the given connection.
func (a *App) initRepos(conn *sql.DB) repos {
	return repos{
		user:            database.NewUserRepo(conn),
		brokerage:       database.NewBrokerageRepo(conn),
		portfolio:       database.NewPortfolioRepo(conn),
		holding:         database.NewHoldingRepo(conn),
		buyTxn:          database.NewBuyTransactionRepo(conn),
		stock:           database.NewStockDataRepo(conn),
		watchlist:       database.NewWatchlistRepo(conn),
		watchlistItem:   database.NewWatchlistItemRepo(conn),
		peak:            database.NewPeakRepo(conn),
		settings:        database.NewSettingsRepo(conn),
		tickerCollector: database.NewTickerCollector(conn),
		priceHistory:    database.NewPriceHistoryRepo(conn),
		divHistory:      database.NewDividendHistoryRepo(conn),
		snapshot:        database.NewSnapshotRepo(conn),
		alert:           database.NewAlertRepo(conn),
		checklistResult: database.NewChecklistResultRepo(conn),
		payday:          database.NewPaydayRepo(conn),
		cashFlow:        database.NewCashFlowRepo(conn),
		txnHistory:      database.NewTransactionHistoryRepo(conn),
		crashCapital:    database.NewCrashCapitalRepo(conn),
	}
}

// initDebugLogging enables debug-level logging if the setting is persisted,
// and rotates old log files.
func (a *App) initDebugLogging(ctx context.Context, settingsRepo *database.SettingsRepo) {
	if dbg, err := settingsRepo.GetSetting(ctx, applog.DebugLoggingKey); err == nil && dbg == "1" {
		applog.SetLevel(slog.LevelDebug)
	}
	if err := applog.RotateLogs(a.logDir, applog.LogRetentionDays); err != nil {
		applog.Warn("log rotation", err, nil)
	}
}

// initServices constructs all application services from repositories and infrastructure.
func (a *App) initServices(
	r repos,
	yahoo *scraper.Yahoo,
	emitter *presenter.WailsEmitter,
	sectorRegistry *watchlistconfig.SectorRegistry,
	indexReg *watchlistconfig.SwappableIndexRegistry,
) services {
	stocks := usecase.NewStockService(r.stock, yahoo)
	portfolios := usecase.NewPortfolioService(
		r.portfolio, r.holding, r.buyTxn, r.brokerage, r.stock, r.peak,
	)
	brokerages := usecase.NewBrokerageService(r.brokerage, r.portfolio, emitter)
	watchlists := usecase.NewWatchlistService(
		r.watchlist, r.watchlistItem, r.stock, indexReg, sectorRegistry,
	)
	screener := usecase.NewScreenerService(r.stock, indexReg, sectorRegistry)
	priceHistory := usecase.NewPriceHistoryService(r.priceHistory, yahoo)
	divHistory := usecase.NewDividendHistoryService(
		r.divHistory, yahoo, r.holding, r.portfolio, r.stock,
	)
	refreshSvc := usecase.NewRefreshService(
		r.stock, yahoo, r.settings, r.tickerCollector, emitter, r.snapshot, r.alert,
	)
	a.refresh = refreshSvc

	dividends := usecase.NewDividendService(r.portfolio, r.holding, r.stock)
	alertSvc := usecase.NewAlertService(r.alert)
	checklists := usecase.NewChecklistService(
		r.checklistResult, r.portfolio, r.holding, r.brokerage, r.stock, alertSvc,
	)
	paydaySvc := usecase.NewPaydayService(r.payday, r.cashFlow, r.portfolio, r.settings)
	transactions := usecase.NewTransactionService(r.txnHistory)
	dashboard := usecase.NewDashboardService(
		r.portfolio, r.holding, r.stock, r.payday, r.txnHistory, sectorRegistry,
	)
	crashPlaybook := usecase.NewCrashPlaybookService(
		r.stock, yahoo, r.portfolio, r.holding, r.crashCapital, r.settings, refreshSvc,
	)

	ghClient := github.NewClient()
	updateChecker := &releaseCheckerAdapter{client: ghClient}
	updateSvc := usecase.NewUpdateService(updateChecker, Version())

	downloader := updater.NewDownloader(ghClient)
	verifier := &updater.SHA256Verifier{}
	extractor := &updater.Extractor{}
	installer := updater.NewPlatformInstaller()
	selfUpdateSvc := usecase.NewSelfUpdateService(
		updateChecker, downloader, verifier, extractor,
		installer, emitter, Version(),
	)

	return services{
		stocks:        stocks,
		portfolios:    portfolios,
		brokerages:    brokerages,
		watchlists:    watchlists,
		screener:      screener,
		priceHistory:  priceHistory,
		divHistory:    divHistory,
		refresh:       refreshSvc,
		dividends:     dividends,
		alerts:        alertSvc,
		checklists:    checklists,
		payday:        paydaySvc,
		transactions:  transactions,
		dashboard:     dashboard,
		crashPlaybook: crashPlaybook,
		update:        updateSvc,
		selfUpdate:    selfUpdateSvc,
	}
}

// bindHandlers wires all presenter handlers to their service dependencies.
func (a *App) bindHandlers(
	ctx context.Context,
	svc services,
	r repos,
	profileID string,
	sectorRegistry *watchlistconfig.SectorRegistry,
) {
	a.StockHandler.Bind(ctx, svc.stocks)
	a.PortfolioHandler.Bind(ctx, svc.portfolios, sectorRegistry)
	a.DividendHandler.Bind(ctx, svc.dividends)
	a.BrokerageHandler.Bind(ctx, profileID, svc.brokerages)
	a.WatchlistHandler.Bind(ctx, profileID, svc.watchlists)
	a.RefreshHandler.Bind(ctx, svc.refresh, r.settings)
	a.ScreenerHandler.Bind(ctx, svc.screener)
	a.PriceHistoryHandler.Bind(ctx, svc.priceHistory)
	a.DividendCalendarHandler.Bind(ctx, svc.divHistory)
	a.AlertHandler.Bind(ctx, svc.alerts)
	a.ChecklistHandler.Bind(ctx, svc.checklists)
	a.PaydayHandler.Bind(ctx, svc.payday)
	a.TransactionHandler.Bind(ctx, svc.transactions)
	a.DashboardHandler.Bind(ctx, svc.dashboard)
	a.CrashPlaybookHandler.Bind(ctx, svc.crashPlaybook, r.portfolio)
	a.UpdateHandler.Bind(ctx, svc.update, svc.selfUpdate, r.settings)
	a.LogHandler.Bind(ctx, r.settings, a.logDir)
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
	assets := make([]usecase.ReleaseAsset, len(rel.Assets))
	for i, ga := range rel.Assets {
		assets[i] = usecase.ReleaseAsset{
			Name:        ga.Name,
			DownloadURL: ga.BrowserDownloadURL,
			Size:        ga.Size,
		}
	}
	return &usecase.ReleaseInfo{
		Version:    rel.Version(),
		ReleaseURL: rel.HTMLURL,
		Assets:     assets,
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
