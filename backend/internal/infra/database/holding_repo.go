package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lugassawan/panen/backend/internal/domain/portfolio"
	"github.com/lugassawan/panen/backend/internal/domain/shared"
)

const (
	holdingInsert = `INSERT INTO holdings
		(id, portfolio_id, ticker, avg_buy_price, lots, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	holdingGetByID = `SELECT id, portfolio_id, ticker, avg_buy_price, lots, created_at, updated_at
		FROM holdings WHERE id = ?`
	holdingGetByPortfolioAndTicker = `SELECT id, portfolio_id, ticker, avg_buy_price, lots, created_at, updated_at
		FROM holdings WHERE portfolio_id = ? AND ticker = ?`
	holdingListByPortfolioID = `SELECT id, portfolio_id, ticker, avg_buy_price, lots, created_at, updated_at
		FROM holdings WHERE portfolio_id = ? ORDER BY ticker`
	holdingUpdate = `UPDATE holdings SET ticker = ?, avg_buy_price = ?, lots = ?, updated_at = ?
		WHERE id = ?`
	holdingDelete = `DELETE FROM holdings WHERE id = ?`
)

// HoldingRepo implements portfolio.HoldingRepository.
type HoldingRepo struct {
	db *sql.DB
}

// NewHoldingRepo creates a new HoldingRepo.
func NewHoldingRepo(db *sql.DB) *HoldingRepo {
	return &HoldingRepo{db: db}
}

func (r *HoldingRepo) Create(ctx context.Context, h *portfolio.Holding) error {
	_, err := r.db.ExecContext(ctx, holdingInsert,
		h.ID, h.PortfolioID, h.Ticker, h.AvgBuyPrice, h.Lots,
		formatTime(h.CreatedAt), formatTime(h.UpdatedAt))
	return err
}

func (r *HoldingRepo) GetByID(ctx context.Context, id string) (*portfolio.Holding, error) {
	var h portfolio.Holding
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, holdingGetByID, id).Scan(
		&h.ID, &h.PortfolioID, &h.Ticker, &h.AvgBuyPrice, &h.Lots,
		&createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if h.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if h.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HoldingRepo) GetByPortfolioAndTicker(
	ctx context.Context,
	portfolioID, ticker string,
) (*portfolio.Holding, error) {
	var h portfolio.Holding
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, holdingGetByPortfolioAndTicker, portfolioID, ticker).Scan(
		&h.ID, &h.PortfolioID, &h.Ticker, &h.AvgBuyPrice, &h.Lots,
		&createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if h.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if h.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HoldingRepo) ListByPortfolioID(ctx context.Context, portfolioID string) ([]*portfolio.Holding, error) {
	rows, err := r.db.QueryContext(ctx, holdingListByPortfolioID, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holdings []*portfolio.Holding
	for rows.Next() {
		var h portfolio.Holding
		var createdAt, updatedAt string
		if err := rows.Scan(&h.ID, &h.PortfolioID, &h.Ticker, &h.AvgBuyPrice, &h.Lots,
			&createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if h.CreatedAt, err = parseTime(createdAt); err != nil {
			return nil, err
		}
		if h.UpdatedAt, err = parseTime(updatedAt); err != nil {
			return nil, err
		}
		holdings = append(holdings, &h)
	}
	return holdings, rows.Err()
}

func (r *HoldingRepo) Update(ctx context.Context, h *portfolio.Holding) error {
	res, err := r.db.ExecContext(ctx, holdingUpdate,
		h.Ticker, h.AvgBuyPrice, h.Lots, formatTime(h.UpdatedAt), h.ID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *HoldingRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, holdingDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}
