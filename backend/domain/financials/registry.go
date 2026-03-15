package financials

import (
	"context"

	"github.com/lugassawan/panen/backend/domain/provider"
)

// Registry provides read and control operations over registered financial statement providers.
type Registry interface {
	List() []provider.Info
	SetEnabled(name string, enabled bool) bool
	HealthCheckAll(ctx context.Context)
}
