package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lugassawan/panen/backend/internal/domain/portfolio"
	"github.com/lugassawan/panen/backend/internal/domain/shared"
)

const (
	txInsert = `INSERT INTO buy_transactions
		(id, holding_id, date, price, lots, fee, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	txGetByID = `SELECT id, holding_id, date, price, lots, fee, created_at
		FROM buy_transactions WHERE id = ?`
	txListByHoldingID = `SELECT id, holding_id, date, price, lots, fee, created_at
		FROM buy_transactions WHERE holding_id = ? ORDER BY date`
	txDelete = `DELETE FROM buy_transactions WHERE id = ?`
)

// BuyTransactionRepo implements portfolio.BuyTransactionRepository.
type BuyTransactionRepo struct {
	db *sql.DB
}

// NewBuyTransactionRepo creates a new BuyTransactionRepo.
func NewBuyTransactionRepo(db *sql.DB) *BuyTransactionRepo {
	return &BuyTransactionRepo{db: db}
}

func (r *BuyTransactionRepo) Create(ctx context.Context, tx *portfolio.BuyTransaction) error {
	_, err := r.db.ExecContext(ctx, txInsert,
		tx.ID, tx.HoldingID, formatTime(tx.Date), tx.Price, tx.Lots, tx.Fee,
		formatTime(tx.CreatedAt))
	return err
}

func (r *BuyTransactionRepo) GetByID(ctx context.Context, id string) (*portfolio.BuyTransaction, error) {
	var tx portfolio.BuyTransaction
	var date, createdAt string
	err := r.db.QueryRowContext(ctx, txGetByID, id).Scan(
		&tx.ID, &tx.HoldingID, &date, &tx.Price, &tx.Lots, &tx.Fee, &createdAt)
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

func (r *BuyTransactionRepo) ListByHoldingID(
	ctx context.Context, holdingID string,
) ([]*portfolio.BuyTransaction, error) {
	rows, err := r.db.QueryContext(ctx, txListByHoldingID, holdingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []*portfolio.BuyTransaction
	for rows.Next() {
		var tx portfolio.BuyTransaction
		var date, createdAt string
		if err := rows.Scan(&tx.ID, &tx.HoldingID, &date, &tx.Price, &tx.Lots, &tx.Fee, &createdAt); err != nil {
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

func (r *BuyTransactionRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, txDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}
