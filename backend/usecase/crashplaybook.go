package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/lugassawan/panen/backend/domain/crashplaybook"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

const (
	ihsgTicker           = "^JKSE"
	marketCacheTTL       = 1 * time.Hour
	settingDeployNormal  = "crash_deploy_pct_normal"
	settingDeployCrash   = "crash_deploy_pct_crash"
	settingDeployExtreme = "crash_deploy_pct_extreme"
)

// ErrInvalidDeploymentSum is returned when deployment percentages don't sum to 100.
var ErrInvalidDeploymentSum = errors.New("deployment percentages must sum to 100")

// PortfolioPlaybook is the full crash playbook for a portfolio.
type PortfolioPlaybook struct {
	Market     *crashplaybook.MarketStatus
	Stocks     []crashplaybook.StockPlaybook
	RefreshMin int
}

// DeploymentPlan shows how crash capital is allocated per level.
type DeploymentPlan struct {
	Total     float64
	Deployed  float64
	Remaining float64
	Levels    []DeploymentLevelPlan
}

// DeploymentLevelPlan shows allocation for a single crash level.
type DeploymentLevelPlan struct {
	Level  crashplaybook.CrashLevel
	Pct    float64
	Amount float64
}

// CrashPlaybookService manages crash playbook computations.
type CrashPlaybookService struct {
	stockData    stock.Repository
	provider     stock.DataProvider
	portfolios   portfolio.Repository
	holdings     portfolio.HoldingRepository
	crashCapital crashplaybook.CrashCapitalRepository
	settings     settings.Repository
	refresh      *RefreshService

	mu            sync.Mutex
	marketCache   *crashplaybook.MarketStatus
	prevCondition crashplaybook.MarketCondition
}

// NewCrashPlaybookService creates a new CrashPlaybookService.
func NewCrashPlaybookService(
	stockData stock.Repository,
	provider stock.DataProvider,
	portfolios portfolio.Repository,
	holdings portfolio.HoldingRepository,
	crashCapital crashplaybook.CrashCapitalRepository,
	settings settings.Repository,
	refresh *RefreshService,
) *CrashPlaybookService {
	return &CrashPlaybookService{
		stockData:    stockData,
		provider:     provider,
		portfolios:   portfolios,
		holdings:     holdings,
		crashCapital: crashCapital,
		settings:     settings,
		refresh:      refresh,
	}
}

// GetMarketStatus returns the current IHSG-based market condition.
func (s *CrashPlaybookService) GetMarketStatus(ctx context.Context) (*crashplaybook.MarketStatus, error) {
	if cached := s.cachedMarketStatus(); cached != nil {
		return cached, nil
	}

	data, err := s.stockData.GetByTickerAndSource(ctx, ihsgTicker, s.provider.Source())
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, fmt.Errorf("get IHSG data: %w", err)
	}

	if data == nil || time.Since(data.FetchedAt) >= marketCacheTTL {
		price, fetchErr := s.provider.FetchPrice(ctx, ihsgTicker)
		if fetchErr != nil {
			// Stale-data fallback: if we have old IHSG data, serve it rather than
			// failing. Market status is advisory, so stale data is preferable to
			// an error that blocks the entire crash playbook page.
			if data != nil {
				return s.buildMarketStatus(data), nil
			}
			return nil, fmt.Errorf("fetch IHSG: %w", fetchErr)
		}
		data = &stock.Data{
			ID:         shared.NewID(),
			Ticker:     ihsgTicker,
			Price:      price.Price,
			High52Week: price.High52Week,
			Low52Week:  price.Low52Week,
			FetchedAt:  time.Now().UTC(),
			Source:     s.provider.Source(),
		}
		if upsertErr := s.stockData.Upsert(ctx, data); upsertErr != nil {
			slog.Warn("failed to persist IHSG data", logKeyErr, upsertErr)
		}
	}

	return s.buildMarketStatus(data), nil
}

