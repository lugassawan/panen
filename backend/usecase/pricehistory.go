package usecase

import (
	"context"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
)

// PriceHistoryService handles on-demand fetching and caching of price history.
type PriceHistoryService struct {
	historyRepo stock.PriceHistoryRepository
	provider    stock.DataProvider
}

// NewPriceHistoryService creates a new PriceHistoryService.
func NewPriceHistoryService(
	historyRepo stock.PriceHistoryRepository,
	provider stock.DataProvider,
) *PriceHistoryService {
	return &PriceHistoryService{
		historyRepo: historyRepo,
		provider:    provider,
	}
}

// GetHistory returns cached price history, refreshing from the provider if stale.
func (s *PriceHistoryService) GetHistory(
	ctx context.Context,
	ticker string,
) ([]stock.PricePoint, error) {
	latest, err := s.historyRepo.LatestDate(ctx, ticker, s.provider.Source())
	if err != nil {
		return nil, err
	}

	if !isFresh(latest) {
		points, err := s.provider.FetchPriceHistory(ctx, ticker)
		if err != nil {
			return nil, err
		}
		if err := s.historyRepo.BulkUpsert(ctx, points); err != nil {
			return nil, err
		}
	}

	return s.historyRepo.GetByTicker(ctx, ticker, s.provider.Source())
}

// isFresh returns true if the latest date is within the last 3 days (UTC).
// A 3-day window avoids unnecessary re-fetches on weekends and single-day holidays.
func isFresh(latest time.Time) bool {
	if latest.IsZero() {
		return false
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	cutoff := today.AddDate(0, 0, -3)
	latestDay := latest.UTC().Truncate(24 * time.Hour)
	return !latestDay.Before(cutoff)
}
