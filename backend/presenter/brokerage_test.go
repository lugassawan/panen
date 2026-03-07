package presenter

import (
	"context"
	"testing"

	"github.com/lugassawan/panen/backend/usecase"
)

func newTestBrokerageHandler() *BrokerageHandler {
	ctx := context.Background()
	brokerageRepo := newMockBrokerageRepo()
	portfolioRepo := newMockPortfolioRepo()
	svc := usecase.NewBrokerageService(brokerageRepo, portfolioRepo, nil)
	return NewBrokerageHandler(ctx, "profile-1", svc)
}

func TestBrokerageHandlerCreateAndGet(t *testing.T) {
	handler := newTestBrokerageHandler()

	resp, err := handler.CreateBrokerageAccount("Ajaib", "AJ", 0.15, 0.25, 0.1, true)
	if err != nil {
		t.Fatalf("CreateBrokerageAccount() error = %v", err)
	}
	if resp.BrokerName != "Ajaib" {
		t.Errorf("BrokerName = %q, want %q", resp.BrokerName, "Ajaib")
	}
	if resp.BrokerCode != "AJ" {
		t.Errorf("BrokerCode = %q, want %q", resp.BrokerCode, "AJ")
	}
	if resp.BuyFeePct != 0.15 {
		t.Errorf("BuyFeePct = %v, want 0.15", resp.BuyFeePct)
	}
	if !resp.IsManualFee {
		t.Error("expected IsManualFee = true")
	}
	if resp.ID == "" {
		t.Error("expected non-empty ID")
	}

	got, err := handler.GetBrokerageAccount(resp.ID)
	if err != nil {
		t.Fatalf("GetBrokerageAccount() error = %v", err)
	}
	if got.ID != resp.ID {
		t.Errorf("ID = %q, want %q", got.ID, resp.ID)
	}
}

func TestBrokerageHandlerUpdate(t *testing.T) {
	handler := newTestBrokerageHandler()

	resp, err := handler.CreateBrokerageAccount("Ajaib", "AJ", 0.15, 0.25, 0.1, false)
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}

	updated, err := handler.UpdateBrokerageAccount(resp.ID, "Stockbit", "SB", 0.10, 0.20, 0.1, true)
	if err != nil {
		t.Fatalf("Update error = %v", err)
	}
	if updated.BrokerName != "Stockbit" {
		t.Errorf("BrokerName = %q, want %q", updated.BrokerName, "Stockbit")
	}
	if updated.BrokerCode != "SB" {
		t.Errorf("BrokerCode = %q, want %q", updated.BrokerCode, "SB")
	}
	if !updated.IsManualFee {
		t.Error("expected IsManualFee = true after update")
	}
}

func TestBrokerageHandlerList(t *testing.T) {
	handler := newTestBrokerageHandler()

	_, _ = handler.CreateBrokerageAccount("Ajaib", "AJ", 0.15, 0.25, 0.1, false)
	_, _ = handler.CreateBrokerageAccount("Stockbit", "SB", 0.10, 0.20, 0.1, false)

	list, err := handler.ListBrokerageAccounts()
	if err != nil {
		t.Fatalf("ListBrokerageAccounts() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("got %d accounts, want 2", len(list))
	}
}

func TestBrokerageHandlerDelete(t *testing.T) {
	handler := newTestBrokerageHandler()

	resp, _ := handler.CreateBrokerageAccount("Ajaib", "AJ", 0.15, 0.25, 0.1, false)
	if err := handler.DeleteBrokerageAccount(resp.ID); err != nil {
		t.Fatalf("Delete error = %v", err)
	}

	_, err := handler.GetBrokerageAccount(resp.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestBrokerageHandlerCreateEmptyName(t *testing.T) {
	handler := newTestBrokerageHandler()

	_, err := handler.CreateBrokerageAccount("", "AJ", 0.15, 0.25, 0.1, false)
	if err == nil {
		t.Error("expected error for empty broker name")
	}
}