// GetPortfolioPlaybook returns crash playbooks for all holdings in a portfolio.
func (s *CrashPlaybookService) GetPortfolioPlaybook(
	ctx context.Context,
	portfolioID string,
) (*PortfolioPlaybook, error) {
	p, err := s.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get portfolio: %w", err)
	}

	holdingList, err := s.holdings.ListByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("list holdings: %w", err)
	}

	market, err := s.GetMarketStatus(ctx)
	if err != nil {
		return nil, err
	}

	deployPcts := s.readDeployPcts(ctx)

	var stocks []crashplaybook.StockPlaybook
	rp := valuation.RiskProfile(p.RiskProfile)
	for _, h := range holdingList {
		sp, err := s.buildStockPlaybook(ctx, h, rp, deployPcts)
		if err != nil {
			continue
		}
		stocks = append(stocks, *sp)
	}

	return &PortfolioPlaybook{
		Market:     market,
		Stocks:     stocks,
		RefreshMin: SuggestedRefreshInterval(market.Condition),
	}, nil
}

// GetDiagnostic evaluates the falling knife diagnostic for a stock.
func (s *CrashPlaybookService) GetDiagnostic(
	ctx context.Context,
	ticker string,
	portfolioID string,
	companyBadNews, fundamentalsOK *bool,
) (*crashplaybook.FallingKnifeDiagnostic, error) {
	market, err := s.GetMarketStatus(ctx)
	if err != nil {
		return nil, err
	}

	marketCrashed := market.Condition == crashplaybook.MarketCrash || market.Condition == crashplaybook.MarketCorrection

	p, err := s.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get portfolio: %w", err)
	}

	belowEntry := false
	data, err := s.stockData.GetByTickerAndSource(ctx, ticker, s.provider.Source())
	if err == nil {
		result, valErr := evaluate(data, valuation.RiskProfile(p.RiskProfile))
		if valErr == nil {
			belowEntry = data.Price <= result.EntryPrice
		}
	}

	signal := crashplaybook.EvaluateDiagnostic(marketCrashed, companyBadNews, fundamentalsOK, belowEntry)

	return &crashplaybook.FallingKnifeDiagnostic{
		MarketCrashed:  marketCrashed,
		CompanyBadNews: companyBadNews,
		FundamentalsOK: fundamentalsOK,
		BelowEntry:     belowEntry,
		Signal:         signal,
	}, nil
}

// GetCrashCapital returns the crash capital for a portfolio.
func (s *CrashPlaybookService) GetCrashCapital(
	ctx context.Context,
	portfolioID string,
) (*crashplaybook.CrashCapital, error) {
	cc, err := s.crashCapital.GetByPortfolioID(ctx, portfolioID)
	if errors.Is(err, shared.ErrNotFound) {
		return &crashplaybook.CrashCapital{PortfolioID: portfolioID}, nil
	}
	return cc, err
}

