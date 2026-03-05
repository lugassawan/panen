package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
)

// newTestDB opens an in-memory SQLite database, applies all migrations,
// and registers cleanup with t.Cleanup.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := Migrate(context.Background(), db.Conn()); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db.Conn()
}

// dbTestFixture provides a pre-created user, brokerage account, and portfolio
// for tests that need a valid portfolio as a foreign-key parent.
type dbTestFixture struct {
	DB          *sql.DB
	PortfolioID string
	Ctx         context.Context
	Now         time.Time
}

// setupDBFixture creates a user, brokerage account, and portfolio in an
// in-memory database. Use the returned fixture's DB to create repo instances.
func setupDBFixture(t *testing.T, mode portfolio.Mode) dbTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)
	p := &user.Profile{ID: shared.NewID(), Name: "Test", CreatedAt: now, UpdatedAt: now}
	if err := userRepo.Create(ctx, p); err != nil {
		t.Fatalf("create profile: %v", err)
	}

	brokerRepo := NewBrokerageRepo(db)
	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: p.ID, BrokerName: "Test", BrokerCode: "TST",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := brokerRepo.Create(ctx, acct); err != nil {
		t.Fatalf("create brokerage: %v", err)
	}

	portRepo := NewPortfolioRepo(db)
	port := &portfolio.Portfolio{
		ID: shared.NewID(), BrokerageAccountID: acct.ID, Name: "Test",
		Mode: mode, RiskProfile: portfolio.RiskProfileModerate,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := portRepo.Create(ctx, port); err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	return dbTestFixture{
		DB:          db,
		PortfolioID: port.ID,
		Ctx:         ctx,
		Now:         now,
	}
}
