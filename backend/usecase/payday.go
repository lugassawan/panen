package usecase

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/domain/shared"
)

const (
	settingPaydayDay = "payday_day"
	monthLayout      = "2006-01"
)

var (
	ErrPaydayNotConfigured = errors.New("payday day not configured")
	ErrInvalidPaydayDay    = errors.New("payday day must be between 0 and 31")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrDeferDateNotFuture  = errors.New("defer date must be in the future")
)

// MonthlyPaydayStatus aggregates payday events for a single month.
type MonthlyPaydayStatus struct {
	Month         string
	PaydayDay     int
	Portfolios    []PortfolioPaydayStatus
	TotalExpected float64
}

// PortfolioPaydayStatus holds the payday state for one portfolio in a given month.
type PortfolioPaydayStatus struct {
	PortfolioID   string
	PortfolioName string
	Mode          string
	Expected      float64
	Actual        float64
	Status        string
	DeferUntil    *time.Time
}

// CashFlowSummary aggregates cash flow data for a portfolio.
type CashFlowSummary struct {
	Items         []CashFlowItem
	TotalInflow   float64
	TotalDeployed float64
	Balance       float64
}

// CashFlowItem represents a single cash flow entry in the summary.
type CashFlowItem struct {
	ID          string
	PortfolioID string
	Type        string
	Amount      float64
	Date        time.Time
	Note        string
	CreatedAt   time.Time
}

// PaydayService handles payday scheduling, confirmation, and cash flow tracking.
type PaydayService struct {
	events     payday.Repository
	cashFlows  payday.CashFlowRepository
	portfolios portfolio.Repository
	settings   settings.Repository
}

// NewPaydayService creates a new PaydayService.
func NewPaydayService(
	events payday.Repository,
	cashFlows payday.CashFlowRepository,
	portfolios portfolio.Repository,
	settings settings.Repository,
) *PaydayService {
	return &PaydayService{
		events:     events,
		cashFlows:  cashFlows,
		portfolios: portfolios,
		settings:   settings,
	}
}

// GetPaydayDay reads the configured payday day from settings.
// Returns 0 if not set or empty.
func (s *PaydayService) GetPaydayDay(ctx context.Context) (int, error) {
	val, err := s.settings.GetSetting(ctx, settingPaydayDay)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if val == "" {
		return 0, nil
	}
	day, parseErr := strconv.Atoi(val)
	if parseErr != nil {
		return 0, nil //nolint:nilerr // treat corrupt/non-numeric setting as "not configured"
	}
	return day, nil
}

// SavePaydayDay validates and persists the payday day setting.
// Pass 0 to disable payday.
func (s *PaydayService) SavePaydayDay(ctx context.Context, day int) error {
	if day < 0 || day > 31 {
		return ErrInvalidPaydayDay
	}
	return s.settings.SetSetting(ctx, settingPaydayDay, strconv.Itoa(day))
}

// GetCurrentMonthStatus builds the payday status for the current month,
// lazy-creating events for portfolios with a monthly addition configured.
func (s *PaydayService) GetCurrentMonthStatus(ctx context.Context) (*MonthlyPaydayStatus, error) {
	paydayDay, err := s.GetPaydayDay(ctx)
	if err != nil {
		return nil, err
	}
	if paydayDay == 0 {
		return nil, ErrPaydayNotConfigured
	}

	now := time.Now().UTC()
	currentMonth := now.Format(monthLayout)

	allPortfolios, err := s.portfolios.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var statuses []PortfolioPaydayStatus
	var totalExpected float64

	for _, p := range allPortfolios {
		if p.MonthlyAddition <= 0 {
			continue
		}

		event, err := s.ensureEvent(ctx, currentMonth, p, paydayDay, now)
		if err != nil {
			return nil, err
		}

		if err := s.autoTransitionDeferred(ctx, event, now); err != nil {
			return nil, err
		}

		totalExpected += event.Expected
		statuses = append(statuses, newPortfolioPaydayStatus(p, event))
	}

	return &MonthlyPaydayStatus{
		Month:         currentMonth,
		PaydayDay:     paydayDay,
		Portfolios:    statuses,
		TotalExpected: totalExpected,
	}, nil
}

