package payday

import (
	"testing"
	"time"
)

func TestNewPaydayEvent(t *testing.T) {
	before := time.Now().UTC()
	event := NewPaydayEvent("2026-03", "port-1", 500000)
	after := time.Now().UTC()

	if event.ID == "" {
		t.Error("NewPaydayEvent() ID should not be empty")
	}
	if event.Month != "2026-03" {
		t.Errorf("NewPaydayEvent() Month = %q, want %q", event.Month, "2026-03")
	}
	if event.PortfolioID != "port-1" {
		t.Errorf("NewPaydayEvent() PortfolioID = %q, want %q", event.PortfolioID, "port-1")
	}
	if event.Expected != 500000 {
		t.Errorf("NewPaydayEvent() Expected = %v, want %v", event.Expected, 500000.0)
	}
	if event.Actual != 0 {
		t.Errorf("NewPaydayEvent() Actual = %v, want %v", event.Actual, 0.0)
	}
	if event.Status != StatusScheduled {
		t.Errorf("NewPaydayEvent() Status = %q, want %q", event.Status, StatusScheduled)
	}
	if event.DeferUntil != nil {
		t.Error("NewPaydayEvent() DeferUntil should be nil")
	}
	if event.ConfirmedAt != nil {
		t.Error("NewPaydayEvent() ConfirmedAt should be nil")
	}
	if event.CreatedAt.Before(before) || event.CreatedAt.After(after) {
		t.Errorf("NewPaydayEvent() CreatedAt = %v, want between %v and %v", event.CreatedAt, before, after)
	}
	if event.UpdatedAt.Before(before) || event.UpdatedAt.After(after) {
		t.Errorf("NewPaydayEvent() UpdatedAt = %v, want between %v and %v", event.UpdatedAt, before, after)
	}
}

func TestNewCashFlow(t *testing.T) {
	date := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	before := time.Now().UTC()
	cf := NewCashFlow("port-1", FlowTypeMonthly, 100000, date, "March deposit")
	after := time.Now().UTC()

	if cf.ID == "" {
		t.Error("NewCashFlow() ID should not be empty")
	}
	if cf.PortfolioID != "port-1" {
		t.Errorf("NewCashFlow() PortfolioID = %q, want %q", cf.PortfolioID, "port-1")
	}
	if cf.Type != FlowTypeMonthly {
		t.Errorf("NewCashFlow() Type = %q, want %q", cf.Type, FlowTypeMonthly)
	}
	if cf.Amount != 100000 {
		t.Errorf("NewCashFlow() Amount = %v, want %v", cf.Amount, 100000.0)
	}
	if !cf.Date.Equal(date) {
		t.Errorf("NewCashFlow() Date = %v, want %v", cf.Date, date)
	}
	if cf.Note != "March deposit" {
		t.Errorf("NewCashFlow() Note = %q, want %q", cf.Note, "March deposit")
	}
	if cf.CreatedAt.Before(before) || cf.CreatedAt.After(after) {
		t.Errorf("NewCashFlow() CreatedAt = %v, want between %v and %v", cf.CreatedAt, before, after)
	}
}
