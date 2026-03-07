package transaction

import "context"

// Repository defines read-only queries for the unified transaction history.
type Repository interface {
	List(ctx context.Context, filter Filter) ([]Record, error)
	Summarize(ctx context.Context, filter Filter) (*Summary, error)
}
