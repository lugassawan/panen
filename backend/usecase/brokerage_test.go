package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/brokerconfig"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
)

func TestBrokerageServiceCreateHappy(t *testing.T) {
	repo := newMockBrokerageRepo()
	svc := NewBrokerageService(repo, newMockPortfolioRepo(), nil)

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Ajaib",
		BrokerCode: "AJAIB", BuyFeePct: 0.15, SellFeePct: 0.25, SellTaxPct: 0.1,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := svc.Create(context.Background(), acct); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), acct.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.BrokerName != "Ajaib" {
		t.Errorf("BrokerName = %q, want Ajaib", got.BrokerName)
	}
}

func TestBrokerageServiceCreateEmptyName(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	acct := &brokerage.Account{ID: shared.NewID(), BrokerName: "  "}
	err := svc.Create(context.Background(), acct)
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("Create() error = %v, want ErrEmptyName", err)
	}
}

func TestBrokerageServiceListByProfileIDHappy(t *testing.T) {
	repo := newMockBrokerageRepo()
	svc := NewBrokerageService(repo, newMockPortfolioRepo(), nil)
	ctx := context.Background()

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Ajaib",
		BuyFeePct: 0.15, SellFeePct: 0.25,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, acct); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := svc.ListByProfileID(ctx, "p1")
	if err != nil {
		t.Fatalf("ListByProfileID() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != acct.ID {
		t.Errorf("ID = %q, want %q", got[0].ID, acct.ID)
	}
}

func TestBrokerageServiceListByProfileIDEmpty(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	got, err := svc.ListByProfileID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("ListByProfileID() error = %v", err)
	}
	if got == nil {
		// nil slice is acceptable for empty results
		return
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func negativeFeeTests() []struct {
	name    string
	buyFee  float64
	sellFee float64
	sellTax float64
} {
	return []struct {
		name    string
		buyFee  float64
		sellFee float64
		sellTax float64
	}{
		{name: "negative buy fee", buyFee: -0.1, sellFee: 0, sellTax: 0},
		{name: "negative sell fee", buyFee: 0, sellFee: -0.1, sellTax: 0},
		{name: "negative sell tax", buyFee: 0, sellFee: 0, sellTax: -0.1},
	}
}

func TestBrokerageServiceCreateNegativeFee(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	for _, tt := range negativeFeeTests() {
		t.Run(tt.name, func(t *testing.T) {
			acct := &brokerage.Account{
				ID: shared.NewID(), BrokerName: "Broker",
				BuyFeePct: tt.buyFee, SellFeePct: tt.sellFee, SellTaxPct: tt.sellTax,
			}
			err := svc.Create(context.Background(), acct)
			if !errors.Is(err, ErrInvalidFee) {
				t.Errorf("Create() error = %v, want ErrInvalidFee", err)
			}
		})
	}
}

func TestBrokerageServiceGetByIDHappy(t *testing.T) {
	repo := newMockBrokerageRepo()
	svc := NewBrokerageService(repo, newMockPortfolioRepo(), nil)
	ctx := context.Background()

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "IPOT",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, acct); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := svc.GetByID(ctx, acct.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.BrokerName != "IPOT" {
		t.Errorf("BrokerName = %q, want IPOT", got.BrokerName)
	}
}

func TestBrokerageServiceGetByIDEmptyID(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	_, err := svc.GetByID(context.Background(), "")
	if !errors.Is(err, ErrEmptyID) {
		t.Errorf("GetByID() error = %v, want ErrEmptyID", err)
	}
}

func TestBrokerageServiceUpdateHappy(t *testing.T) {
	repo := newMockBrokerageRepo()
	svc := NewBrokerageService(repo, newMockPortfolioRepo(), nil)
	ctx := context.Background()

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Stockbit",
		BuyFeePct: 0.15, SellFeePct: 0.25,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, acct); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	acct.BrokerName = "Bibit"
	acct.BuyFeePct = 0.20
	if err := svc.Update(ctx, acct); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.GetByID(ctx, acct.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.BrokerName != "Bibit" {
		t.Errorf("BrokerName = %q, want Bibit", got.BrokerName)
	}
	if got.BuyFeePct != 0.20 {
		t.Errorf("BuyFeePct = %f, want 0.20", got.BuyFeePct)
	}
}

func TestBrokerageServiceUpdateEmptyName(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	acct := &brokerage.Account{ID: shared.NewID(), BrokerName: ""}
	err := svc.Update(context.Background(), acct)
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("Update() error = %v, want ErrEmptyName", err)
	}
}

func TestBrokerageServiceUpdateNegativeFees(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	for _, tt := range negativeFeeTests() {
		t.Run(tt.name, func(t *testing.T) {
			acct := &brokerage.Account{
				ID: shared.NewID(), BrokerName: "Broker",
				BuyFeePct: tt.buyFee, SellFeePct: tt.sellFee, SellTaxPct: tt.sellTax,
			}
			err := svc.Update(context.Background(), acct)
			if !errors.Is(err, ErrInvalidFee) {
				t.Errorf("Update() error = %v, want ErrInvalidFee", err)
			}
		})
	}
}

