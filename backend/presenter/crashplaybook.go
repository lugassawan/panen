package presenter

import (
	"context"
	"fmt"

	"github.com/lugassawan/panen/backend/domain/crashplaybook"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/usecase"
)

// CrashPlaybookHandler handles crash playbook requests from the frontend.
type CrashPlaybookHandler struct {
	ctx        context.Context
	service    *usecase.CrashPlaybookService
	portfolios portfolio.Repository
}

// NewCrashPlaybookHandler creates a new CrashPlaybookHandler.
func NewCrashPlaybookHandler(
	ctx context.Context,
	service *usecase.CrashPlaybookService,
	portfolios portfolio.Repository,
) *CrashPlaybookHandler {
	h := &CrashPlaybookHandler{}
	h.Bind(ctx, service, portfolios)
	return h
}

func (h *CrashPlaybookHandler) Bind(
	ctx context.Context,
	service *usecase.CrashPlaybookService,
	portfolios portfolio.Repository,
) {
	h.ctx = ctx
	h.service = service
	h.portfolios = portfolios
}

// ListAllPortfolios returns all portfolios for the portfolio selector.
func (h *CrashPlaybookHandler) ListAllPortfolios() ([]*PortfolioResponse, error) {
	all, err := h.portfolios.ListAll(h.ctx)
	if err != nil {
		return nil, fmt.Errorf("list all portfolios: %w", err)
	}
	result := make([]*PortfolioResponse, len(all))
	for i, p := range all {
		result[i] = newPortfolioResponse(p)
	}
	return result, nil
}

// GetMarketStatus returns the current IHSG market condition.
func (h *CrashPlaybookHandler) GetMarketStatus() (*MarketStatusResponse, error) {
	status, err := h.service.GetMarketStatus(h.ctx)
	if err != nil {
		return nil, fmt.Errorf("get market status: %w", err)
	}
	return newMarketStatusResponse(status), nil
}

// GetPortfolioPlaybook returns crash playbooks for all holdings in a portfolio.
func (h *CrashPlaybookHandler) GetPortfolioPlaybook(portfolioID string) (*PortfolioPlaybookResponse, error) {
	pb, err := h.service.GetPortfolioPlaybook(h.ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get portfolio playbook: %w", err)
	}
	return newPortfolioPlaybookResponse(pb), nil
}

// GetDiagnostic evaluates the falling knife diagnostic for a stock.
func (h *CrashPlaybookHandler) GetDiagnostic(
	ticker, portfolioID string,
	companyBadNews, fundamentalsOK *bool,
) (*DiagnosticResponse, error) {
	diag, err := h.service.GetDiagnostic(h.ctx, ticker, portfolioID, companyBadNews, fundamentalsOK)
	if err != nil {
		return nil, fmt.Errorf("get diagnostic: %w", err)
	}
	return newDiagnosticResponse(diag), nil
}

// GetCrashCapital returns the crash capital for a portfolio.
func (h *CrashPlaybookHandler) GetCrashCapital(portfolioID string) (*CrashCapitalResponse, error) {
	cc, err := h.service.GetCrashCapital(h.ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get crash capital: %w", err)
	}
	return &CrashCapitalResponse{
		PortfolioID: cc.PortfolioID,
		Amount:      cc.Amount,
		Deployed:    cc.Deployed,
	}, nil
}

// SaveCrashCapital persists the crash capital amount for a portfolio.
func (h *CrashPlaybookHandler) SaveCrashCapital(portfolioID string, amount float64) error {
	if err := h.service.SaveCrashCapital(h.ctx, portfolioID, amount); err != nil {
		return fmt.Errorf("save crash capital: %w", err)
	}
	return nil
}

// GetDeploymentPlan returns the capital allocation breakdown per crash level.
func (h *CrashPlaybookHandler) GetDeploymentPlan(portfolioID string) (*DeploymentPlanResponse, error) {
	plan, err := h.service.GetDeploymentPlan(h.ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get deployment plan: %w", err)
	}
	return newDeploymentPlanResponse(plan), nil
}

// GetDeploymentSettings returns the deployment percentage settings.
func (h *CrashPlaybookHandler) GetDeploymentSettings() (*DeploymentSettingsResponse, error) {
	normal, crash, extreme, err := h.service.GetDeploymentSettings(h.ctx)
	if err != nil {
		return nil, fmt.Errorf("get deployment settings: %w", err)
	}
	return &DeploymentSettingsResponse{
		Normal:  normal,
		Crash:   crash,
		Extreme: extreme,
	}, nil
}

// SaveDeploymentSettings persists deployment percentage settings.
func (h *CrashPlaybookHandler) SaveDeploymentSettings(normal, crash, extreme float64) error {
	if err := h.service.SaveDeploymentSettings(h.ctx, normal, crash, extreme); err != nil {
		return fmt.Errorf("save deployment settings: %w", err)
	}
	return nil
}

func newMarketStatusResponse(s *crashplaybook.MarketStatus) *MarketStatusResponse {
	return &MarketStatusResponse{
		Condition:   string(s.Condition),
		IHSGPrice:   s.IHSGPrice,
		IHSGPeak:    s.IHSGPeak,
		DrawdownPct: s.DrawdownPct,
		FetchedAt:   formatDTO(s.FetchedAt),
	}
}

func newPortfolioPlaybookResponse(pb *usecase.PortfolioPlaybook) *PortfolioPlaybookResponse {
	stocks := make([]StockPlaybookResponse, len(pb.Stocks))
	for i, sp := range pb.Stocks {
		stocks[i] = newStockPlaybookResponse(sp)
	}
	return &PortfolioPlaybookResponse{
		Market:     *newMarketStatusResponse(pb.Market),
		Stocks:     stocks,
		RefreshMin: pb.RefreshMin,
	}
}

func newStockPlaybookResponse(sp crashplaybook.StockPlaybook) StockPlaybookResponse {
	levels := make([]ResponseLevelResponse, len(sp.Levels))
	for i, l := range sp.Levels {
		levels[i] = ResponseLevelResponse{
			Level:        string(l.Level),
			TriggerPrice: l.TriggerPrice,
			DeployPct:    l.DeployPct,
		}
	}
	resp := StockPlaybookResponse{
		Ticker:       sp.Ticker,
		CurrentPrice: sp.CurrentPrice,
		EntryPrice:   sp.EntryPrice,
		Levels:       levels,
	}
	if sp.ActiveLevel != nil {
		s := string(*sp.ActiveLevel)
		resp.ActiveLevel = &s
	}
	return resp
}

func newDiagnosticResponse(d *crashplaybook.FallingKnifeDiagnostic) *DiagnosticResponse {
	return &DiagnosticResponse{
		MarketCrashed:  d.MarketCrashed,
		CompanyBadNews: d.CompanyBadNews,
		FundamentalsOK: d.FundamentalsOK,
		BelowEntry:     d.BelowEntry,
		Signal:         string(d.Signal),
	}
}

func newDeploymentPlanResponse(plan *usecase.DeploymentPlan) *DeploymentPlanResponse {
	levels := make([]DeploymentLevelPlanResponse, len(plan.Levels))
	for i, l := range plan.Levels {
		levels[i] = DeploymentLevelPlanResponse{
			Level:  string(l.Level),
			Pct:    l.Pct,
			Amount: l.Amount,
		}
	}
	return &DeploymentPlanResponse{
		Total:     plan.Total,
		Deployed:  plan.Deployed,
		Remaining: plan.Remaining,
		Levels:    levels,
	}
}
