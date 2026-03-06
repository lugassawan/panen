package stock

import "context"

// SnapshotRepository stores historical financial snapshots for change detection.
type SnapshotRepository interface {
	Insert(ctx context.Context, data *Data) error
	GetLatest(ctx context.Context, ticker, source string) (*Data, error)
	Cleanup(ctx context.Context, ticker string, keepN int) error
}
