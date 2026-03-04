package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
)

type checklistTestFixture struct {
	repo          *ChecklistResultRepo
	portfolioRepo *PortfolioRepo
	portfolioID   string
	ctx           context.Context
	now           time.Time
}

func setupChecklistTest(t *testing.T) checklistTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)
	broRepo := NewBrokerageRepo(db)
	portRepo := NewPortfolioRepo(db)

	p := &user.Profile{
		ID:        shared.NewID(),
		Name:      "Test User",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := userRepo.Create(ctx, p); err != nil {
		t.Fatalf("create profile: %v", err)
	}

	a := &brokerage.Account{
		ID:         shared.NewID(),
		ProfileID:  p.ID,
		BrokerName: "Broker",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := broRepo.Create(ctx, a); err != nil {
		t.Fatalf("create brokerage: %v", err)
	}

	port := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: a.ID,
		Name:               "Test Portfolio",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := portRepo.Create(ctx, port); err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	return checklistTestFixture{
		repo:          NewChecklistResultRepo(db),
		portfolioRepo: portRepo,
		portfolioID:   port.ID,
		ctx:           ctx,
		now:           now,
	}
}

func TestChecklistResultRepoUpsertAndGet(t *testing.T) {
	f := setupChecklistTest(t)

	cr := &checklist.ChecklistResult{
		ID:          shared.NewID(),
		PortfolioID: f.portfolioID,
		Ticker:      "BBCA",
		Action:      checklist.ActionBuy,
		ManualChecks: map[string]bool{
			"management_quality": true,
			"competitive_moat":   false,
		},
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Upsert(f.ctx, cr); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := f.repo.Get(f.ctx, f.portfolioID, "BBCA", checklist.ActionBuy)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != cr.ID {
		t.Errorf("ID = %q, want %q", got.ID, cr.ID)
	}
	if got.PortfolioID != f.portfolioID {
		t.Errorf("PortfolioID = %q, want %q", got.PortfolioID, f.portfolioID)
	}
	if got.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want %q", got.Ticker, "BBCA")
	}
	if got.Action != checklist.ActionBuy {
		t.Errorf("Action = %q, want %q", got.Action, checklist.ActionBuy)
	}
	if len(got.ManualChecks) != 2 {
		t.Fatalf("ManualChecks len = %d, want 2", len(got.ManualChecks))
	}
	if !got.ManualChecks["management_quality"] {
		t.Error("ManualChecks[management_quality] = false, want true")
	}
	if got.ManualChecks["competitive_moat"] {
		t.Error("ManualChecks[competitive_moat] = true, want false")
	}
	if !got.CreatedAt.Equal(f.now) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, f.now)
	}
	if !got.UpdatedAt.Equal(f.now) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, f.now)
	}
}