func newPortfolioPaydayStatus(p *portfolio.Portfolio, event *payday.PaydayEvent) PortfolioPaydayStatus {
	return PortfolioPaydayStatus{
		PortfolioID:   p.ID,
		PortfolioName: p.Name,
		Mode:          string(p.Mode),
		Expected:      event.Expected,
		Actual:        event.Actual,
		Status:        string(event.Status),
		DeferUntil:    event.DeferUntil,
	}
}

// ConfirmPayday marks the current month's payday event as confirmed for a portfolio
// and records a cash flow entry.
func (s *PaydayService) ConfirmPayday(ctx context.Context, portfolioID string, actualAmount float64) error {
	now := time.Now().UTC()
	currentMonth := now.Format(monthLayout)

	event, err := s.events.GetByMonthAndPortfolio(ctx, currentMonth, portfolioID)
	if err != nil {
		return err
	}

	if !payday.ValidTransition(event.Status, payday.StatusConfirmed) {
		return ErrInvalidTransition
	}

	event.Status = payday.StatusConfirmed
	event.Actual = actualAmount
	event.ConfirmedAt = &now
	event.UpdatedAt = now
	if err := s.events.Update(ctx, event); err != nil {
		return err
	}

	cf := payday.NewCashFlow(portfolioID, payday.FlowTypeMonthly, actualAmount, now, "Monthly payday")
	return s.cashFlows.Create(ctx, cf)
}

// DeferPayday defers the current month's payday event to a later date.
// The deferUntil date must be in the future (after today).
func (s *PaydayService) DeferPayday(ctx context.Context, portfolioID string, deferUntil time.Time) error {
	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)
	deferDate := deferUntil.Truncate(24 * time.Hour)
	if !deferDate.After(today) {
		return ErrDeferDateNotFuture
	}

	currentMonth := now.Format(monthLayout)

	event, err := s.events.GetByMonthAndPortfolio(ctx, currentMonth, portfolioID)
	if err != nil {
		return err
	}

	if !payday.ValidTransition(event.Status, payday.StatusDeferred) {
		return ErrInvalidTransition
	}

	event.Status = payday.StatusDeferred
	event.DeferUntil = &deferUntil
	event.UpdatedAt = now
	return s.events.Update(ctx, event)
}

// SkipPayday marks the current month's payday event as skipped for a portfolio.
func (s *PaydayService) SkipPayday(ctx context.Context, portfolioID string) error {
	now := time.Now().UTC()
	currentMonth := now.Format(monthLayout)

	event, err := s.events.GetByMonthAndPortfolio(ctx, currentMonth, portfolioID)
	if err != nil {
		return err
	}

	if !payday.ValidTransition(event.Status, payday.StatusSkipped) {
		return ErrInvalidTransition
	}

	event.Status = payday.StatusSkipped
	event.UpdatedAt = now
	return s.events.Update(ctx, event)
}

// GetPaydayHistory returns payday statuses for all past months (excluding current),
// sorted by month descending.
func (s *PaydayService) GetPaydayHistory(ctx context.Context) ([]*MonthlyPaydayStatus, error) {
	paydayDay, err := s.GetPaydayDay(ctx)
	if err != nil {
		return nil, err
	}

	allPortfolios, err := s.portfolios.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	currentMonth := time.Now().UTC().Format(monthLayout)

	// Build a lookup of portfolios by ID.
	portfolioMap := shared.IndexBy(allPortfolios, func(p *portfolio.Portfolio) string { return p.ID })

	// Collect all events grouped by month.
	monthEvents := make(map[string][]*payday.PaydayEvent)
	for _, p := range allPortfolios {
		if p.MonthlyAddition <= 0 {
			continue
		}
		events, err := s.events.ListByPortfolioID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		for _, ev := range events {
			if ev.Month == currentMonth {
				continue
			}
			monthEvents[ev.Month] = append(monthEvents[ev.Month], ev)
		}
	}

	// Build MonthlyPaydayStatus for each past month.
	var result []*MonthlyPaydayStatus
	for month, events := range monthEvents {
		result = append(result, buildMonthStatus(month, paydayDay, events, portfolioMap))
	}

	// Sort by month descending.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Month > result[j].Month
	})

	return result, nil
}

