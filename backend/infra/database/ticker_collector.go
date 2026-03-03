package database

import (
	"context"
	"database/sql"
	"sort"
)

const tickerCollectAll = `
SELECT DISTINCT ticker FROM watchlist_items
UNION
SELECT DISTINCT ticker FROM holdings`

// TickerCollector collects unique tickers across watchlists and holdings.
type TickerCollector struct {
	db *sql.DB
}

// NewTickerCollector creates a new TickerCollector.
func NewTickerCollector(db *sql.DB) *TickerCollector {
	return &TickerCollector{db: db}
}

// CollectAll returns sorted unique tickers from both watchlist items and holdings.
func (c *TickerCollector) CollectAll(ctx context.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, tickerCollectAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickers []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tickers = append(tickers, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Strings(tickers)
	return tickers, nil
}
