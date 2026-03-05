package trailingstop

import "context"

// PeakRepository defines persistence operations for holding peak prices.
type PeakRepository interface {
	Upsert(ctx context.Context, peak *HoldingPeak) error
	GetByHoldingID(ctx context.Context, holdingID string) (*HoldingPeak, error)
	ListByHoldingIDs(ctx context.Context, holdingIDs []string) ([]*HoldingPeak, error)
}
