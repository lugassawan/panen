package database

import (
	"context"
	"database/sql"

	"github.com/lugassawan/panen/backend/domain/brokerage"
)

const (
	brokerageInsert = `INSERT INTO brokerage_accounts
		(id, profile_id, broker_name, broker_code, buy_fee_pct, sell_fee_pct, sell_tax_pct,
		 is_manual_fee, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	brokerageGetByID = `SELECT id, profile_id, broker_name, broker_code, buy_fee_pct, sell_fee_pct,
		sell_tax_pct, is_manual_fee, created_at, updated_at FROM brokerage_accounts WHERE id = ?`
	brokerageListByProfileID = `SELECT id, profile_id, broker_name, broker_code, buy_fee_pct,
		sell_fee_pct, sell_tax_pct, is_manual_fee, created_at, updated_at FROM brokerage_accounts
		WHERE profile_id = ? ORDER BY created_at`
	brokerageListNonManual = `SELECT id, profile_id, broker_name, broker_code, buy_fee_pct,
		sell_fee_pct, sell_tax_pct, is_manual_fee, created_at, updated_at FROM brokerage_accounts
		WHERE profile_id = ? AND is_manual_fee = 0 ORDER BY created_at`
	brokerageUpdate = `UPDATE brokerage_accounts SET broker_name = ?, broker_code = ?, buy_fee_pct = ?,
		sell_fee_pct = ?, sell_tax_pct = ?, is_manual_fee = ?, updated_at = ? WHERE id = ?`
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
		a.ID, a.ProfileID, a.BrokerName, a.BrokerCode, a.BuyFeePct, a.SellFeePct,
		a.SellTaxPct, boolToInt(a.IsManualFee), formatTime(a.CreatedAt), formatTime(a.UpdatedAt))
	return err
}

func (r *BrokerageRepo) GetByID(ctx context.Context, id string) (*brokerage.Account, error) {
	return queryRow(ctx, r.db, brokerageGetByID, scanBrokerageAccount, id)
}

func (r *BrokerageRepo) ListByProfileID(ctx context.Context, profileID string) ([]*brokerage.Account, error) {
	return queryAll(ctx, r.db, brokerageListByProfileID, scanBrokerageAccount, profileID)
}

func (r *BrokerageRepo) ListNonManualByProfileID(ctx context.Context, profileID string) ([]*brokerage.Account, error) {
	return queryAll(ctx, r.db, brokerageListNonManual, scanBrokerageAccount, profileID)
}

func (r *BrokerageRepo) Update(ctx context.Context, a *brokerage.Account) error {
	res, err := r.db.ExecContext(ctx, brokerageUpdate,
		a.BrokerName, a.BrokerCode, a.BuyFeePct, a.SellFeePct,
		a.SellTaxPct, boolToInt(a.IsManualFee), formatTime(a.UpdatedAt), a.ID)
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

func scanBrokerageAccount(scan func(dest ...any) error) (*brokerage.Account, error) {
	var a brokerage.Account
	var isManual int
	var createdAt, updatedAt string
	if err := scan(&a.ID, &a.ProfileID, &a.BrokerName, &a.BrokerCode, &a.BuyFeePct,
		&a.SellFeePct, &a.SellTaxPct, &isManual, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	a.IsManualFee = isManual != 0
	var err error
	if a.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if a.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &a, nil
}