func TestChecklistResultRepoGetNotFound(t *testing.T) {
	f := setupChecklistTest(t)

	_, err := f.repo.Get(f.ctx, f.portfolioID, "NONEXIST", checklist.ActionBuy)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestChecklistResultRepoUpsertConflictUpdate(t *testing.T) {
	f := setupChecklistTest(t)

	cr := &checklist.ChecklistResult{
		ID:          shared.NewID(),
		PortfolioID: f.portfolioID,
		Ticker:      "BBRI",
		Action:      checklist.ActionAverageDown,
		ManualChecks: map[string]bool{
			"check_a": false,
		},
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Upsert(f.ctx, cr); err != nil {
		t.Fatalf("Upsert() first error = %v", err)
	}

	updatedAt := f.now.Add(time.Hour)
	cr2 := &checklist.ChecklistResult{
		ID:          shared.NewID(),
		PortfolioID: f.portfolioID,
		Ticker:      "BBRI",
		Action:      checklist.ActionAverageDown,
		ManualChecks: map[string]bool{
			"check_a": true,
			"check_b": true,
		},
		CreatedAt: f.now,
		UpdatedAt: updatedAt,
	}
	if err := f.repo.Upsert(f.ctx, cr2); err != nil {
		t.Fatalf("Upsert() conflict update error = %v", err)
	}

	got, err := f.repo.Get(f.ctx, f.portfolioID, "BBRI", checklist.ActionAverageDown)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	// ID should remain the original since ON CONFLICT updates only manual_checks and updated_at.
	if got.ID != cr.ID {
		t.Errorf("ID = %q, want original %q", got.ID, cr.ID)
	}
	if len(got.ManualChecks) != 2 {
		t.Fatalf("ManualChecks len = %d, want 2", len(got.ManualChecks))
	}
	if !got.ManualChecks["check_a"] {
		t.Error("ManualChecks[check_a] = false, want true")
	}
	if !got.ManualChecks["check_b"] {
		t.Error("ManualChecks[check_b] = false, want true")
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, updatedAt)
	}
}

func TestChecklistResultRepoDeleteByID(t *testing.T) {
	f := setupChecklistTest(t)

	cr := &checklist.ChecklistResult{
		ID:           shared.NewID(),
		PortfolioID:  f.portfolioID,
		Ticker:       "TLKM",
		Action:       checklist.ActionHold,
		ManualChecks: map[string]bool{},
		CreatedAt:    f.now,
		UpdatedAt:    f.now,
	}
	if err := f.repo.Upsert(f.ctx, cr); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := f.repo.Delete(f.ctx, cr.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := f.repo.Get(f.ctx, f.portfolioID, "TLKM", checklist.ActionHold)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Get() after delete error = %v, want ErrNotFound", err)
	}
}

func TestChecklistResultRepoDeleteByIDNotFound(t *testing.T) {
	f := setupChecklistTest(t)

	err := f.repo.Delete(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestChecklistResultRepoDeleteByPortfolioID(t *testing.T) {
	f := setupChecklistTest(t)

	actions := []checklist.ActionType{
		checklist.ActionBuy,
		checklist.ActionSellExit,
		checklist.ActionHold,
	}
	for _, action := range actions {
		cr := &checklist.ChecklistResult{
			ID:           shared.NewID(),
			PortfolioID:  f.portfolioID,
			Ticker:       "BMRI",
			Action:       action,
			ManualChecks: map[string]bool{},
			CreatedAt:    f.now,
			UpdatedAt:    f.now,
		}
		if err := f.repo.Upsert(f.ctx, cr); err != nil {
			t.Fatalf("Upsert() %s error = %v", action, err)
		}
	}

	if err := f.repo.DeleteByPortfolioID(f.ctx, f.portfolioID); err != nil {
		t.Fatalf("DeleteByPortfolioID() error = %v", err)
	}

	for _, action := range actions {
		_, err := f.repo.Get(f.ctx, f.portfolioID, "BMRI", action)
		if !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Get() %s after DeleteByPortfolioID error = %v, want ErrNotFound", action, err)
		}
	}
}

func TestChecklistResultRepoDeleteByPortfolioIDEmpty(t *testing.T) {
	f := setupChecklistTest(t)

	// Should not error even when no rows match.
	if err := f.repo.DeleteByPortfolioID(f.ctx, f.portfolioID); err != nil {
		t.Fatalf("DeleteByPortfolioID() with no rows error = %v", err)
	}
}

func TestChecklistResultRepoCascadeDeleteOnPortfolio(t *testing.T) {
	f := setupChecklistTest(t)

	cr := &checklist.ChecklistResult{
		ID:           shared.NewID(),
		PortfolioID:  f.portfolioID,
		Ticker:       "ASII",
		Action:       checklist.ActionSellStop,
		ManualChecks: map[string]bool{"stop_loss_hit": true},
		CreatedAt:    f.now,
		UpdatedAt:    f.now,
	}
	if err := f.repo.Upsert(f.ctx, cr); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	// Verify it exists before cascade.
	_, err := f.repo.Get(f.ctx, f.portfolioID, "ASII", checklist.ActionSellStop)
	if err != nil {
		t.Fatalf("Get() before cascade error = %v", err)
	}

	// Delete the portfolio — should cascade to checklist_results.
	if err := f.portfolioRepo.Delete(f.ctx, f.portfolioID); err != nil {
		t.Fatalf("Delete() portfolio error = %v", err)
	}

	_, err = f.repo.Get(f.ctx, f.portfolioID, "ASII", checklist.ActionSellStop)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Get() after cascade delete error = %v, want ErrNotFound", err)
	}
}
