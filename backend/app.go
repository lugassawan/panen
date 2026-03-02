package backend

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
	"github.com/lugassawan/panen/backend/infra/database"
	"github.com/lugassawan/panen/backend/infra/platform"
	"github.com/lugassawan/panen/backend/infra/scraper"
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
	yahoo := scraper.NewYahoo()

	stocks := usecase.NewStockService(stockRepo, yahoo)
	portfolios := usecase.NewPortfolioService(portfolioRepo, holdingRepo, buyTxnRepo, brokerageRepo, stockRepo)
	brokerages := usecase.NewBrokerageService(brokerageRepo)

	profileID, err := ensureDefaultUser(ctx, userRepo)
	if err != nil {
		runtime.LogFatalf(ctx, "ensure default user: %v", err)
	}

	a.StockHandler = presenter.NewStockHandler(ctx, stocks)
	a.PortfolioHandler = presenter.NewPortfolioHandler(ctx, portfolios)
	a.BrokerageHandler = presenter.NewBrokerageHandler(ctx, profileID, brokerages)
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

	now := time.Now().UTC()
	p := &user.Profile{
		ID:        shared.NewID(),
		Name:      platform.Username(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := users.Create(ctx, p); err != nil {
		return "", err
	}
	return p.ID, nil
}
