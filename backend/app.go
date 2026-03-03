package backend

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lugassawan/panen/backend/domain/user"
	brokerConfigLoader "github.com/lugassawan/panen/backend/infra/brokerconfig"
	"github.com/lugassawan/panen/backend/infra/database"
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
	db *database.DB
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
	portfolios := usecase.NewPortfolioService(portfolioRepo, holdingRepo, buyTxnRepo, brokerageRepo, stockRepo)
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

	a.StockHandler = presenter.NewStockHandler(ctx, stocks)
	a.PortfolioHandler = presenter.NewPortfolioHandler(ctx, portfolios)
	a.BrokerageHandler = presenter.NewBrokerageHandler(ctx, profileID, brokerages)
	a.BrokerConfigHandler = presenter.NewBrokerConfigHandler(brokerConfigs)
	a.WatchlistHandler = presenter.NewWatchlistHandler(ctx, profileID, watchlistSvc)
}

// Shutdown closes the database connection.
func (a *App) Shutdown(ctx context.Context) {
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			runtime.LogErrorf(ctx, "close database: %v", err)
		}
	}
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
