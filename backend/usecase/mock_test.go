package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
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
