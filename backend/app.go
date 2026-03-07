package backend

import (
	"context"
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

	db, err := database.Open(a.dbPath)
	if err != nil {
		wailsRuntime.LogFatalf(ctx, "open database: %v", err)
	}
	a.db = db

	if err := database.Migrate(ctx, db.Conn()); err != nil {
		wailsRuntime.LogFatalf(ctx, "migrate database: %v", err)
	}

	if err := a.backup.RunDaily(a.dbPath, backupDir); err != nil {
		wailsRuntime.LogWarningf(ctx, "daily backup: %v", err)
	}
	if err := a.backup.Cleanup(backupDir, 7); err != nil {
		wailsRuntime.LogWarningf(ctx, "backup cleanup: %v", err)
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

	wailsEmitter := presenter.NewWailsEmitter(ctx)

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
	brokerages := usecase.NewBrokerageService(brokerageRepo, portfolioRepo, wailsEmitter)

	profileID, err := ensureDefaultUser(ctx, userRepo)
	if err != nil {
		wailsRuntime.LogFatalf(ctx, "ensure default user: %v", err)
	}

	settingsRepo := database.NewSettingsRepo(conn)

	if dbg, err := settingsRepo.GetSetting(ctx, applog.DebugLoggingKey); err == nil && dbg == "1" {
		applog.SetLevel(slog.LevelDebug)
	}
	if err := applog.RotateLogs(a.logDir, applog.LogRetentionDays); err != nil {
		applog.Warn("log rotation", err, nil)
	}
	a.LogHandler.Bind(ctx, settingsRepo, a.logDir)

	tickerCollector := database.NewTickerCollector(conn)

	liveDeps := liveconfig.Deps{
		Settings: settingsRepo,
		Emitter:  wailsEmitter,
	}

	brokerLoader := brokerConfigLoader.NewLoader(dataDir, liveDeps)
	brokerResult := brokerLoader.Load(ctx)

	indexLoader := watchlistconfig.NewIndexLoader(dataDir, liveDeps)
	indexResult := indexLoader.Load(ctx)
	sectorRegistry := watchlistconfig.NewSectorRegistry()
	swappableIndexReg := watchlistconfig.NewSwappableIndexRegistry(indexResult.Data)
	watchlistSvc := usecase.NewWatchlistService(
		watchlistRepo,
		watchlistItemRepo,
		stockRepo,
		swappableIndexReg,
		sectorRegistry,
	)

	screenerSvc := usecase.NewScreenerService(stockRepo, swappableIndexReg, sectorRegistry)
	a.ScreenerHandler.Bind(ctx, screenerSvc)

	priceHistoryRepo := database.NewPriceHistoryRepo(conn)
	priceHistorySvc := usecase.NewPriceHistoryService(priceHistoryRepo, yahoo)
	a.PriceHistoryHandler.Bind(ctx, priceHistorySvc)

	divHistoryRepo := database.NewDividendHistoryRepo(conn)
	divHistorySvc := usecase.NewDividendHistoryService(divHistoryRepo, yahoo, holdingRepo, portfolioRepo, stockRepo)
	a.DividendCalendarHandler.Bind(ctx, divHistorySvc)

	snapshotRepo := database.NewSnapshotRepo(conn)
	alertRepo := database.NewAlertRepo(conn)

	refreshSvc := usecase.NewRefreshService(
		stockRepo, yahoo, settingsRepo, tickerCollector, wailsEmitter, snapshotRepo, alertRepo,
	)
	a.refresh = refreshSvc

	dividendSvc := usecase.NewDividendService(portfolioRepo, holdingRepo, stockRepo)

	a.StockHandler.Bind(ctx, stocks)
	a.PortfolioHandler.Bind(ctx, portfolios, sectorRegistry)
	a.DividendHandler.Bind(ctx, dividendSvc)
	a.BrokerageHandler.Bind(ctx, profileID, brokerages)
	a.BrokerConfigHandler.Bind(brokerResult.Data)
	a.WatchlistHandler.Bind(ctx, profileID, watchlistSvc)
	a.RefreshHandler.Bind(ctx, refreshSvc, settingsRepo)

	a.Init(ctx)
	a.RegisterLoader("brokers", brokerLoader, func(_ context.Context) {
		configs := brokerLoader.LastResult().Data
		a.BrokerConfigHandler.Bind(configs)
		if _, err := brokerages.SyncFeesFromConfig(ctx, profileID, configs); err != nil {
			applog.Warn("broker fee sync", err, nil)
		}
	})
	a.RegisterLoader("indices", indexLoader, func(_ context.Context) {
		swappableIndexReg.Swap(indexLoader.LastResult().Data)
	})

	alertSvc := usecase.NewAlertService(alertRepo)
	a.AlertHandler.Bind(ctx, alertSvc)

	checklistRepo := database.NewChecklistResultRepo(conn)
	checklistSvc := usecase.NewChecklistService(
		checklistRepo, portfolioRepo, holdingRepo, brokerageRepo, stockRepo, alertSvc,
	)
	a.ChecklistHandler.Bind(ctx, checklistSvc)

	paydayRepo := database.NewPaydayRepo(conn)
	cashFlowRepo := database.NewCashFlowRepo(conn)
	paydaySvc := usecase.NewPaydayService(paydayRepo, cashFlowRepo, portfolioRepo, settingsRepo)
	a.PaydayHandler.Bind(ctx, paydaySvc)

	txnHistoryRepo := database.NewTransactionHistoryRepo(conn)
	transactionSvc := usecase.NewTransactionService(txnHistoryRepo)
	a.TransactionHandler.Bind(ctx, transactionSvc)

	dashboardSvc := usecase.NewDashboardService(
		portfolioRepo, holdingRepo, stockRepo, paydayRepo, txnHistoryRepo, sectorRegistry,
	)
	a.DashboardHandler.Bind(ctx, dashboardSvc)

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
	a.CrashPlaybookHandler.Bind(ctx, crashPlaybookSvc, portfolioRepo)

	ghClient := github.NewClient()
	updateChecker := &releaseCheckerAdapter{client: ghClient}
	updateSvc := usecase.NewUpdateService(updateChecker, Version())

	downloader := updater.NewDownloader(ghClient)
	verifier := &updater.SHA256Verifier{}
	extractor := &updater.Extractor{}
	installer := updater.NewPlatformInstaller()
	selfUpdateSvc := usecase.NewSelfUpdateService(
		updateChecker, downloader, verifier, extractor,
		installer, wailsEmitter, Version(),
	)

	a.UpdateHandler.Bind(ctx, updateSvc, selfUpdateSvc, settingsRepo)

	a.BackupHandler.Bind(ctx, a.backup, a.dbPath, a.backupDir)
	a.BindBackup(a.backup, a.dbPath, a.backupDir)

	refreshSvc.Start(ctx)

	go a.CheckForUpdateOnStartup()
	go selfUpdateSvc.CleanupPreviousUpdate()
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
