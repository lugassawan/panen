package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
	"github.com/lugassawan/panen/backend/domain/watchlist"
)

// IndexRegistry provides lookup access to index compositions.
type IndexRegistry interface {
	Tickers(name string) ([]string, bool)
	Names() []string
}

// SectorRegistry maps tickers to their sector.
type SectorRegistry interface {
	SectorOf(ticker string) string
	AllSectors() []string
}

// WatchlistItemWithData is a use-case-layer composite carrying a watchlist item
// together with its optional sector, stock data, and valuation result.
type WatchlistItemWithData struct {
	Ticker    string
	Sector    string
	StockData *stock.Data
	Valuation *valuation.ValuationResult
}

// WatchlistService handles watchlist and watchlist item operations.
type WatchlistService struct {
	watchlists     watchlist.Repository
	items          watchlist.ItemRepository
	stockData      stock.Repository
	indexRegistry  IndexRegistry
	sectorRegistry SectorRegistry
}

// NewWatchlistService creates a new WatchlistService.
func NewWatchlistService(
	watchlists watchlist.Repository,
	items watchlist.ItemRepository,
	stockData stock.Repository,
	indexRegistry IndexRegistry,
	sectorRegistry SectorRegistry,
) *WatchlistService {
	return &WatchlistService{
		watchlists:     watchlists,
		items:          items,
		stockData:      stockData,
		indexRegistry:  indexRegistry,
		sectorRegistry: sectorRegistry,
	}
}

// CreateWatchlist validates and persists a new watchlist for the given profile.
func (s *WatchlistService) CreateWatchlist(ctx context.Context, profileID, name string) (*watchlist.Watchlist, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyName
	}

	existing, err := s.watchlists.ListByProfileID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	for _, w := range existing {
		if strings.EqualFold(w.Name, name) {
			return nil, ErrWatchlistNameTaken
		}
	}

	w := watchlist.NewWatchlist(profileID, name)
	if err := s.watchlists.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

// ListWatchlists returns all watchlists for the given profile.
func (s *WatchlistService) ListWatchlists(ctx context.Context, profileID string) ([]*watchlist.Watchlist, error) {
	return s.watchlists.ListByProfileID(ctx, profileID)
}

// RenameWatchlist validates and updates the name of the given watchlist.
func (s *WatchlistService) RenameWatchlist(ctx context.Context, id, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}

	w, err := s.watchlists.GetByID(ctx, id)
	if err != nil {
		return err
	}

	siblings, err := s.watchlists.ListByProfileID(ctx, w.ProfileID)
	if err != nil {
		return err
	}
	for _, sib := range siblings {
		if sib.ID == id {
			continue
		}
		if strings.EqualFold(sib.Name, name) {
			return ErrWatchlistNameTaken
		}
	}

	w.Name = name
	w.UpdatedAt = time.Now().UTC()
	return s.watchlists.Update(ctx, w)
}

// DeleteWatchlist removes the watchlist with the given ID.
func (s *WatchlistService) DeleteWatchlist(ctx context.Context, id string) error {
	return s.watchlists.Delete(ctx, id)
}

// AddTicker adds a ticker to the given watchlist, returning an error if it already exists.
func (s *WatchlistService) AddTicker(ctx context.Context, watchlistID, ticker string) error {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if ticker == "" {
		return ErrEmptyTicker
	}

	exists, err := s.items.ExistsByWatchlistAndTicker(ctx, watchlistID, ticker)
	if err != nil {
		return err
	}
	if exists {
		return watchlist.ErrTickerAlreadyInWatchlist
	}

	item := watchlist.NewItem(watchlistID, ticker)
	return s.items.Add(ctx, item)
}

// RemoveTicker removes a ticker from the given watchlist.
func (s *WatchlistService) RemoveTicker(ctx context.Context, watchlistID, ticker string) error {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if ticker == "" {
		return ErrEmptyTicker
	}
	return s.items.Remove(ctx, watchlistID, ticker)
}

// ListItems returns all items in the watchlist, enriched with stock data and valuation.
// If sectorFilter is non-empty, only items matching that sector are returned.
func (s *WatchlistService) ListItems(
	ctx context.Context,
	watchlistID, sectorFilter string,
) ([]*WatchlistItemWithData, error) {
	items, err := s.items.ListByWatchlistID(ctx, watchlistID)
	if err != nil {
		return nil, err
	}

	tickers := make([]string, len(items))
	for i, item := range items {
		tickers[i] = item.Ticker
	}

	return s.enrichTickers(ctx, tickers, sectorFilter), nil
}

// ListPresetItems resolves tickers from the named index and returns enriched items.
// Returns ErrUnknownIndex if the index name is not registered.
func (s *WatchlistService) ListPresetItems(
	ctx context.Context,
	indexName, sectorFilter string,
) ([]*WatchlistItemWithData, error) {
	tickers, ok := s.indexRegistry.Tickers(indexName)
	if !ok {
		return nil, ErrUnknownIndex
	}

	return s.enrichTickers(ctx, tickers, sectorFilter), nil
}

// ListIndexNames returns a sorted list of all registered index names.
func (s *WatchlistService) ListIndexNames() []string {
	return s.indexRegistry.Names()
}

// ListSectors returns a sorted list of all unique sector names.
func (s *WatchlistService) ListSectors() []string {
	return s.sectorRegistry.AllSectors()
}

// enrichTickers fetches stock data and valuation for each ticker, then applies
// an optional sector filter.
func (s *WatchlistService) enrichTickers(
	ctx context.Context,
	tickers []string,
	sectorFilter string,
) []*WatchlistItemWithData {
	result := make([]*WatchlistItemWithData, 0, len(tickers))

	for _, ticker := range tickers {
		sector := s.sectorRegistry.SectorOf(ticker)
		if sectorFilter != "" && sector != sectorFilter {
			continue
		}

		item := &WatchlistItemWithData{
			Ticker: ticker,
			Sector: sector,
		}

		data, err := s.stockData.GetByTicker(ctx, ticker)
		if err == nil {
			item.StockData = data
			input := valuation.ValuationInput{
				Ticker:      data.Ticker,
				Price:       data.Price,
				EPS:         data.EPS,
				BVPS:        data.BVPS,
				PBV:         data.PBV,
				PER:         data.PER,
				RiskProfile: valuation.RiskModerate, // default for watchlist browsing
			}
			val, valErr := valuation.Evaluate(input)
			if valErr == nil {
				item.Valuation = val
			}
		}

		result = append(result, item)
	}

	return result
}
