package presenter

import "testing"

func TestTransactionHandlerUnbound(t *testing.T) {
	h := &TransactionHandler{}

	resp, err := h.ListTransactions("", "", "", "", "", "date", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil, want empty TransactionListResponse")
	}
	if len(resp.Items) != 0 {
		t.Errorf("Items length = %d, want 0", len(resp.Items))
	}
}
