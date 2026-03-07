package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
)

const (
	sellTxInsert = `INSERT INTO sell_transactions
		(id, holding_id, date, price, lots, fee, tax, realized_gain, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	sellTxGetByID = `SELECT id, holding_id, date, price, lots, fee, tax, realized_gain, created_at
		FROM sell_transactions WHERE id = ?`
	sellTxListByHoldingID = `SELECT id, holding_id, date, price, lots, fee, tax, realized_gain, created_at
		FROM sell_transactions WHERE holding_id = ? ORDER BY date`
	sellTxDelete = `DELETE FROM sell_transactions WHERE id = ?`
)

type SellTransactionRepo struct {
	db *sql.DB
}

func NewSellTransactionRepo(db *sql.DB) *SellTransactionRepo {
	return &SellTransactionRepo{db: db}
}

func (r *SellTransactionRepo) Create(ctx context.Context, tx *portfolio.SellTransaction) error {
	_, err := r.db.ExecContext(ctx, sellTxInsert,
		tx.ID, tx.HoldingID, formatTime(tx.Date), tx.Price, tx.Lots, tx.Fee,
		tx.Tax, tx.RealizedGain, formatTime(tx.CreatedAt))
	return err
}

func (r *SellTransactionRepo) GetByID(ctx context.Context, id string) (*portfolio.SellTransaction, error) {
	var tx portfolio.SellTransaction
	var date, createdAt string
	err := r.db.QueryRowContext(ctx, sellTxGetByID, id).Scan(
		&tx.ID, &tx.HoldingID, &date, &tx.Price, &tx.Lots, &tx.Fee,
		&tx.Tax, &tx.RealizedGain, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if tx.Date, err = parseTime(date); err != nil {
		return nil, err
	}
	if tx.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *SellTransactionRepo) ListByHoldingID(
	ctx context.Context, holdingID string,
) ([]*portfolio.SellTransaction, error) {
	rows, err := r.db.QueryContext(ctx, sellTxListByHoldingID, holdingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []*portfolio.SellTransaction
	for rows.Next() {
		var tx portfolio.SellTransaction
		var date, createdAt string
		if err := rows.Scan(
			&tx.ID, &tx.HoldingID, &date, &tx.Price, &tx.Lots, &tx.Fee,
			&tx.Tax, &tx.RealizedGain, &createdAt,
		); err != nil {
			return nil, err
		}
		if tx.Date, err = parseTime(date); err != nil {
			return nil, err
		}
		if tx.CreatedAt, err = parseTime(createdAt); err != nil {
			return nil, err
		}
		txns = append(txns, &tx)
	}
	return txns, rows.Err()
}

func (r *SellTransactionRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, sellTxDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}
