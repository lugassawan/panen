package database

import (
	"context"
	"database/sql"

	"github.com/lugassawan/panen/backend/domain/portfolio"
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
	return QueryRow(ctx, r.db, txGetByID, scanBuyTransaction, id)
}

func (r *BuyTransactionRepo) ListByHoldingID(
	ctx context.Context, holdingID string,
) ([]*portfolio.BuyTransaction, error) {
	return QueryAll(ctx, r.db, txListByHoldingID, scanBuyTransaction, holdingID)
}

func (r *BuyTransactionRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, txDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func scanBuyTransaction(scan func(dest ...any) error) (*portfolio.BuyTransaction, error) {
	var tx portfolio.BuyTransaction
	var date, createdAt string
	if err := scan(&tx.ID, &tx.HoldingID, &date, &tx.Price, &tx.Lots, &tx.Fee, &createdAt); err != nil {
		return nil, err
	}
	var err error
	if tx.Date, err = parseTime(date); err != nil {
		return nil, err
	}
	if tx.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	return &tx, nil
}