// SaveCrashCapital persists the crash capital amount for a portfolio.
func (s *CrashPlaybookService) SaveCrashCapital(ctx context.Context, portfolioID string, amount float64) error {
	now := time.Now().UTC()
	cc, err := s.crashCapital.GetByPortfolioID(ctx, portfolioID)
	switch {
	case errors.Is(err, shared.ErrNotFound):
		cc = &crashplaybook.CrashCapital{
			ID:          shared.NewID(),
			PortfolioID: portfolioID,
			Amount:      amount,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	case err != nil:
		return err
	default:
		cc.Amount = amount
		cc.UpdatedAt = now
	}
	return s.crashCapital.Upsert(ctx, cc)
}

// GetDeploymentPlan computes the capital allocation per crash level.
func (s *CrashPlaybookService) GetDeploymentPlan(ctx context.Context, portfolioID string) (*DeploymentPlan, error) {
	cc, err := s.GetCrashCapital(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	deployPcts := s.readDeployPcts(ctx)

	remaining := cc.Amount - cc.Deployed
	if remaining < 0 {
		remaining = 0
	}

	levels := []DeploymentLevelPlan{
		{Level: crashplaybook.LevelNormalDip, Pct: deployPcts[0], Amount: cc.Amount * deployPcts[0] / 100},
		{Level: crashplaybook.LevelCrash, Pct: deployPcts[1], Amount: cc.Amount * deployPcts[1] / 100},
		{Level: crashplaybook.LevelExtreme, Pct: deployPcts[2], Amount: cc.Amount * deployPcts[2] / 100},
	}

	return &DeploymentPlan{
		Total:     cc.Amount,
		Deployed:  cc.Deployed,
		Remaining: remaining,
		Levels:    levels,
	}, nil
}

// GetDeploymentSettings reads deployment percentage settings.
func (s *CrashPlaybookService) GetDeploymentSettings(ctx context.Context) (normal, crash, extreme float64, err error) {
	normal = s.readFloatSetting(ctx, settingDeployNormal, 30)
	crash = s.readFloatSetting(ctx, settingDeployCrash, 40)
	extreme = s.readFloatSetting(ctx, settingDeployExtreme, 30)
	return normal, crash, extreme, nil
}

// SaveDeploymentSettings persists deployment percentage settings. Sum must equal 100.
func (s *CrashPlaybookService) SaveDeploymentSettings(ctx context.Context, normal, crash, extreme float64) error {
	if math.Abs(normal+crash+extreme-100) > 0.01 {
		return ErrInvalidDeploymentSum
	}
	if err := s.settings.SetSetting(ctx, settingDeployNormal, strconv.FormatFloat(normal, 'f', -1, 64)); err != nil {
		return err
	}
	if err := s.settings.SetSetting(ctx, settingDeployCrash, strconv.FormatFloat(crash, 'f', -1, 64)); err != nil {
		return err
	}
	return s.settings.SetSetting(ctx, settingDeployExtreme, strconv.FormatFloat(extreme, 'f', -1, 64))
}

// SuggestedRefreshInterval returns the suggested refresh interval in minutes for a market condition.
func SuggestedRefreshInterval(condition crashplaybook.MarketCondition) int {
	switch condition {
	case crashplaybook.MarketCrash:
		return 180
	case crashplaybook.MarketCorrection:
		return 240
	case crashplaybook.MarketElevated:
		return 360
	case crashplaybook.MarketRecovery:
		return 360
	default:
		return 720
	}
}

func (s *CrashPlaybookService) cachedMarketStatus() *crashplaybook.MarketStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.marketCache != nil && time.Since(s.marketCache.FetchedAt) < marketCacheTTL {
		cached := *s.marketCache
		return &cached
	}
	return nil
}

func (s *CrashPlaybookService) readDeployPcts(ctx context.Context) [3]float64 {
	normal, crash, extreme, err := s.GetDeploymentSettings(ctx)
	if err != nil {
		return [3]float64{30, 40, 30}
	}
	return [3]float64{normal, crash, extreme}
}

func (s *CrashPlaybookService) readFloatSetting(ctx context.Context, key string, fallback float64) float64 {
	val, err := s.settings.GetSetting(ctx, key)
	if err != nil {
		return fallback
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fallback
	}
	return f
}

func (s *CrashPlaybookService) buildMarketStatus(data *stock.Data) *crashplaybook.MarketStatus {
	drawdown := crashplaybook.DrawdownPct(data.Price, data.High52Week)

	s.mu.Lock()
	defer s.mu.Unlock()

	condition := crashplaybook.DetectMarketCondition(data.Price, data.High52Week, s.prevCondition)

	if condition != s.prevCondition && s.refresh != nil {
		if condition == crashplaybook.MarketNormal {
			s.refresh.SetIntervalOverride(0)
		} else {
			s.refresh.SetIntervalOverride(SuggestedRefreshInterval(condition))
		}
	}

	s.prevCondition = condition

	status := &crashplaybook.MarketStatus{
		Condition:   condition,
		IHSGPrice:   data.Price,
		IHSGPeak:    data.High52Week,
		DrawdownPct: drawdown,
		FetchedAt:   data.FetchedAt,
	}
	s.marketCache = status
	return status
}

func (s *CrashPlaybookService) buildStockPlaybook(
	ctx context.Context,
	h *portfolio.Holding,
	rp valuation.RiskProfile,
	deployPcts [3]float64,
) (*crashplaybook.StockPlaybook, error) {
	data, err := s.stockData.GetByTickerAndSource(ctx, h.Ticker, s.provider.Source())
	if err != nil {
		return nil, err
	}

	result, err := evaluate(data, rp)
	if err != nil {
		return nil, err
	}

	levels := crashplaybook.ComputeResponseLevels(result.EntryPrice, data.Low52Week, deployPcts)
	active := crashplaybook.DetermineActiveLevel(data.Price, levels)

	return &crashplaybook.StockPlaybook{
		Ticker:       h.Ticker,
		CurrentPrice: data.Price,
		EntryPrice:   result.EntryPrice,
		Levels:       levels,
		ActiveLevel:  active,
	}, nil
}