// GetCashFlowSummary returns a summary of cash flows for a portfolio.
func (s *PaydayService) GetCashFlowSummary(ctx context.Context, portfolioID string) (*CashFlowSummary, error) {
	flows, err := s.cashFlows.ListByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	items := make([]CashFlowItem, len(flows))
	var totalInflow float64
	for i, cf := range flows {
		items[i] = CashFlowItem{
			ID:          cf.ID,
			PortfolioID: cf.PortfolioID,
			Type:        string(cf.Type),
			Amount:      cf.Amount,
			Date:        cf.Date,
			Note:        cf.Note,
			CreatedAt:   cf.CreatedAt,
		}
		totalInflow += cf.Amount
	}

	totalDeployed := 0.0 // TODO: calculate from buy transactions when portfolio deployment tracking is added
	return &CashFlowSummary{
		Items:         items,
		TotalInflow:   totalInflow,
		TotalDeployed: totalDeployed,
		Balance:       totalInflow - totalDeployed,
	}, nil
}

// ensureEvent retrieves or lazy-creates a payday event for the given month and portfolio.
func (s *PaydayService) ensureEvent(
	ctx context.Context,
	month string,
	p *portfolio.Portfolio,
	paydayDay int,
	now time.Time,
) (*payday.PaydayEvent, error) {
	event, err := s.events.GetByMonthAndPortfolio(ctx, month, p.ID)
	if err == nil {
		return event, nil
	}
	if !errors.Is(err, shared.ErrNotFound) {
		return nil, err
	}
	initialStatus := payday.StatusScheduled
	if now.Day() >= paydayDay {
		initialStatus = payday.StatusPending
	}
	event = payday.NewPaydayEvent(month, p.ID, p.MonthlyAddition)
	event.Status = initialStatus
	if err := s.events.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

// autoTransitionDeferred moves a DEFERRED event back to PENDING if DeferUntil has passed.
func (s *PaydayService) autoTransitionDeferred(
	ctx context.Context,
	event *payday.PaydayEvent,
	now time.Time,
) error {
	if event.Status != payday.StatusDeferred || event.DeferUntil == nil {
		return nil
	}
	deferDate := event.DeferUntil.Truncate(24 * time.Hour)
	today := now.Truncate(24 * time.Hour)
	if deferDate.After(today) {
		return nil
	}
	event.Status = payday.StatusPending
	event.UpdatedAt = now
	return s.events.Update(ctx, event)
}

func buildMonthStatus(
	month string,
	paydayDay int,
	events []*payday.PaydayEvent,
	portfolioMap map[string]*portfolio.Portfolio,
) *MonthlyPaydayStatus {
	statuses := make([]PortfolioPaydayStatus, 0, len(events))
	var totalExpected float64
	for _, ev := range events {
		p := portfolioMap[ev.PortfolioID]
		name := ""
		mode := ""
		if p != nil {
			name = p.Name
			mode = string(p.Mode)
		}
		totalExpected += ev.Expected
		statuses = append(statuses, PortfolioPaydayStatus{
			PortfolioID:   ev.PortfolioID,
			PortfolioName: name,
			Mode:          mode,
			Expected:      ev.Expected,
			Actual:        ev.Actual,
			Status:        string(ev.Status),
			DeferUntil:    ev.DeferUntil,
		})
	}
	return &MonthlyPaydayStatus{
		Month:         month,
		PaydayDay:     paydayDay,
		Portfolios:    statuses,
		TotalExpected: totalExpected,
	}
}
