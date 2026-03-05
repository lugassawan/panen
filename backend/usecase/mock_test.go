package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/trailingstop"
)

// mockStockRepo is an in-memory stock.Repository for testing.
type mockStockRepo struct {
	mu    sync.Mutex
	items map[string]*stock.Data // keyed by "ticker:source"
}

func newMockStockRepo() *mockStockRepo {
	return &mockStockRepo{items: make(map[string]*stock.Data)}
}

func (r *mockStockRepo) Upsert(_ context.Context, d *stock.Data) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[d.Ticker+":"+d.Source] = d
	return nil
}

func (r *mockStockRepo) GetByTicker(_ context.Context, ticker string) (*stock.Data, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, d := range r.items {
		if d.Ticker == ticker {
			return d, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (r *mockStockRepo) GetByTickerAndSource(_ context.Context, ticker, source string) (*stock.Data, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.items[ticker+":"+source]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return d, nil
}

func (r *mockStockRepo) DeleteOlderThan(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

func (r *mockStockRepo) ListAllTickers(_ context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := make(map[string]bool)
	var tickers []string
	for _, d := range r.items {
		if !seen[d.Ticker] {
			seen[d.Ticker] = true
			tickers = append(tickers, d.Ticker)
		}
	}
	return tickers, nil
}

// mockProvider is an in-memory stock.DataProvider for testing.
type mockProvider struct {
	source    string
	priceFunc func(ctx context.Context, ticker string) (*stock.PriceResult, error)
	finFunc   func(ctx context.Context, ticker string) (*stock.FinancialResult, error)
	callCount int
	mu        sync.Mutex
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		source: "mock",
		priceFunc: func(_ context.Context, _ string) (*stock.PriceResult, error) {
			return &stock.PriceResult{Price: 8500, High52Week: 9000, Low52Week: 7000}, nil
		},
		finFunc: func(_ context.Context, _ string) (*stock.FinancialResult, error) {
			return &stock.FinancialResult{
				EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
				PBV: 2.8, PER: 17, DividendYield: 2.5, PayoutRatio: 40,
			}, nil
		},
	}
}

func (p *mockProvider) Source() string { return p.source }

func (p *mockProvider) FetchPrice(ctx context.Context, ticker string) (*stock.PriceResult, error) {
	p.mu.Lock()
	p.callCount++
	p.mu.Unlock()
	return p.priceFunc(ctx, ticker)
}

func (p *mockProvider) FetchFinancials(ctx context.Context, ticker string) (*stock.FinancialResult, error) {
	return p.finFunc(ctx, ticker)
}

func (p *mockProvider) FetchPriceHistory(_ context.Context, _ string) ([]stock.PricePoint, error) {
	return nil, nil
}

// mockBrokerageRepo is an in-memory brokerage.Repository for testing.
type mockBrokerageRepo struct {
	mu    sync.Mutex
	items map[string]*brokerage.Account
}

func newMockBrokerageRepo() *mockBrokerageRepo {
	return &mockBrokerageRepo{items: make(map[string]*brokerage.Account)}
}

func (r *mockBrokerageRepo) Create(_ context.Context, a *brokerage.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[a.ID] = a
	return nil
}

func (r *mockBrokerageRepo) GetByID(_ context.Context, id string) (*brokerage.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.items[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return a, nil
}

func (r *mockBrokerageRepo) ListByProfileID(_ context.Context, profileID string) ([]*brokerage.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*brokerage.Account
	for _, a := range r.items {
		if a.ProfileID == profileID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (r *mockBrokerageRepo) Update(_ context.Context, a *brokerage.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[a.ID]; !ok {
		return shared.ErrNotFound
	}
	r.items[a.ID] = a
	return nil
}

func (r *mockBrokerageRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return shared.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

// mockPortfolioRepo is an in-memory portfolio.Repository for testing.
type mockPortfolioRepo struct {
	mu    sync.Mutex
	items map[string]*portfolio.Portfolio
}

func newMockPortfolioRepo() *mockPortfolioRepo {
	return &mockPortfolioRepo{items: make(map[string]*portfolio.Portfolio)}
}

func (r *mockPortfolioRepo) Create(_ context.Context, p *portfolio.Portfolio) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[p.ID] = p
	return nil
}

func (r *mockPortfolioRepo) GetByID(_ context.Context, id string) (*portfolio.Portfolio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.items[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return p, nil
}

func (r *mockPortfolioRepo) ListAll(_ context.Context) ([]*portfolio.Portfolio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*portfolio.Portfolio, 0, len(r.items))
	for _, p := range r.items {
		result = append(result, p)
	}
	return result, nil
}

func (r *mockPortfolioRepo) ListByBrokerageAccountID(
	_ context.Context,
	brokerageAccountID string,
) ([]*portfolio.Portfolio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*portfolio.Portfolio
	for _, p := range r.items {
		if p.BrokerageAccountID == brokerageAccountID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *mockPortfolioRepo) Update(_ context.Context, p *portfolio.Portfolio) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[p.ID]; !ok {
		return shared.ErrNotFound
	}
	r.items[p.ID] = p
	return nil
}

func (r *mockPortfolioRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return shared.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

// mockHoldingRepo is an in-memory portfolio.HoldingRepository for testing.
type mockHoldingRepo struct {
	mu    sync.Mutex
	items map[string]*portfolio.Holding
}

func newMockHoldingRepo() *mockHoldingRepo {
	return &mockHoldingRepo{items: make(map[string]*portfolio.Holding)}
}

func (r *mockHoldingRepo) Create(_ context.Context, h *portfolio.Holding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[h.ID] = h
	return nil
}

func (r *mockHoldingRepo) GetByID(_ context.Context, id string) (*portfolio.Holding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.items[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return h, nil
}

func (r *mockHoldingRepo) GetByPortfolioAndTicker(
	_ context.Context,
	portfolioID, ticker string,
) (*portfolio.Holding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, h := range r.items {
		if h.PortfolioID == portfolioID && h.Ticker == ticker {
			return h, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (r *mockHoldingRepo) ListByPortfolioID(_ context.Context, portfolioID string) ([]*portfolio.Holding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*portfolio.Holding
	for _, h := range r.items {
		if h.PortfolioID == portfolioID {
			result = append(result, h)
		}
	}
	return result, nil
}

func (r *mockHoldingRepo) Update(_ context.Context, h *portfolio.Holding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[h.ID]; !ok {
		return shared.ErrNotFound
	}
	r.items[h.ID] = h
	return nil
}

func (r *mockHoldingRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return shared.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

// mockBuyTxnRepo is an in-memory portfolio.BuyTransactionRepository for testing.
type mockBuyTxnRepo struct {
	mu    sync.Mutex
	items map[string]*portfolio.BuyTransaction
}

func newMockBuyTxnRepo() *mockBuyTxnRepo {
	return &mockBuyTxnRepo{items: make(map[string]*portfolio.BuyTransaction)}
}

func (r *mockBuyTxnRepo) Create(_ context.Context, tx *portfolio.BuyTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[tx.ID] = tx
	return nil
}

func (r *mockBuyTxnRepo) GetByID(_ context.Context, id string) (*portfolio.BuyTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.items[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return tx, nil
}

func (r *mockBuyTxnRepo) ListByHoldingID(_ context.Context, holdingID string) ([]*portfolio.BuyTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*portfolio.BuyTransaction
	for _, tx := range r.items {
		if tx.HoldingID == holdingID {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (r *mockBuyTxnRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return shared.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

// mockSettingsRepo implements settings.Repository for testing.
type mockSettingsRepo struct {
	mu       sync.Mutex
	settings *settings.RefreshSettings
	kv       map[string]string
}

func newMockSettingsRepo() *mockSettingsRepo {
	return &mockSettingsRepo{
		settings: settings.DefaultRefreshSettings(),
		kv:       make(map[string]string),
	}
}

func (r *mockSettingsRepo) GetRefreshSettings(_ context.Context) (*settings.RefreshSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Return a copy to avoid data races.
	s := *r.settings
	return &s, nil
}

func (r *mockSettingsRepo) SaveRefreshSettings(_ context.Context, s *settings.RefreshSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings = s
	return nil
}

func (r *mockSettingsRepo) GetSetting(_ context.Context, key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.kv[key], nil
}

func (r *mockSettingsRepo) SetSetting(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.kv[key] = value
	return nil
}

// mockTickerCollector implements TickerCollector for testing.
type mockTickerCollector struct {
	mu      sync.Mutex
	tickers []string
	err     error
}

func newMockTickerCollector(tickers ...string) *mockTickerCollector {
	return &mockTickerCollector{tickers: tickers}
}

func (c *mockTickerCollector) CollectAll(_ context.Context) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return nil, c.err
	}
	result := make([]string, len(c.tickers))
	copy(result, c.tickers)
	return result, nil
}

// mockEventEmitter implements EventEmitter for testing.
type mockEventEmitter struct {
	mu     sync.Mutex
	events []emittedEvent
}

type emittedEvent struct {
	name string
	data any
}

func newMockEventEmitter() *mockEventEmitter {
	return &mockEventEmitter{}
}

func (e *mockEventEmitter) Emit(eventName string, data any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = append(e.events, emittedEvent{name: eventName, data: data})
}

func (e *mockEventEmitter) eventsByName(name string) []emittedEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	var result []emittedEvent
	for _, ev := range e.events {
		if ev.name == name {
			result = append(result, ev)
		}
	}
	return result
}

// mockChecklistResultRepo is an in-memory checklist.Repository for testing.
type mockChecklistResultRepo struct {
	mu    sync.Mutex
	items map[string]*checklist.ChecklistResult // keyed by "portfolioID:ticker:action"
	byID  map[string]*checklist.ChecklistResult // keyed by ID
}

func newMockChecklistResultRepo() *mockChecklistResultRepo {
	return &mockChecklistResultRepo{
		items: make(map[string]*checklist.ChecklistResult),
		byID:  make(map[string]*checklist.ChecklistResult),
	}
}

func checklistKey(portfolioID, ticker string, action checklist.ActionType) string {
	return portfolioID + ":" + ticker + ":" + string(action)
}

func (r *mockChecklistResultRepo) Upsert(_ context.Context, cr *checklist.ChecklistResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := checklistKey(cr.PortfolioID, cr.Ticker, cr.Action)
	r.items[key] = cr
	r.byID[cr.ID] = cr
	return nil
}

func (r *mockChecklistResultRepo) Get(
	_ context.Context,
	portfolioID, ticker string,
	action checklist.ActionType,
) (*checklist.ChecklistResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := checklistKey(portfolioID, ticker, action)
	cr, ok := r.items[key]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return cr, nil
}

func (r *mockChecklistResultRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cr, ok := r.byID[id]
	if !ok {
		return shared.ErrNotFound
	}
	key := checklistKey(cr.PortfolioID, cr.Ticker, cr.Action)
	delete(r.items, key)
	delete(r.byID, id)
	return nil
}

func (r *mockChecklistResultRepo) DeleteByPortfolioID(_ context.Context, portfolioID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, cr := range r.items {
		if cr.PortfolioID == portfolioID {
			delete(r.items, key)
			delete(r.byID, cr.ID)
		}
	}
	return nil
}

// mockPaydayRepo is an in-memory payday.Repository for testing.
type mockPaydayRepo struct {
	mu     sync.Mutex
	events map[string]*payday.PaydayEvent // key: month+":"+portfolioID
}

func newMockPaydayRepo() *mockPaydayRepo {
	return &mockPaydayRepo{events: make(map[string]*payday.PaydayEvent)}
}

func paydayKey(month, portfolioID string) string {
	return month + ":" + portfolioID
}

func (r *mockPaydayRepo) Create(_ context.Context, event *payday.PaydayEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[paydayKey(event.Month, event.PortfolioID)] = event
	return nil
}

func (r *mockPaydayRepo) GetByMonthAndPortfolio(
	_ context.Context,
	month, portfolioID string,
) (*payday.PaydayEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ev, ok := r.events[paydayKey(month, portfolioID)]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return ev, nil
}

func (r *mockPaydayRepo) ListByMonth(_ context.Context, month string) ([]*payday.PaydayEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*payday.PaydayEvent
	for _, ev := range r.events {
		if ev.Month == month {
			result = append(result, ev)
		}
	}
	return result, nil
}

func (r *mockPaydayRepo) ListByPortfolioID(_ context.Context, portfolioID string) ([]*payday.PaydayEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*payday.PaydayEvent
	for _, ev := range r.events {
		if ev.PortfolioID == portfolioID {
			result = append(result, ev)
		}
	}
	return result, nil
}

func (r *mockPaydayRepo) Update(_ context.Context, event *payday.PaydayEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := paydayKey(event.Month, event.PortfolioID)
	if _, ok := r.events[key]; !ok {
		return shared.ErrNotFound
	}
	r.events[key] = event
	return nil
}

// mockCashFlowRepo is an in-memory payday.CashFlowRepository for testing.
type mockCashFlowRepo struct {
	mu    sync.Mutex
	flows map[string][]*payday.CashFlow // key: portfolioID
}

func newMockCashFlowRepo() *mockCashFlowRepo {
	return &mockCashFlowRepo{flows: make(map[string][]*payday.CashFlow)}
}

func (r *mockCashFlowRepo) Create(_ context.Context, cf *payday.CashFlow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.flows[cf.PortfolioID] = append(r.flows[cf.PortfolioID], cf)
	return nil
}

func (r *mockCashFlowRepo) ListByPortfolioID(_ context.Context, portfolioID string) ([]*payday.CashFlow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.flows[portfolioID], nil
}

func (r *mockCashFlowRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for pid, flows := range r.flows {
		for i, cf := range flows {
			if cf.ID == id {
				r.flows[pid] = append(flows[:i], flows[i+1:]...)
				return nil
			}
		}
	}
	return shared.ErrNotFound
}

// mockPeakRepo is an in-memory trailingstop.PeakRepository for testing.
type mockPeakRepo struct {
	mu    sync.Mutex
	items map[string]*trailingstop.HoldingPeak // keyed by holdingID
}

func newMockPeakRepo() *mockPeakRepo {
	return &mockPeakRepo{items: make(map[string]*trailingstop.HoldingPeak)}
}

func (r *mockPeakRepo) Upsert(_ context.Context, peak *trailingstop.HoldingPeak) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[peak.HoldingID] = peak
	return nil
}

func (r *mockPeakRepo) GetByHoldingID(_ context.Context, holdingID string) (*trailingstop.HoldingPeak, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	hp, ok := r.items[holdingID]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return hp, nil
}

func (r *mockPeakRepo) ListByHoldingIDs(_ context.Context, holdingIDs []string) ([]*trailingstop.HoldingPeak, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*trailingstop.HoldingPeak
	for _, id := range holdingIDs {
		if hp, ok := r.items[id]; ok {
			result = append(result, hp)
		}
	}
	return result, nil
}