func TestBrokerageServiceDeleteHappy(t *testing.T) {
	brokerageRepo := newMockBrokerageRepo()
	svc := NewBrokerageService(brokerageRepo, newMockPortfolioRepo(), nil)
	ctx := context.Background()

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "IPOT",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := brokerageRepo.Create(ctx, acct); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := svc.Delete(ctx, acct.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := brokerageRepo.GetByID(ctx, acct.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() after delete error = %v, want ErrNotFound", err)
	}
}

func TestBrokerageServiceDeleteHasPortfolios(t *testing.T) {
	brokerageRepo := newMockBrokerageRepo()
	portfolioRepo := newMockPortfolioRepo()
	svc := NewBrokerageService(brokerageRepo, portfolioRepo, nil)
	ctx := context.Background()

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "IPOT",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := brokerageRepo.Create(ctx, acct); err != nil {
		t.Fatalf("Create brokerage error = %v", err)
	}

	p := &portfolio.Portfolio{
		ID: shared.NewID(), BrokerageAccountID: acct.ID,
		Name: "My Portfolio", Mode: portfolio.ModeValue,
		RiskProfile: portfolio.RiskProfileModerate,
		CreatedAt:   time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := portfolioRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create portfolio error = %v", err)
	}

	err := svc.Delete(ctx, acct.ID)
	if !errors.Is(err, ErrHasDependents) {
		t.Errorf("Delete() error = %v, want ErrHasDependents", err)
	}
}

func TestBrokerageServiceDeleteEmptyID(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	err := svc.Delete(context.Background(), "")
	if !errors.Is(err, ErrEmptyID) {
		t.Errorf("Delete() error = %v, want ErrEmptyID", err)
	}
}

func syncFeeConfigs() []*brokerconfig.BrokerConfig {
	return []*brokerconfig.BrokerConfig{
		{Code: "AJ", BuyFeePct: 0.20, SellFeePct: 0.30, SellTaxPct: 0.1},
		{Code: "ST", BuyFeePct: 0.12, SellFeePct: 0.22, SellTaxPct: 0.1},
	}
}

func seedAccount(t *testing.T, repo *mockBrokerageRepo, acct *brokerage.Account) {
	t.Helper()
	if err := repo.Create(context.Background(), acct); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestSyncFeesFromConfigMatchingAccounts(t *testing.T) {
	repo := newMockBrokerageRepo()
	emitter := &mockEventEmitter{}
	svc := NewBrokerageService(repo, newMockPortfolioRepo(), emitter)
	ctx := context.Background()

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Ajaib",
		BrokerCode: "AJ", BuyFeePct: 0.15, SellFeePct: 0.25, SellTaxPct: 0.1,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	seedAccount(t, repo, acct)

	count, err := svc.SyncFeesFromConfig(ctx, "p1", syncFeeConfigs())
	if err != nil {
		t.Fatalf("SyncFeesFromConfig() error = %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	got, _ := repo.GetByID(ctx, acct.ID)
	if got.BuyFeePct != 0.20 {
		t.Errorf("BuyFeePct = %v, want 0.20", got.BuyFeePct)
	}
	if got.SellFeePct != 0.30 {
		t.Errorf("SellFeePct = %v, want 0.30", got.SellFeePct)
	}

	events := emitter.eventsByName(shared.EventBrokerFeesSynced)
	if len(events) != 1 {
		t.Fatalf("expected 1 %s event, got %d", shared.EventBrokerFeesSynced, len(events))
	}
}

func TestSyncFeesFromConfigSkipsManual(t *testing.T) {
	repo := newMockBrokerageRepo()
	svc := NewBrokerageService(repo, newMockPortfolioRepo(), nil)

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Ajaib",
		BrokerCode: "AJ", BuyFeePct: 0.15, SellFeePct: 0.25, SellTaxPct: 0.1,
		IsManualFee: true,
		CreatedAt:   time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	seedAccount(t, repo, acct)

	count, err := svc.SyncFeesFromConfig(context.Background(), "p1", syncFeeConfigs())
	if err != nil {
		t.Fatalf("SyncFeesFromConfig() error = %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestSyncFeesFromConfigSkipsUnmatched(t *testing.T) {
	repo := newMockBrokerageRepo()
	svc := NewBrokerageService(repo, newMockPortfolioRepo(), nil)

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Custom",
		BrokerCode: "XX", BuyFeePct: 0.15, SellFeePct: 0.25, SellTaxPct: 0.1,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	seedAccount(t, repo, acct)

	count, err := svc.SyncFeesFromConfig(context.Background(), "p1", syncFeeConfigs())
	if err != nil {
		t.Fatalf("SyncFeesFromConfig() error = %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestSyncFeesFromConfigSkipsAlreadyMatching(t *testing.T) {
	repo := newMockBrokerageRepo()
	svc := NewBrokerageService(repo, newMockPortfolioRepo(), nil)

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Ajaib",
		BrokerCode: "AJ", BuyFeePct: 0.20, SellFeePct: 0.30, SellTaxPct: 0.1,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	seedAccount(t, repo, acct)

	count, err := svc.SyncFeesFromConfig(context.Background(), "p1", syncFeeConfigs())
	if err != nil {
		t.Fatalf("SyncFeesFromConfig() error = %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestSyncFeesFromConfigEmptyConfigs(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	count, err := svc.SyncFeesFromConfig(context.Background(), "p1", nil)
	if err != nil {
		t.Fatalf("SyncFeesFromConfig() error = %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}
