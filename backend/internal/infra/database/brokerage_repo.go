package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lugassawan/panen/backend/internal/domain/brokerage"
	"github.com/lugassawan/panen/backend/internal/domain/shared"
)

const (
	brokerageInsert = `INSERT INTO brokerage_accounts
		(id, profile_id, broker_name, buy_fee_pct, sell_fee_pct, is_manual_fee, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	brokerageGetByID = `SELECT id, profile_id, broker_name, buy_fee_pct, sell_fee_pct,
		is_manual_fee, created_at, updated_at FROM brokerage_accounts WHERE id = ?`
	brokerageListByProfileID = `SELECT id, profile_id, broker_name, buy_fee_pct, sell_fee_pct,
		is_manual_fee, created_at, updated_at FROM brokerage_accounts
		WHERE profile_id = ? ORDER BY created_at`
	brokerageUpdate = `UPDATE brokerage_accounts SET broker_name = ?, buy_fee_pct = ?,
		sell_fee_pct = ?, is_manual_fee = ?, updated_at = ? WHERE id = ?`
	brokerageDelete = `DELETE FROM brokerage_accounts WHERE id = ?`
)

// BrokerageRepo implements brokerage.Repository.
type BrokerageRepo struct {
	db *sql.DB
}

// NewBrokerageRepo creates a new BrokerageRepo.
func NewBrokerageRepo(db *sql.DB) *BrokerageRepo {
	return &BrokerageRepo{db: db}
}

func (r *BrokerageRepo) Create(ctx context.Context, a *brokerage.Account) error {
	_, err := r.db.ExecContext(ctx, brokerageInsert,
		a.ID, a.ProfileID, a.BrokerName, a.BuyFeePct, a.SellFeePct,
		boolToInt(a.IsManualFee), formatTime(a.CreatedAt), formatTime(a.UpdatedAt))
	return err
}

func (r *BrokerageRepo) GetByID(ctx context.Context, id string) (*brokerage.Account, error) {
	var a brokerage.Account
	var isManual int
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, brokerageGetByID, id).Scan(
		&a.ID, &a.ProfileID, &a.BrokerName, &a.BuyFeePct, &a.SellFeePct,
		&isManual, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.IsManualFee = isManual != 0
	if a.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if a.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *BrokerageRepo) ListByProfileID(ctx context.Context, profileID string) ([]*brokerage.Account, error) {
	rows, err := r.db.QueryContext(ctx, brokerageListByProfileID, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*brokerage.Account
	for rows.Next() {
		var a brokerage.Account
		var isManual int
		var createdAt, updatedAt string
		if err := rows.Scan(&a.ID, &a.ProfileID, &a.BrokerName, &a.BuyFeePct,
			&a.SellFeePct, &isManual, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		a.IsManualFee = isManual != 0
		if a.CreatedAt, err = parseTime(createdAt); err != nil {
			return nil, err
		}
		if a.UpdatedAt, err = parseTime(updatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, &a)
	}
	return accounts, rows.Err()
}

func (r *BrokerageRepo) Update(ctx context.Context, a *brokerage.Account) error {
	res, err := r.db.ExecContext(ctx, brokerageUpdate,
		a.BrokerName, a.BuyFeePct, a.SellFeePct,
		boolToInt(a.IsManualFee), formatTime(a.UpdatedAt), a.ID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *BrokerageRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, brokerageDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}
