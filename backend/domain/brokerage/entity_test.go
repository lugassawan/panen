package brokerage

import "testing"

func TestNewAccount(t *testing.T) {
	a := NewAccount("profile-1", "Ajaib", "AJ", 0.15, 0.25, 0.1)

	if a.ID == "" {
		t.Error("expected non-empty ID")
	}
	if a.ProfileID != "profile-1" {
		t.Errorf("ProfileID = %q, want %q", a.ProfileID, "profile-1")
	}
	if a.BrokerName != "Ajaib" {
		t.Errorf("BrokerName = %q, want %q", a.BrokerName, "Ajaib")
	}
	if a.BrokerCode != "AJ" {
		t.Errorf("BrokerCode = %q, want %q", a.BrokerCode, "AJ")
	}
	if a.BuyFeePct != 0.15 {
		t.Errorf("BuyFeePct = %v, want %v", a.BuyFeePct, 0.15)
	}
	if a.SellFeePct != 0.25 {
		t.Errorf("SellFeePct = %v, want %v", a.SellFeePct, 0.25)
	}
	if a.SellTaxPct != 0.1 {
		t.Errorf("SellTaxPct = %v, want %v", a.SellTaxPct, 0.1)
	}
	if a.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if a.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
	if a.CreatedAt != a.UpdatedAt {
		t.Error("expected CreatedAt == UpdatedAt for new account")
	}
}

func TestNewAccountGeneratesUniqueIDs(t *testing.T) {
	a1 := NewAccount("p1", "B1", "C1", 0, 0, 0)
	a2 := NewAccount("p1", "B1", "C1", 0, 0, 0)

	if a1.ID == a2.ID {
		t.Error("expected unique IDs for different accounts")
	}
}
