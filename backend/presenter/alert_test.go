package presenter

import "testing"

func TestAlertHandlerNilReceiver(t *testing.T) {
	var h *AlertHandler

	t.Run("GetAlertCount returns zero", func(t *testing.T) {
		count, err := h.GetAlertCount()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Fatalf("got %d, want 0", count)
		}
	})

	t.Run("GetActiveAlerts returns nil", func(t *testing.T) {
		alerts, err := h.GetActiveAlerts()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if alerts != nil {
			t.Fatalf("got %v, want nil", alerts)
		}
	})

	t.Run("GetAlertsByTicker returns nil", func(t *testing.T) {
		alerts, err := h.GetAlertsByTicker("BBCA")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if alerts != nil {
			t.Fatalf("got %v, want nil", alerts)
		}
	})

	t.Run("AcknowledgeAlert returns nil", func(t *testing.T) {
		err := h.AcknowledgeAlert("some-id")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
